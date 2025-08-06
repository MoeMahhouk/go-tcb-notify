package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
	"github.com/sirupsen/logrus"
)

type TCBFetcher struct {
	config       *config.Config
	db           *sql.DB
	client       *http.Client
	fmspcService *FMSPCService
}

func NewTCBFetcher(cfg *config.Config, db *sql.DB, fmspcService *FMSPCService) *TCBFetcher {
	return &TCBFetcher{
		config:       cfg,
		db:           db,
		fmspcService: fmspcService,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (f *TCBFetcher) Start(ctx context.Context) {
	logrus.Info("Starting TCB fetcher service")

	// Use the interval from config
	interval := f.config.TCBCheckInterval
	if interval == 0 {
		logrus.Warn("TCB check interval is zero, using default 1h")
		interval = time.Hour
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run initial check
	if err := f.CheckForUpdates(ctx); err != nil {
		logrus.WithError(err).Error("Initial TCB check failed")
	}

	for {
		select {
		case <-ctx.Done():
			logrus.Info("TCB fetcher service stopped")
			return
		case <-ticker.C:
			if err := f.CheckForUpdates(ctx); err != nil {
				logrus.WithError(err).Error("TCB check failed")
			}
		}
	}
}

func (f *TCBFetcher) CheckForUpdates(ctx context.Context) error {
	logrus.Info("Checking for TCB updates...")

	// Get all FMSPCs from the database
	fmspcs, err := f.fmspcService.GetAllFMSPCs()
	if err != nil {
		return fmt.Errorf("failed to get FMSPCs: %w", err)
	}

	if len(fmspcs) == 0 {
		logrus.Info("No FMSPCs found, fetching from Intel PCS API...")
		if err := f.fmspcService.FetchAndStoreAllFMSPCs(ctx); err != nil {
			return fmt.Errorf("failed to fetch FMSPCs: %w", err)
		}

		// Get FMSPCs again after fetching
		fmspcs, err = f.fmspcService.GetAllFMSPCs()
		if err != nil {
			return fmt.Errorf("failed to get FMSPCs after fetching: %w", err)
		}
	}

	logrus.WithField("count", len(fmspcs)).Info("Checking TCB info for FMSPCs")

	for _, fmspc := range fmspcs {
		if err := f.fetchTCBInfo(ctx, fmspc.FMSPC); err != nil {
			logrus.WithError(err).WithField("fmspc", fmspc.FMSPC).Error("Failed to fetch TCB info")
		}
	}

	return nil
}

func (f *TCBFetcher) fetchTCBInfo(ctx context.Context, fmspc string) error {
	url := fmt.Sprintf("%s/tdx/certification/v4/tcb?fmspc=%s", f.config.PCSBaseURL, fmspc)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch TCB info: %w", err)
	}
	defer resp.Body.Close()

	// Handle different status codes appropriately
	switch resp.StatusCode {
	case http.StatusOK:
		// Continue processing
	case http.StatusNotFound:
		// This FMSPC doesn't support TDX - log and skip
		logrus.WithField("fmspc", fmspc).Debug("FMSPC does not support TDX, skipping")
		return nil
	default:
		return fmt.Errorf("unexpected status code %d for FMSPC %s", resp.StatusCode, fmspc)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var tcbResponse struct {
		TCBInfo models.TCBInfo `json:"tcbInfo"`
	}

	if err := json.Unmarshal(body, &tcbResponse); err != nil {
		return fmt.Errorf("failed to parse TCB response: %w", err)
	}

	// Set the FMSPC from the URL parameter since it might not be in the response
	tcbResponse.TCBInfo.FMSPC = fmspc

	// Check if this is a new TCB evaluation data number
	if f.isNewTCBInfo(fmspc, tcbResponse.TCBInfo.TCBEvaluationDataNumber) {
		logrus.WithFields(logrus.Fields{
			"fmspc":                   fmspc,
			"tcbEvaluationDataNumber": tcbResponse.TCBInfo.TCBEvaluationDataNumber,
		}).Info("New TCB info detected")

		if err := f.storeTCBInfo(tcbResponse.TCBInfo, body); err != nil {
			return fmt.Errorf("failed to store TCB info: %w", err)
		}
	}

	return nil
}

func (f *TCBFetcher) isNewTCBInfo(fmspc string, evalDataNumber int) bool {
	var count int
	query := "SELECT COUNT(*) FROM tdx_tcb_info WHERE fmspc = $1 AND tcb_evaluation_data_number = $2"
	err := f.db.QueryRow(query, fmspc, evalDataNumber).Scan(&count)
	if err != nil {
		logrus.WithError(err).Error("Failed to check for existing TCB info")
		return false
	}
	return count == 0
}

func (f *TCBFetcher) storeTCBInfo(tcbInfo models.TCBInfo, rawResponse []byte) error {
	query := `
		INSERT INTO tdx_tcb_info (
			fmspc, version, issue_date, next_update, tcb_type,
			tcb_evaluation_data_number, tcb_levels, raw_response
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := f.db.Exec(query,
		tcbInfo.FMSPC,
		tcbInfo.Version,
		tcbInfo.IssueDate,
		tcbInfo.NextUpdate,
		tcbInfo.TCBType,
		tcbInfo.TCBEvaluationDataNumber,
		tcbInfo.TCBLevels,
		rawResponse,
	)

	if err != nil {
		return fmt.Errorf("failed to insert TCB info: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"fmspc":                   tcbInfo.FMSPC,
		"tcbEvaluationDataNumber": tcbInfo.TCBEvaluationDataNumber,
	}).Info("Stored new TCB info")

	return nil
}
