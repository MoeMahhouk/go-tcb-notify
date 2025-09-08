package pcs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sirupsen/logrus"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/storage"
)

const ServiceName = "fetch-pcs"

// Fetcher handles fetching TCB information from Intel PCS
type Fetcher struct {
	// Storage interfaces
	tcbStore   storage.TCBStore
	quoteStore storage.QuoteStore
	alertStore storage.AlertStore

	// External dependencies
	client *Client

	// Configuration and logging
	config *config.PCS
	logger *logrus.Entry

	// Legacy database connection (for queries not yet refactored)
	db clickhouse.Conn
}

// NewFetcher creates a new PCS fetcher service with dependency injection
func NewFetcher(
	tcbStore storage.TCBStore,
	quoteStore storage.QuoteStore,
	alertStore storage.AlertStore,
	cfg *config.PCS,
) *Fetcher {
	return &Fetcher{
		tcbStore:   tcbStore,
		quoteStore: quoteStore,
		alertStore: alertStore,
		client:     NewClient(cfg.BaseURL, cfg.APIKey),
		config:     cfg,
		logger:     logrus.WithField("service", ServiceName),
	}
}

// Run starts the PCS fetching service
func (f *Fetcher) Run(ctx context.Context) error {
	f.logger.Info("Starting PCS fetcher service")

	// Initial fetch
	if err := f.FetchAll(ctx); err != nil {
		f.logger.WithError(err).Error("Initial PCS fetch failed")
	}

	ticker := time.NewTicker(f.config.POLL_INTERVAL)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			f.logger.Info("PCS fetcher service stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := f.FetchAll(ctx); err != nil {
				f.logger.WithError(err).Error("Periodic PCS fetch failed")
			}
		}
	}
}

// FetchAll fetches all FMSPCs and their TCB info from Intel PCS
func (f *Fetcher) FetchAll(ctx context.Context) error {
	start := time.Now()

	// Step 1: Fetch all FMSPCs from Intel
	fmspcs, err := f.fetchFMSPCs(ctx)
	if err != nil {
		return fmt.Errorf("fetch FMSPCs: %w", err)
	}

	f.logger.WithField("count", len(fmspcs)).Info("Fetched FMSPCs from Intel PCS")

	// Step 2: Fetch TCB info for each FMSPC
	stats := f.fetchTCBInfoForAll(ctx, fmspcs)

	f.logger.WithFields(logrus.Fields{
		"duration": time.Since(start),
		"total":    stats.Total,
		"updated":  stats.Updated,
		"errors":   stats.Errors,
		"skipped":  stats.Skipped,
	}).Info("Completed PCS fetch cycle")

	return nil
}

// fetchFMSPCs fetches all FMSPCs from Intel PCS API
func (f *Fetcher) fetchFMSPCs(ctx context.Context) ([]models.FMSPCResponse, error) {
	// Fetch from Intel API
	fmspcs, err := f.client.GetFMSPCs(ctx, "all")
	if err != nil {
		return nil, err
	}

	// Store in database
	for _, fmspc := range fmspcs {
		if err := f.storeFMSPC(ctx, fmspc); err != nil {
			f.logger.WithError(err).WithField("fmspc", fmspc.FMSPC).Error("Failed to store FMSPC")
		}
	}

	return fmspcs, nil
}

// FetchStats contains statistics for a fetch operation
type FetchStats struct {
	Total   int
	Updated int
	Skipped int
	Errors  int
}

// fetchTCBInfoForAll fetches TCB info for all FMSPCs
func (f *Fetcher) fetchTCBInfoForAll(ctx context.Context, fmspcs []models.FMSPCResponse) FetchStats {
	stats := FetchStats{Total: len(fmspcs)}

	for _, fmspc := range fmspcs {
		updated, err := f.fetchAndStoreTCBInfo(ctx, fmspc.FMSPC)
		if err != nil {
			f.logger.WithError(err).WithField("fmspc", fmspc.FMSPC).Error("Failed to fetch TCB info")
			stats.Errors++
			continue
		}

		if updated {
			stats.Updated++
		} else {
			stats.Skipped++
		}

		// Small delay to avoid rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	return stats
}

// fetchAndStoreTCBInfo fetches and stores TCB info for a single FMSPC
func (f *Fetcher) fetchAndStoreTCBInfo(ctx context.Context, fmspc string) (bool, error) {
	// Check if update is needed
	currentEvalNum, err := f.getCurrentEvalNumber(ctx, fmspc)
	if err != nil && err != ErrNotFound {
		return false, fmt.Errorf("get current eval number: %w", err)
	}

	// Fetch from Intel
	tcbResp, err := f.client.GetTCBInfo(ctx, fmspc)
	if err != nil {
		if err == ErrNotFound {
			f.logger.WithField("fmspc", fmspc).Debug("No TCB info available")
			return false, nil
		}
		return false, err
	}

	// Parse the TCB info to get evaluation number
	var tcbInfo models.TCBInfo
	if err := json.Unmarshal(tcbResp.TCBInfo, &tcbInfo); err != nil {
		return false, fmt.Errorf("unmarshal TCB info: %w", err)
	}

	// Check if updated (or if this is the first time we're storing this FMSPC)
	updated := tcbInfo.TCBEvaluationDataNumber > currentEvalNum
	isFirstTime := err == ErrNotFound

	// Always store TCB info if it's new or updated
	if updated || isFirstTime {
		if err := f.storeTCBInfo(ctx, fmspc, tcbResp); err != nil {
			return false, fmt.Errorf("store TCB info: %w", err)
		}

		if isFirstTime {
			f.logger.WithFields(logrus.Fields{
				"fmspc":    fmspc,
				"eval_num": tcbInfo.TCBEvaluationDataNumber,
			}).Info("Stored new TCB info for FMSPC")
		} else if updated && currentEvalNum > 0 {
			f.logger.WithFields(logrus.Fields{
				"fmspc":    fmspc,
				"old_eval": currentEvalNum,
				"new_eval": tcbInfo.TCBEvaluationDataNumber,
			}).Info("TCB update detected")

			// Create alert immediately when TCB update is detected
			if err := f.createTCBUpdateAlert(ctx, fmspc, currentEvalNum, tcbInfo.TCBEvaluationDataNumber); err != nil {
				f.logger.WithError(err).WithField("fmspc", fmspc).Error("Failed to create TCB update alert")
			}
		}

		return true, nil
	}

	f.logger.WithFields(logrus.Fields{
		"fmspc":        fmspc,
		"current_eval": currentEvalNum,
		"fetched_eval": tcbInfo.TCBEvaluationDataNumber,
	}).Debug("TCB info already up to date, skipping")

	return false, nil
}

// storeFMSPC stores an FMSPC using the interface
func (f *Fetcher) storeFMSPC(ctx context.Context, fmspc models.FMSPCResponse) error {
	platform := fmspc.Platform
	if platform == "" {
		platform = "ALL"
	}

	return f.tcbStore.StoreFMSPC(ctx, strings.ToUpper(fmspc.FMSPC), platform)
}

// storeTCBInfo stores TCB info using the interface
func (f *Fetcher) storeTCBInfo(ctx context.Context, fmspc string, tcbResp *models.TCBInfoResponse) error {
	// Parse the TCB info from the raw JSON
	var tcbInfo models.TCBInfo
	if err := json.Unmarshal(tcbResp.TCBInfo, &tcbInfo); err != nil {
		return fmt.Errorf("unmarshal TCB info: %w", err)
	}

	rawJSON, err := json.Marshal(tcbResp)
	if err != nil {
		return fmt.Errorf("marshal raw response: %w", err)
	}

	f.logger.WithFields(logrus.Fields{
		"fmspc":       strings.ToUpper(fmspc),
		"eval_num":    tcbInfo.TCBEvaluationDataNumber,
		"tcb_type":    tcbInfo.TCBType,
		"issue_date":  tcbInfo.IssueDate,
		"next_update": tcbInfo.NextUpdate,
	}).Debug("Storing TCB info")

	// Prepare TCB info for storage
	tcbInfo.FMSPC = strings.ToUpper(fmspc)
	tcbInfo.RawJSON = string(rawJSON)

	return f.tcbStore.StoreTCBInfo(ctx, &tcbInfo)
}

// getCurrentEvalNumber gets the current TCB evaluation number for an FMSPC
func (f *Fetcher) getCurrentEvalNumber(ctx context.Context, fmspc string) (uint32, error) {
	evalNum, err := f.tcbStore.GetCurrentEvalNumber(ctx, fmspc)
	if err != nil {
		if strings.Contains(err.Error(), "no TCB info found") {
			return 0, ErrNotFound
		}
		return 0, err
	}
	return evalNum, nil
}

// createTCBUpdateAlert creates an alert when a TCB update is detected
func (f *Fetcher) createTCBUpdateAlert(ctx context.Context, fmspc string, oldEval, newEval uint32) error {
	// Count affected registered quotes using interface
	affectedCount, err := f.quoteStore.CountAffectedQuotes(ctx, fmspc)
	if err != nil {
		f.logger.WithError(err).Warn("Failed to count affected quotes, using 0")
		affectedCount = 0
	}

	// Create alert with details
	details := fmt.Sprintf("TCB evaluation updated from %d to %d for FMSPC %s, affecting %d registered quotes",
		oldEval, newEval, fmspc, affectedCount)

	f.logger.WithFields(logrus.Fields{
		"fmspc":           fmspc,
		"old_eval":        oldEval,
		"new_eval":        newEval,
		"affected_quotes": affectedCount,
	}).Warn("TCB update detected, creating alert")

	// Create alert using interface
	alert := &models.TCBAlert{
		FMSPC:               fmspc,
		OldEvalNumber:       oldEval,
		NewEvalNumber:       newEval,
		AffectedQuotesCount: uint32(affectedCount),
		Details:             details,
		CreatedAt:           time.Now(),
		Acknowledged:        false,
	}

	return f.alertStore.CreateAlert(ctx, alert)
}

var ErrNotFound = fmt.Errorf("not found")
