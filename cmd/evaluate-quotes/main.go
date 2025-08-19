package main

import (
	"context"
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
	"github.com/MoeMahhouk/go-tcb-notify/pkg/tdx"
	"github.com/google/go-tdx-guest/verify"
)

const serviceName = "evaluate-quotes"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logrus.Info("Shutting down evaluate-quotes service...")
		cancel()
	}()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logrus.Fatalf("Failed to load config: %v", err)
	}

	// Setup logging
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	// Connect to ClickHouse
	ch, err := clickdb.Open(ctx, &cfg.ClickHouse)
	if err != nil {
		logrus.Fatalf("Failed to connect to ClickHouse: %v", err)
	}
	defer ch.Close()

	logrus.WithFields(logrus.Fields{
		"service":           serviceName,
		"poll_interval":     cfg.EvaluateQuotes.PollInterval,
		"batch_size":        cfg.EvaluateQuotes.BatchSize,
		"get_collateral":    cfg.EvaluateQuotes.GetCollateral,
		"check_revocations": cfg.EvaluateQuotes.CheckRevocations,
	}).Info("Starting quote evaluation service")

	// Create evaluator
	evaluator := &QuoteEvaluator{
		clickhouse: ch,
		config:     cfg,
		parser:     tdx.NewQuoteParser(),
	}

	// Main evaluation loop
	ticker := time.NewTicker(cfg.EvaluateQuotes.PollInterval)
	defer ticker.Stop()

	// Evaluate immediately on startup
	evaluator.evaluateAllQuotes(ctx)

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Evaluation service stopped")
			return
		case <-ticker.C:
			evaluator.evaluateAllQuotes(ctx)
		}
	}
}

type QuoteEvaluator struct {
	clickhouse clickhouse.Conn
	config     *config.Config
	parser     *tdx.QuoteParser
}

// evaluateAllQuotes fetches and evaluates ALL registered quotes
func (e *QuoteEvaluator) evaluateAllQuotes(ctx context.Context) {
	start := time.Now()

	// Fetch all registered quotes
	rows, err := e.clickhouse.Query(ctx, clickdb.GetAllRegisteredQuotes)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch registered quotes")
		return
	}
	defer rows.Close()

	evaluatedCount := 0
	invalidCount := 0
	validCount := 0

	for rows.Next() {
		var (
			serviceAddress string
			blockNumber    uint64
			blockTime      time.Time
			txHash         string
			logIndex       uint32
			quoteBytes     []byte
			quoteLen       uint32
			quoteHash      string
		)

		if err := rows.Scan(&serviceAddress, &blockNumber, &blockTime, &txHash,
			&logIndex, &quoteBytes, &quoteLen, &quoteHash); err != nil {
			logrus.WithError(err).Error("Failed to scan quote row")
			continue
		}

		// Evaluate this quote
		status, tcbStatus, errorMsg := e.evaluateQuote(ctx, quoteBytes)

		// Check for status change
		e.checkAndRecordStatusChange(ctx, serviceAddress, quoteHash, status, tcbStatus)

		// Store evaluation result
		if err := e.storeEvaluation(ctx, serviceAddress, quoteHash, quoteLen,
			status, tcbStatus, errorMsg, blockNumber, logIndex, blockTime, quoteBytes); err != nil {
			logrus.WithError(err).WithField("address", serviceAddress).Error("Failed to store evaluation")
			continue
		}

		evaluatedCount++
		if status == "Valid" {
			validCount++
		} else {
			invalidCount++
		}
	}

	duration := time.Since(start)
	logrus.WithFields(logrus.Fields{
		"evaluated": evaluatedCount,
		"valid":     validCount,
		"invalid":   invalidCount,
		"duration":  duration,
	}).Info("Completed quote evaluation cycle")
}

// evaluateQuote verifies a single quote and returns its status
func (e *QuoteEvaluator) evaluateQuote(ctx context.Context, quoteBytes []byte) (status string, tcbStatus string, errorMsg string) {
	// Create verification options
	opts := verify.Options{
		GetCollateral:    e.config.EvaluateQuotes.GetCollateral,
		CheckRevocations: e.config.EvaluateQuotes.CheckRevocations,
	}

	// Verify the quote using google-go-tdx-guest library
	err := verify.TdxQuote(quoteBytes, &opts)

	if err != nil {
		// Quote is invalid
		status = "Invalid"
		errorMsg = err.Error()

		// Try to determine TCB status from error message
		// The verify library returns specific error messages for TCB issues
		switch {
		case strings.Contains(errorMsg, "TCB"):
			tcbStatus = "OutOfDate"
		case strings.Contains(errorMsg, "revoked"):
			tcbStatus = "Revoked"
		default:
			tcbStatus = "Invalid"
		}
	} else {
		// Quote is valid
		status = "Valid"
		tcbStatus = "UpToDate"
		errorMsg = ""
	}

	return status, tcbStatus, errorMsg
}

// checkAndRecordStatusChange tracks if a quote's status has changed
func (e *QuoteEvaluator) checkAndRecordStatusChange(ctx context.Context,
	serviceAddress, quoteHash, newStatus, newTCBStatus string) {

	// Get the last evaluation
	row := e.clickhouse.QueryRow(ctx, clickdb.GetLastEvaluation, serviceAddress, quoteHash)

	var prevStatus, prevTCBStatus string
	err := row.Scan(&prevStatus, &prevTCBStatus)

	if err == nil && (prevStatus != newStatus || prevTCBStatus != newTCBStatus) {
		// Status has changed, record it in history
		if err := e.clickhouse.Exec(ctx, clickdb.InsertQuoteEvaluationHistory,
			serviceAddress, quoteHash, prevStatus, newStatus, prevTCBStatus, newTCBStatus); err != nil {
			logrus.WithError(err).Warn("Failed to record status change")
		} else {
			logrus.WithFields(logrus.Fields{
				"address":         serviceAddress,
				"previous_status": prevStatus,
				"new_status":      newStatus,
				"previous_tcb":    prevTCBStatus,
				"new_tcb":         newTCBStatus,
			}).Info("Quote status changed")
		}
	}
}

// storeEvaluation saves the evaluation result to the database
func (e *QuoteEvaluator) storeEvaluation(ctx context.Context,
	serviceAddress, quoteHash string, quoteLen uint32,
	status, tcbStatus, errorMsg string,
	blockNumber uint64, logIndex uint32, blockTime time.Time,
	quoteBytes []byte) error {

	// Parse the quote to extract components (optional, for detailed tracking)
	var fmspc string
	var sgxComponents, tdxComponents string
	var pceSvn uint16
	var mrTd, mrSeam, mrSignerSeam, reportData string

	parsed, err := e.parser.ParseQuote(quoteBytes)
	if err == nil {
		fmspc = parsed.FMSPC
		mrTd = parsed.MrTd
		mrSeam = parsed.MrSeam
		mrSignerSeam = parsed.MrSignerSeam
		reportData = parsed.ReportData
		// Extract TCB components if parsing succeeded
		if (parsed.TCBComponents != models.TCBComponents{}) {
			sgxComponents = string(parsed.TCBComponents.SGXComponents[:])
			tdxComponents = string(parsed.TCBComponents.TDXComponents[:])
			pceSvn = parsed.TCBComponents.PCESVN
		}
	}

	// Store the evaluation
	return e.clickhouse.Exec(ctx, clickdb.InsertQuoteEvaluation,
		serviceAddress, quoteHash, quoteLen,
		status, tcbStatus, errorMsg,
		fmspc, sgxComponents, tdxComponents, pceSvn,
		blockNumber, logIndex, blockTime,
		mrTd, mrSeam, mrSignerSeam, reportData)
}
