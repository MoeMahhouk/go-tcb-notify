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
	clickdb "github.com/MoeMahhouk/go-tcb-notify/internal/storage/clickhouse"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
)

const ServiceName = "fetch-pcs"

// Fetcher handles fetching TCB information from Intel PCS
type Fetcher struct {
	db     clickhouse.Conn
	client *Client
	config *config.PCS
	logger *logrus.Entry
}

// NewFetcher creates a new PCS fetcher service
func NewFetcher(db clickhouse.Conn, cfg *config.PCS) *Fetcher {
	return &Fetcher{
		db:     db,
		client: NewClient(cfg.BaseURL, cfg.APIKey),
		config: cfg,
		logger: logrus.WithField("service", ServiceName),
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

// storeFMSPC stores an FMSPC in the database
func (f *Fetcher) storeFMSPC(ctx context.Context, fmspc models.FMSPCResponse) error {
	platform := fmspc.Platform
	if platform == "" {
		platform = "ALL"
	}

	return f.db.Exec(ctx, clickdb.UpsertPCSFMSPC,
		strings.ToUpper(fmspc.FMSPC),
		platform,
	)
}

// storeTCBInfo stores TCB info in the database
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

	// Convert TCB levels to JSON string for storage
	tcbLevelsJSON, err := json.Marshal(tcbInfo.TCBLevels)
	if err != nil {
		return fmt.Errorf("marshal TCB levels: %w", err)
	}

	f.logger.WithFields(logrus.Fields{
		"fmspc":       strings.ToUpper(fmspc),
		"eval_num":    tcbInfo.TCBEvaluationDataNumber,
		"tcb_type":    tcbInfo.TCBType,
		"issue_date":  tcbInfo.IssueDate,
		"next_update": tcbInfo.NextUpdate,
	}).Debug("Storing TCB info")

	return f.db.Exec(ctx, clickdb.InsertPCSTCBInfo,
		strings.ToUpper(fmspc),
		tcbInfo.TCBEvaluationDataNumber,
		tcbInfo.IssueDate,
		tcbInfo.NextUpdate,
		tcbInfo.TCBType,
		string(tcbLevelsJSON),
		string(rawJSON),
	)
}

// getCurrentEvalNumber gets the current TCB evaluation number for an FMSPC
func (f *Fetcher) getCurrentEvalNumber(ctx context.Context, fmspc string) (uint32, error) {
	var evalNum uint32
	row := f.db.QueryRow(ctx, clickdb.CheckTCBUpdate, fmspc)
	if err := row.Scan(&evalNum); err != nil {
		// Check if it's a "no rows" error, which means this FMSPC doesn't exist yet
		if strings.Contains(err.Error(), "no rows") || strings.Contains(err.Error(), "EOF") {
			return 0, ErrNotFound
		}
		return 0, err
	}
	return evalNum, nil
}

// createTCBUpdateAlert creates an alert when a TCB update is detected
func (f *Fetcher) createTCBUpdateAlert(ctx context.Context, fmspc string, oldEval, newEval uint32) error {
	// Count affected registered quotes
	affectedCount := f.countAffectedQuotes(ctx, fmspc)

	// Create alert with details
	details := fmt.Sprintf("TCB evaluation updated from %d to %d for FMSPC %s, affecting %d registered quotes",
		oldEval, newEval, fmspc, affectedCount)

	f.logger.WithFields(logrus.Fields{
		"fmspc":           fmspc,
		"old_eval":        oldEval,
		"new_eval":        newEval,
		"affected_quotes": affectedCount,
	}).Warn("TCB update detected, creating alert")

	return f.createAlert(ctx, fmspc, oldEval, newEval, affectedCount, details)
}

// countAffectedQuotes counts quotes affected by an FMSPC change
func (f *Fetcher) countAffectedQuotes(ctx context.Context, fmspc string) uint32 {
	var count uint32
	row := f.db.QueryRow(ctx, clickdb.CountAffectedRegisteredQuotes, fmspc)
	_ = row.Scan(&count) // Ignore error
	return count
}

// createAlert creates a TCB change alert
func (f *Fetcher) createAlert(ctx context.Context, fmspc string, oldEval, newEval uint32,
	affectedCount uint32, details string) error {
	return f.db.Exec(ctx, clickdb.InsertTCBChangeAlert,
		fmspc, oldEval, newEval, affectedCount, details,
	)
}

var ErrNotFound = fmt.Errorf("not found")
