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
	config *config.Config
	db     *sql.DB
	client *http.Client
}

func NewTCBFetcher(cfg *config.Config, db *sql.DB) *TCBFetcher {
	return &TCBFetcher{
		config: cfg,
		db:     db,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (f *TCBFetcher) Start(ctx context.Context) {
	ticker := time.NewTicker(f.config.TCBCheckInterval)
	defer ticker.Stop()

	// Run initial check
	f.checkTCBUpdates(ctx)

	for {
		select {
		case <-ctx.Done():
			logrus.Info("TCB Fetcher stopping...")
			return
		case <-ticker.C:
			f.checkTCBUpdates(ctx)
		}
	}
}

func (f *TCBFetcher) checkTCBUpdates(ctx context.Context) {
	logrus.Info("Checking for TCB updates...")

	// Get list of FMSPCs to monitor (this would typically come from the registry)
	fmspcs := []string{"50806F000000"} // Example FMSPC

	for _, fmspc := range fmspcs {
		if err := f.fetchTCBInfo(ctx, fmspc); err != nil {
			logrus.WithError(err).WithField("fmspc", fmspc).Error("Failed to fetch TCB info")
		}
	}
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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
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

	// Check if this is a new TCB evaluation data number
	if f.isNewTCBInfo(fmspc, tcbResponse.TCBInfo.TCBEvaluationDataNumber) {
		logrus.WithFields(logrus.Fields{
			"fmspc":                     fmspc,
			"tcbEvaluationDataNumber":   tcbResponse.TCBInfo.TCBEvaluationDataNumber,
		}).Info("New TCB info detected")

		if err := f.storeTCBInfo(tcbResponse.TCBInfo, body); err != nil {
			return fmt.Errorf("failed to store TCB info: %w", err)
		}

		// Trigger quote checking
		// This would notify the quote checker service
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