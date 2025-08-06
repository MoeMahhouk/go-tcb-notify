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

type FMSPCService struct {
	config *config.Config
	db     *sql.DB
	client *http.Client
}

func NewFMSPCService(cfg *config.Config, db *sql.DB) *FMSPCService {
	return &FMSPCService{
		config: cfg,
		db:     db,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (f *FMSPCService) FetchAndStoreAllFMSPCs(ctx context.Context) error {
	logrus.Info("Fetching all FMSPCs from Intel PCS API...")

	url := fmt.Sprintf("%s/sgx/certification/v4/fmspcs?platform=all", f.config.PCSBaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch FMSPCs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var fmspcs []models.FMSPCResponse
	if err := json.Unmarshal(body, &fmspcs); err != nil {
		return fmt.Errorf("failed to parse FMSPC response: %w", err)
	}

	logrus.WithField("count", len(fmspcs)).Info("Fetched FMSPCs from Intel PCS")

	// Store FMSPCs in database
	for _, fmspc := range fmspcs {
		if err := f.storeFMSPC(fmspc); err != nil {
			logrus.WithError(err).WithField("fmspc", fmspc.FMSPC).Error("Failed to store FMSPC")
		}
	}

	return nil
}

func (f *FMSPCService) storeFMSPC(fmspc models.FMSPCResponse) error {
	query := `
		INSERT INTO fmspcs (fmspc, platform, updated_at) 
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (fmspc) 
		DO UPDATE SET platform = EXCLUDED.platform, updated_at = CURRENT_TIMESTAMP`

	_, err := f.db.Exec(query, fmspc.FMSPC, fmspc.Platform)
	if err != nil {
		return fmt.Errorf("failed to store FMSPC %s: %w", fmspc.FMSPC, err)
	}

	logrus.WithFields(logrus.Fields{
		"fmspc":    fmspc.FMSPC,
		"platform": fmspc.Platform,
	}).Debug("Stored FMSPC")

	return nil
}

func (f *FMSPCService) GetAllFMSPCs() ([]models.FMSPC, error) {
	query := "SELECT fmspc, platform, created_at, updated_at FROM fmspcs ORDER BY fmspc"

	rows, err := f.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query FMSPCs: %w", err)
	}
	defer rows.Close()

	var fmspcs []models.FMSPC
	for rows.Next() {
		var fmspc models.FMSPC
		if err := rows.Scan(&fmspc.FMSPC, &fmspc.Platform, &fmspc.CreatedAt, &fmspc.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan FMSPC: %w", err)
		}
		fmspcs = append(fmspcs, fmspc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating FMSPCs: %w", err)
	}

	return fmspcs, nil
}
