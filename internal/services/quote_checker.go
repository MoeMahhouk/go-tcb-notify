package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/tdx"
	"github.com/sirupsen/logrus"
)

type QuoteChecker struct {
	config          *config.Config
	db              *sql.DB
	registryService *RegistryService
	alertPublisher  *AlertPublisher
	tcbVerifier     *TCBVerifier
	quoteParser     *tdx.QuoteParser
}

// QuoteTCBComponents represents the TCB components extracted from a quote
type QuoteTCBComponents struct {
	SGXTCBComponents [16]uint8 `json:"sgxTcbComponents"`
	TDXTCBComponents [16]uint8 `json:"tdxTcbComponents"`
	PCESVN           int       `json:"pcesvn"`
}

func NewQuoteChecker(cfg *config.Config, db *sql.DB, registryService *RegistryService, alertPublisher *AlertPublisher) *QuoteChecker {
	return &QuoteChecker{
		config:          cfg,
		db:              db,
		registryService: registryService,
		alertPublisher:  alertPublisher,
		tcbVerifier:     NewTCBVerifier(db),
		quoteParser:     tdx.NewQuoteParser(),
	}
}

func (c *QuoteChecker) Start(ctx context.Context) {
	logrus.Info("Starting Quote Checker service...")

	ticker := time.NewTicker(c.config.QuoteCheckInterval)
	defer ticker.Stop()

	// Initial check on startup
	c.performFullCheck(ctx)

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Quote Checker stopping...")
			return
		case <-ticker.C:
			c.performFullCheck(ctx)
		}
	}
}

func (c *QuoteChecker) performFullCheck(ctx context.Context) {
	logrus.Info("Performing full quote check cycle...")

	// Step 1: Fetch latest quotes from registry
	if err := c.registryService.FetchQuotesFromRegistry(ctx); err != nil {
		logrus.WithError(err).Error("Failed to fetch quotes from registry")
	}

	// Step 2: Check all monitored quotes
	if err := c.checkMonitoredQuotes(ctx); err != nil {
		logrus.WithError(err).Error("Failed to check monitored quotes")
	}

	logrus.Info("Completed quote check cycle")
}

func (c *QuoteChecker) checkMonitoredQuotes(ctx context.Context) error {
	quotes, err := c.registryService.GetMonitoredQuotes()
	if err != nil {
		return fmt.Errorf("failed to get monitored quotes: %w", err)
	}

	logrus.WithField("count", len(quotes)).Info("Checking monitored quotes")

	for _, quote := range quotes {
		if err := c.verifyQuote(ctx, quote); err != nil {
			logrus.WithError(err).WithField("address", quote.Address).Error("Failed to verify quote")
		}
	}

	return nil
}

func (c *QuoteChecker) verifyQuote(ctx context.Context, quote *models.MonitoredQuote) error {
	// Parse the TDX quote using go-tdx-guest library
	parsedQuote, err := c.quoteParser.ParseQuote(quote.QuoteData)
	if err != nil {
		return fmt.Errorf("failed to parse TDX quote: %w", err)
	}

	// Extract FMSPC from the parsed quote
	fmspc := parsedQuote.FMSPC
	if fmspc == "" {
		return fmt.Errorf("no FMSPC found in quote")
	}

	// Update the quote's FMSPC if it's different
	if quote.FMSPC != fmspc {
		quote.FMSPC = fmspc
		if err := c.updateQuoteFMSPC(quote.Address, fmspc); err != nil {
			logrus.WithError(err).Error("Failed to update quote FMSPC")
		}
	}

	// Get current TCB info for the quote's FMSPC
	tcbInfo, err := c.getCurrentTCBInfo(fmspc)
	if err != nil {
		if err == sql.ErrNoRows {
			logrus.WithField("fmspc", fmspc).Warn("No TCB info found for FMSPC, skipping verification")
			return nil
		}
		return fmt.Errorf("failed to get TCB info for FMSPC %s: %w", fmspc, err)
	}

	// Extract TCB components from parsed quote
	quoteTCB := &QuoteTCBComponents{
		SGXTCBComponents: parsedQuote.TCBComponents.SGXComponents,
		TDXTCBComponents: parsedQuote.TCBComponents.TDXComponents,
		PCESVN:           int(parsedQuote.TCBComponents.PCESVN),
	}

	// Store the extracted TCB components
	tcbComponentsJSON, err := json.Marshal(quoteTCB)
	if err != nil {
		return fmt.Errorf("failed to marshal TCB components: %w", err)
	}
	quote.TCBComponents = tcbComponentsJSON

	// Verify quote against current TCB levels
	result, err := c.tcbVerifier.VerifyQuote(quoteTCB, tcbInfo)
	if err != nil {
		return fmt.Errorf("failed to verify quote: %w", err)
	}

	// Update quote status and trigger alerts if needed
	return c.handleVerificationResult(quote, result, tcbInfo)
}

func (c *QuoteChecker) handleVerificationResult(quote *models.MonitoredQuote, result *TCBVerificationResult, tcbInfo *models.TCBInfo) error {
	previousStatus := quote.CurrentStatus
	newStatus := result.Status
	needsUpdate := result.Status != "UpToDate"

	// Update quote in database
	if err := c.updateQuoteStatus(quote.Address, newStatus, needsUpdate); err != nil {
		return fmt.Errorf("failed to update quote status: %w", err)
	}

	// Trigger alert if status degraded
	if c.shouldTriggerAlert(previousStatus, newStatus, quote.Address) {
		alert := &models.Alert{
			Severity:  c.getAlertSeverity(newStatus),
			Source:    "go-tcb-notify",
			Timestamp: time.Now(),
			Quote: models.QuoteAlert{
				Address:                 quote.Address,
				Reason:                  result.Reason,
				PreviousStatus:          previousStatus,
				NewStatus:               newStatus,
				WorkloadID:              quote.WorkloadID,
				TCBEvaluationDataNumber: tcbInfo.TCBEvaluationDataNumber,
				AdvisoryIDs:             result.AdvisoryIDs,
				FMSPC:                   quote.FMSPC,
				SuggestedAction:         c.getSuggestedAction(newStatus),
			},
		}

		if err := c.alertPublisher.SendAlert(alert); err != nil {
			logrus.WithError(err).Error("Failed to send alert")
		}
	}

	return nil
}

func (c *QuoteChecker) getCurrentTCBInfo(fmspc string) (*models.TCBInfo, error) {
	query := `
		SELECT fmspc, version, issue_date, next_update, tcb_type, 
		       tcb_evaluation_data_number, tcb_levels, raw_response, created_at
		FROM tdx_tcb_info 
		WHERE fmspc = $1 
		ORDER BY tcb_evaluation_data_number DESC 
		LIMIT 1`

	tcbInfo := &models.TCBInfo{}
	err := c.db.QueryRow(query, fmspc).Scan(
		&tcbInfo.FMSPC,
		&tcbInfo.Version,
		&tcbInfo.IssueDate,
		&tcbInfo.NextUpdate,
		&tcbInfo.TCBType,
		&tcbInfo.TCBEvaluationDataNumber,
		&tcbInfo.TCBLevels,
		&tcbInfo.RawResponse,
		&tcbInfo.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return tcbInfo, nil
}

func (c *QuoteChecker) updateQuoteFMSPC(address, fmspc string) error {
	_, err := c.db.Exec(`
		UPDATE monitored_tdx_quotes 
		SET fmspc = $1
		WHERE address = $2`,
		fmspc, address)
	return err
}

func (c *QuoteChecker) updateQuoteStatus(address, status string, needsUpdate bool) error {
	_, err := c.db.Exec(`
		UPDATE monitored_tdx_quotes 
		SET current_status = $1, needs_update = $2, last_checked = CURRENT_TIMESTAMP
		WHERE address = $3`,
		status, needsUpdate, address)
	return err
}

func (c *QuoteChecker) shouldTriggerAlert(previousStatus, newStatus, address string) bool {
	// Don't alert if status improved or stayed the same
	if previousStatus == newStatus || newStatus == "UpToDate" {
		return false
	}

	// Check alert cooldown
	var lastAlert time.Time
	err := c.db.QueryRow(`
		SELECT COALESCE(MAX(sent_at), '1970-01-01') 
		FROM alert_history 
		WHERE quote_address = $1`,
		address).Scan(&lastAlert)

	if err == nil && time.Since(lastAlert) < c.config.AlertCooldown {
		logrus.WithFields(logrus.Fields{
			"address":   address,
			"lastAlert": lastAlert,
			"cooldown":  c.config.AlertCooldown,
		}).Debug("Alert still in cooldown period")
		return false
	}

	return true
}

func (c *QuoteChecker) getAlertSeverity(status string) string {
	switch status {
	case "OutOfDate":
		return "warning"
	case "Revoked":
		return "critical"
	case "ConfigurationNeeded":
		return "warning"
	default:
		return "info"
	}
}

func (c *QuoteChecker) getSuggestedAction(status string) string {
	switch status {
	case "OutOfDate":
		return "Update TDX platform and re-attest"
	case "Revoked":
		return "Platform compromised - immediate re-attestation required"
	case "ConfigurationNeeded":
		return "Update platform configuration and re-attest"
	default:
		return "Monitor for updates"
	}
}
