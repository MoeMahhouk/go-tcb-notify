package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sirupsen/logrus"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	clickdb "github.com/MoeMahhouk/go-tcb-notify/internal/storage/clickhouse"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/pcs"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/tdx"
)

const serviceName = "fetch-pcs"

type PCSFetcher struct {
	clickhouse clickhouse.Conn
	httpClient *http.Client
	config     *config.Config
	baseURL    string
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logrus.Info("Shutting down...")
		cancel()
	}()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// Setup logging
	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	fetcher, err := NewPCSFetcher(ctx, cfg)
	if err != nil {
		log.Fatalf("create fetcher: %v", err)
	}
	defer fetcher.Close()

	if err := fetcher.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("fetcher failed: %v", err)
	}
}

func NewPCSFetcher(ctx context.Context, cfg *config.Config) (*PCSFetcher, error) {
	ch, err := clickdb.Open(ctx, &cfg.ClickHouse)
	if err != nil {
		return nil, fmt.Errorf("open clickhouse: %w", err)
	}

	baseURL := cfg.PCS.BaseURL
	if baseURL == "" {
		baseURL = "https://api.trustedservices.intel.com"
	}

	return &PCSFetcher{
		clickhouse: ch,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		config:  cfg,
		baseURL: baseURL,
	}, nil
}

func (f *PCSFetcher) Close() error {
	if f.clickhouse != nil {
		return f.clickhouse.Close()
	}
	return nil
}

func (f *PCSFetcher) Run(ctx context.Context) error {
	logrus.WithField("service", serviceName).Info("Starting PCS fetcher")

	// Initial fetch
	if err := f.fetchAllIntelPCSData(ctx); err != nil {
		logrus.WithError(err).Error("Initial Intel PCS fetch failed")
	}

	// Set up periodic fetch
	ticker := time.NewTicker(f.config.PCS.POLL_INTERVAL)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := f.fetchAllIntelPCSData(ctx); err != nil {
				logrus.WithError(err).Error("Periodic Intel PCS fetch failed")
			}
		}
	}
}

// fetchAllIntelPCSData fetches ALL data from Intel PCS (global tracking)
// This data is then used by evaluate-quotes service
func (f *PCSFetcher) fetchAllIntelPCSData(ctx context.Context) error {
	logrus.WithField("service", serviceName).Info("Starting Intel PCS global data fetch")

	// Step 1: Fetch ALL FMSPCs from Intel PCS
	allFMSPCs, err := f.fetchAllFMSPCsFromIntel(ctx)
	if err != nil {
		return fmt.Errorf("fetch Intel FMSPCs: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"service": serviceName,
		"count":   len(allFMSPCs),
	}).Info("Fetched Intel PCS FMSPCs")

	// Step 2: Fetch TCB info for each FMSPC
	successCount := 0
	errorCount := 0
	skipCount := 0

	for _, fmspc := range allFMSPCs {
		// Check if we need to update (based on NextUpdate field)
		needsUpdate, err := f.checkIfTCBUpdateNeeded(ctx, fmspc)
		if err != nil {
			logrus.WithError(err).WithField("fmspc", fmspc).Debug("Error checking TCB update status")
			needsUpdate = true // Fetch if we can't determine
		}

		if !needsUpdate {
			skipCount++
			continue
		}

		if err := f.fetchAndStoreTCBInfo(ctx, fmspc); err != nil {
			logrus.WithError(err).WithField("fmspc", fmspc).Error("Failed to fetch TCB info")
			errorCount++
			continue
		}
		successCount++

		// Small delay to avoid rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	logrus.WithFields(logrus.Fields{
		"service": serviceName,
		"total":   len(allFMSPCs),
		"success": successCount,
		"errors":  errorCount,
		"skipped": skipCount,
	}).Info("Intel PCS TCB fetch complete")

	// Step 3: Detect TCB changes (comparing versions)
	if err := f.detectTCBChanges(ctx); err != nil {
		logrus.WithError(err).Error("Failed to detect TCB changes")
	}

	// Step 4: Create global alerts for significant changes
	if err := f.createGlobalTCBAlerts(ctx); err != nil {
		logrus.WithError(err).Error("Failed to create global TCB alerts")
	}

	return nil
}

// fetchAllFMSPCsFromIntel fetches ALL FMSPCs from Intel PCS API
func (f *PCSFetcher) fetchAllFMSPCsFromIntel(ctx context.Context) ([]string, error) {
	// Correct URL: sgx/certification/v4/fmspcs?platform=all
	url := fmt.Sprintf("%s/sgx/certification/v4/fmspcs?platform=all", f.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Add API key if configured
	if f.config.PCS.APIKey != "" {
		req.Header.Set("Ocp-Apim-Subscription-Key", f.config.PCS.APIKey)
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch FMSPCs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("PCS API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Parse response - Intel returns array of FMSPC objects
	var fmspcList []models.FMSPCResponse
	if err := json.Unmarshal(body, &fmspcList); err != nil {
		return nil, fmt.Errorf("parse FMSPC list: %w", err)
	}

	// Store all FMSPCs in the pcs_fmspcs table
	var fmspcs []string
	for _, item := range fmspcList {
		fmspc := strings.ToUpper(item.FMSPC)
		platform := item.Platform
		if platform == "" {
			platform = "ALL"
		}

		// Store in database
		if err := f.clickhouse.Exec(ctx, clickdb.UpsertPCSFMSPC, fmspc, platform); err != nil {
			logrus.WithError(err).WithField("fmspc", fmspc).Error("Failed to store FMSPC")
			continue
		}

		fmspcs = append(fmspcs, fmspc)
	}

	return fmspcs, nil
}

// checkIfTCBUpdateNeeded checks if TCB info needs updating based on NextUpdate field
func (f *PCSFetcher) checkIfTCBUpdateNeeded(ctx context.Context, fmspc string) (bool, error) {
	row := f.clickhouse.QueryRow(ctx, clickdb.GetLatestTCBInfo, fmspc)

	var (
		storedFMSPC string
		evalNum     uint32
		tcbLevels   string
		rawJSON     string
	)

	if err := row.Scan(&storedFMSPC, &evalNum, &tcbLevels, &rawJSON); err != nil {
		// No existing TCB info, needs fetch
		return true, nil
	}

	// Parse the stored TCB info to check NextUpdate
	var tcbData pcs.TCBInfoData
	if err := json.Unmarshal([]byte(rawJSON), &struct {
		TcbInfo *pcs.TCBInfoData `json:"tcbInfo"`
	}{&tcbData}); err != nil {
		// Can't parse, fetch new
		return true, nil
	}

	// If NextUpdate has passed, we need to fetch
	return time.Now().After(tcbData.NextUpdate), nil
}

func (f *PCSFetcher) fetchAndStoreTCBInfo(ctx context.Context, fmspc string) error {
	url := fmt.Sprintf("%s/tdx/certification/v4/tcb?fmspc=%s", f.baseURL, fmspc)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// Add API key if configured
	if f.config.PCS.APIKey != "" {
		req.Header.Set("Ocp-Apim-Subscription-Key", f.config.PCS.APIKey)
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetch TCB info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// Some FMSPCs might not have TCB info yet
		logrus.WithField("fmspc", fmspc).Debug("No TCB info available for FMSPC")
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("PCS API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// Parse TCB info using the extracted type
	var tcbResp pcs.TCBInfoResponse
	if err := json.Unmarshal(body, &tcbResp); err != nil {
		return fmt.Errorf("parse TCB info: %w", err)
	}

	tcbInfo := tcbResp.TcbInfo
	tcbInfo.FMSPC = strings.ToUpper(fmspc)
	tcbInfo.RawJSON = string(body)

	// Convert TCB levels to JSON for storage
	tcbLevelsJSON, err := json.Marshal(tcbInfo.TcbLevels)
	if err != nil {
		return fmt.Errorf("marshal TCB levels: %w", err)
	}

	// Store using the query from queries.go
	err = f.clickhouse.Exec(ctx, clickdb.UpsertPCSTCBInfo,
		tcbInfo.FMSPC,
		tcbInfo.TCBEvaluationDataNumber,
		tcbInfo.IssueDate,
		tcbInfo.NextUpdate,
		tcbInfo.TCBType,
		string(tcbLevelsJSON),
		tcbInfo.RawJSON,
	)

	if err != nil {
		return fmt.Errorf("store TCB info: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"service":              serviceName,
		"fmspc":                fmspc,
		"evaluationDataNumber": tcbInfo.TCBEvaluationDataNumber,
		"tcbLevels":            len(tcbInfo.TcbLevels),
		"nextUpdate":           tcbInfo.NextUpdate.Format(time.RFC3339),
	}).Debug("Stored TCB info")

	return nil
}

func (f *PCSFetcher) detectTCBChanges(ctx context.Context) error {
	rows, err := f.clickhouse.Query(ctx, clickdb.GetTCBChanges)
	if err != nil {
		return fmt.Errorf("query TCB changes: %w", err)
	}
	defer rows.Close()

	totalChanges := 0
	for rows.Next() {
		var fmspc string
		var currentEval, previousEval uint32
		var currentLevelsJSON, previousLevelsJSON string

		if err := rows.Scan(&fmspc, &currentEval, &previousEval, &currentLevelsJSON, &previousLevelsJSON); err != nil {
			continue
		}

		// Analyze what changed
		changes := f.analyzeTCBChanges(fmspc, currentLevelsJSON, previousLevelsJSON, currentEval)
		for _, change := range changes {
			if err := f.recordTCBChange(ctx, &change); err != nil {
				logrus.WithError(err).Error("Failed to record TCB change")
			}
		}

		if len(changes) > 0 {
			logrus.WithFields(logrus.Fields{
				"service":      serviceName,
				"fmspc":        fmspc,
				"previousEval": previousEval,
				"currentEval":  currentEval,
				"changes":      len(changes),
			}).Info("Detected Intel TCB changes")

			totalChanges += len(changes)
		}
	}

	if totalChanges > 0 {
		logrus.WithFields(logrus.Fields{
			"service":      serviceName,
			"totalChanges": totalChanges,
		}).Info("Total Intel TCB component changes detected")
	}

	return rows.Err()
}

func (f *PCSFetcher) analyzeTCBChanges(fmspc, currentJSON, previousJSON string, evalNum uint32) []models.TCBChange {
	var changes []models.TCBChange

	var currentLevels, previousLevels []models.TCBLevel
	json.Unmarshal([]byte(currentJSON), &currentLevels)
	json.Unmarshal([]byte(previousJSON), &previousLevels)

	if len(currentLevels) == 0 || len(previousLevels) == 0 {
		return changes
	}

	// Compare the first (highest) TCB level components
	current := currentLevels[0]
	previous := previousLevels[0]

	// Check SGX components
	for i, curr := range current.TCB.SGXComponents {
		if i < len(previous.TCB.SGXComponents) {
			prev := previous.TCB.SGXComponents[i]
			if curr.SVN != prev.SVN {
				changeType := "UPGRADE"
				if curr.SVN < prev.SVN {
					changeType = "DOWNGRADE"
				}
				changes = append(changes, models.TCBChange{
					FMSPC:            fmspc,
					ComponentIndex:   i,
					ComponentName:    tdx.GetComponentName(i),
					ComponentType:    tdx.GetComponentType(i),
					OldVersion:       prev.SVN,
					NewVersion:       curr.SVN,
					ChangeType:       changeType,
					EvaluationNumber: evalNum,
					DetectedAt:       time.Now().UTC(),
				})
			}
		}
	}

	// Check TDX components
	for i, curr := range current.TCB.TDXComponents {
		if i < len(previous.TCB.TDXComponents) {
			prev := previous.TCB.TDXComponents[i]
			if curr.SVN != prev.SVN {
				changeType := "UPGRADE"
				if curr.SVN < prev.SVN {
					changeType = "DOWNGRADE"
				}

				componentIndex := 16 + i // TDX components start at index 16
				changes = append(changes, models.TCBChange{
					FMSPC:            fmspc,
					ComponentIndex:   componentIndex,
					ComponentName:    tdx.GetComponentName(componentIndex),
					ComponentType:    tdx.GetComponentType(componentIndex),
					OldVersion:       prev.SVN,
					NewVersion:       curr.SVN,
					ChangeType:       changeType,
					EvaluationNumber: evalNum,
					DetectedAt:       time.Now().UTC(),
				})
			}
		}
	}

	return changes
}

func (f *PCSFetcher) recordTCBChange(ctx context.Context, change *models.TCBChange) error {
	return f.clickhouse.Exec(ctx, clickdb.InsertTCBComponentChange,
		change.FMSPC,
		change.ComponentIndex,
		change.ComponentName,
		change.ComponentType,
		change.OldVersion,
		change.NewVersion,
		change.ChangeType,
		change.EvaluationNumber,
	)
}

func (f *PCSFetcher) createGlobalTCBAlerts(ctx context.Context) error {
	// Use query from queries.go
	rows, err := f.clickhouse.Query(ctx, clickdb.GetRecentTCBChanges)
	if err != nil {
		return fmt.Errorf("query recent TCB changes: %w", err)
	}
	defer rows.Close()

	alertCount := 0
	for rows.Next() {
		var alert pcs.TCBChangeAlert

		if err := rows.Scan(&alert.FMSPC, &alert.EvaluationNumber,
			&alert.TotalChanges, &alert.Downgrades); err != nil {
			continue
		}

		// Determine severity based on changes
		alert.Severity = f.calculateAlertSeverity(alert.TotalChanges, alert.Downgrades)

		// Create alert details
		details := map[string]interface{}{
			"fmspc":            alert.FMSPC,
			"evaluationNumber": alert.EvaluationNumber,
			"totalChanges":     alert.TotalChanges,
			"downgrades":       alert.Downgrades,
			"source":           "Intel PCS",
			"service":          serviceName,
			"timestamp":        time.Now().UTC(),
		}
		detailsJSON, _ := json.Marshal(details)

		alertID := uint64(time.Now().UnixNano())
		if err := f.clickhouse.Exec(ctx, clickdb.InsertGlobalTCBAlert,
			alertID, alert.Severity, alert.FMSPC, alert.EvaluationNumber, string(detailsJSON)); err != nil {
			logrus.WithError(err).Error("Failed to create global TCB alert")
			continue
		}

		alertCount++
	}

	if alertCount > 0 {
		logrus.WithFields(logrus.Fields{
			"service": serviceName,
			"count":   alertCount,
		}).Info("Created global TCB update alerts")
	}

	return rows.Err()
}

func (f *PCSFetcher) calculateAlertSeverity(changeCount, downgradeCount uint32) string {
	if downgradeCount > 3 {
		return "CRITICAL"
	}
	if changeCount > 10 || downgradeCount > 0 {
		return "HIGH"
	}
	if changeCount > 5 {
		return "MEDIUM"
	}
	return "LOW"
}
