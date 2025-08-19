package evaluator

import (
	"context"
	"encoding/hex"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/go-tdx-guest/verify"
	"github.com/sirupsen/logrus"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	clickdb "github.com/MoeMahhouk/go-tcb-notify/internal/storage/clickhouse"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/tdx"
)

const ServiceName = "evaluate-quotes"

// Evaluator handles TDX quote evaluation
type Evaluator struct {
	db         clickhouse.Conn
	parser     *tdx.QuoteParser
	config     *config.EvaluateQuotes
	logger     *logrus.Entry
	verifyOpts *verify.Options
}

// NewEvaluator creates a new quote evaluator service
func NewEvaluator(db clickhouse.Conn, cfg *config.EvaluateQuotes) *Evaluator {
	return &Evaluator{
		db:     db,
		parser: tdx.NewQuoteParser(),
		config: cfg,
		logger: logrus.WithField("service", ServiceName),
		verifyOpts: &verify.Options{
			GetCollateral:    cfg.GetCollateral,
			CheckRevocations: cfg.CheckRevocations,
		},
	}
}

// Run starts the quote evaluation service
func (e *Evaluator) Run(ctx context.Context) error {
	e.logger.Info("Starting quote evaluator service")

	// Evaluate immediately on startup
	e.evaluateAll(ctx)

	ticker := time.NewTicker(e.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			e.logger.Info("Quote evaluator service stopped")
			return ctx.Err()
		case <-ticker.C:
			e.evaluateAll(ctx)
		}
	}
}

// evaluateAll evaluates all registered quotes
func (e *Evaluator) evaluateAll(ctx context.Context) {
	start := time.Now()

	quotes, err := e.fetchAllQuotes(ctx)
	if err != nil {
		e.logger.WithError(err).Error("Failed to fetch quotes")
		return
	}

	stats := e.evaluateQuotes(ctx, quotes)

	e.logger.WithFields(logrus.Fields{
		"duration": time.Since(start),
		"total":    stats.Total,
		"valid":    stats.Valid,
		"invalid":  stats.Invalid,
		"changed":  stats.Changed,
	}).Info("Completed quote evaluation cycle")
}

// EvaluationStats contains statistics for an evaluation cycle
type EvaluationStats struct {
	Total   int
	Valid   int
	Invalid int
	Changed int
}

// fetchAllQuotes fetches all registered quotes from the database
func (e *Evaluator) fetchAllQuotes(ctx context.Context) ([]models.RegistryQuote, error) {
	rows, err := e.db.Query(ctx, clickdb.GetAllRegisteredQuotes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quotes []models.RegistryQuote
	for rows.Next() {
		var quote models.RegistryQuote
		if err := rows.Scan(
			&quote.ServiceAddress,
			&quote.BlockNumber,
			&quote.BlockTime,
			&quote.TxHash,
			&quote.LogIndex,
			&quote.QuoteBytes,
			&quote.QuoteLength,
			&quote.QuoteHash,
		); err != nil {
			e.logger.WithError(err).Error("Failed to scan quote")
			continue
		}
		quotes = append(quotes, quote)
	}

	return quotes, rows.Err()
}

// evaluateQuotes evaluates a batch of quotes
func (e *Evaluator) evaluateQuotes(ctx context.Context, quotes []models.RegistryQuote) EvaluationStats {
	stats := EvaluationStats{Total: len(quotes)}

	for _, quote := range quotes {
		evaluation := e.evaluateQuote(quote)

		// Check for status changes
		if e.checkAndRecordStatusChange(ctx, evaluation) {
			stats.Changed++
		}

		// Store evaluation result
		if err := e.storeEvaluation(ctx, evaluation); err != nil {
			e.logger.WithError(err).WithField("address", quote.ServiceAddress).Error("Failed to store evaluation")
			continue
		}

		if evaluation.Status == models.StatusValid {
			stats.Valid++
		} else {
			stats.Invalid++
		}
	}

	return stats
}

// evaluateQuote evaluates a single quote
func (e *Evaluator) evaluateQuote(quote models.RegistryQuote) *models.QuoteEvaluation {
	evaluation := &models.QuoteEvaluation{
		ServiceAddress: quote.ServiceAddress,
		QuoteHash:      quote.QuoteHash,
		QuoteLength:    quote.QuoteLength,
		BlockNumber:    quote.BlockNumber,
		LogIndex:       quote.LogIndex,
		BlockTime:      quote.BlockTime,
		EvaluatedAt:    time.Now(),
	}

	// Parse the quote
	parsed, err := e.parser.ParseQuote(quote.QuoteBytes)
	if err != nil {
		evaluation.Status = models.StatusInvalidFormat
		evaluation.TCBStatus = models.TCBStatusNotApplicable
		evaluation.ErrorMessage = err.Error()
		return evaluation
	}

	// Extract components
	evaluation.FMSPC = parsed.FMSPC
	if parsed.TCBComponents != (models.TCBComponents{}) {
		evaluation.TCBComponents = parsed.TCBComponents
	}

	// Verify the quote
	err = verify.TdxQuote(quote.QuoteBytes, e.verifyOpts)
	if err != nil {
		evaluation.Status = models.StatusInvalid
		evaluation.ErrorMessage = err.Error()

		// Determine TCB status from error
		evaluation.TCBStatus = e.determineTCBStatus(err.Error())
	} else {
		evaluation.Status = models.StatusValid
		evaluation.TCBStatus = models.TCBStatusUpToDate
	}

	return evaluation
}

// determineTCBStatus determines TCB status from verification error
func (e *Evaluator) determineTCBStatus(errMsg string) models.TCBStatus {
	// Simple heuristic based on error message
	switch {
	case strings.Contains(errMsg, "TCB"):
		return models.TCBStatusOutOfDate
	case strings.Contains(errMsg, "revoked"):
		return models.TCBStatusRevoked
	case strings.Contains(errMsg, "configuration"):
		return models.TCBStatusConfigurationNeeded
	default:
		return models.TCBStatusUnknown
	}
}

// checkAndRecordStatusChange checks if a quote's status has changed
func (e *Evaluator) checkAndRecordStatusChange(ctx context.Context, eval *models.QuoteEvaluation) bool {
	// Get previous evaluation
	row := e.db.QueryRow(ctx, clickdb.GetLastEvaluation, eval.ServiceAddress, eval.QuoteHash)

	var prevStatus string
	var prevTCBStatus string
	if err := row.Scan(&prevStatus, &prevTCBStatus); err != nil {
		// No previous evaluation or error
		return false
	}

	// Check if status changed
	if prevStatus == string(eval.Status) && prevTCBStatus == string(eval.TCBStatus) {
		return false
	}

	// Record status change
	if err := e.db.Exec(ctx, clickdb.InsertQuoteEvaluationHistory,
		eval.ServiceAddress,
		eval.QuoteHash,
		prevStatus,
		string(eval.Status),
		prevTCBStatus,
		string(eval.TCBStatus),
	); err != nil {
		e.logger.WithError(err).Warn("Failed to record status change")
	} else {
		e.logger.WithFields(logrus.Fields{
			"address":     eval.ServiceAddress,
			"prev_status": prevStatus,
			"new_status":  eval.Status,
			"prev_tcb":    prevTCBStatus,
			"new_tcb":     eval.TCBStatus,
		}).Info("Quote status changed")
	}

	return true
}

// storeEvaluation stores an evaluation result
func (e *Evaluator) storeEvaluation(ctx context.Context, eval *models.QuoteEvaluation) error {
	// Convert components to storage format
	sgxComponents := ""
	tdxComponents := ""

	if eval.TCBComponents != (models.TCBComponents{}) {
		sgxComponents = hex.EncodeToString(eval.TCBComponents.SGXComponents[:])
		tdxComponents = hex.EncodeToString(eval.TCBComponents.TDXComponents[:])
	}

	return e.db.Exec(ctx, clickdb.InsertQuoteEvaluation,
		eval.ServiceAddress,
		eval.QuoteHash,
		eval.QuoteLength,
		string(eval.Status),
		string(eval.TCBStatus),
		eval.ErrorMessage,
		eval.FMSPC,
		sgxComponents,
		tdxComponents,
		eval.TCBComponents.PCESVN,
		eval.BlockNumber,
		eval.LogIndex,
		eval.BlockTime,
		eval.MrTd,
		eval.MrSeam,
		eval.MrSignerSeam,
		eval.ReportData,
	)
}
