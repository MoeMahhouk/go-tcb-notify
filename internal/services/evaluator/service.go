package evaluator

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/storage"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/tdx"
)

const ServiceName = "evaluate-quotes"

// Evaluator handles TDX quote evaluation
type Evaluator struct {
	// Storage interfaces
	quoteStore      storage.QuoteStore
	evaluationStore storage.EvaluationStore

	// Business logic
	verifier *tdx.QuoteVerifier

	// Configuration and logging
	config *config.EvaluateQuotes
	logger *logrus.Entry
}

// NewEvaluator creates a new quote evaluator service with dependency injection
func NewEvaluator(
	quoteStore storage.QuoteStore,
	evaluationStore storage.EvaluationStore,
	cfg *config.EvaluateQuotes,
) *Evaluator {
	return &Evaluator{
		quoteStore:      quoteStore,
		evaluationStore: evaluationStore,
		verifier:        tdx.NewQuoteVerifier(cfg.GetCollateral, cfg.CheckRevocations),
		config:          cfg,
		logger:          logrus.WithField("service", ServiceName),
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

// fetchAllQuotes fetches all registered quotes using the interface
func (e *Evaluator) fetchAllQuotes(ctx context.Context) ([]*models.RegistryQuote, error) {
	return e.quoteStore.GetActiveQuotes(ctx)
}

// evaluateQuotes evaluates a batch of quotes
func (e *Evaluator) evaluateQuotes(ctx context.Context, quotes []*models.RegistryQuote) EvaluationStats {
	stats := EvaluationStats{Total: len(quotes)}

	for _, quote := range quotes {
		evaluation := e.evaluateQuote(*quote)

		// Check for status changes
		if e.checkAndRecordStatusChange(ctx, evaluation) {
			stats.Changed++
		}

		// Store evaluation result using interface
		if err := e.evaluationStore.StoreEvaluation(ctx, evaluation); err != nil {
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
	// Use QuoteVerifier's unified interface for verification and evaluation
	evaluation, err := e.verifier.VerifyAndEvaluateQuote(quote.QuoteBytes)
	if err != nil {
		// This shouldn't happen as VerifyAndEvaluateQuote handles errors internally
		e.logger.WithError(err).Error("Unexpected error from QuoteVerifier")
		return &models.QuoteEvaluation{
			ServiceAddress: quote.ServiceAddress,
			QuoteHash:      quote.QuoteHash,
			QuoteLength:    quote.QuoteLength,
			BlockNumber:    quote.BlockNumber,
			LogIndex:       quote.LogIndex,
			BlockTime:      quote.BlockTime,
			EvaluatedAt:    time.Now(),
			Status:         models.StatusInvalid,
			TCBStatus:      models.TCBStatusNotApplicable,
			ErrorMessage:   err.Error(),
		}
	}

	// Add blockchain metadata that QuoteVerifier doesn't know about
	evaluation.ServiceAddress = quote.ServiceAddress
	evaluation.BlockNumber = quote.BlockNumber
	evaluation.LogIndex = quote.LogIndex
	evaluation.BlockTime = quote.BlockTime
	evaluation.EvaluatedAt = time.Now()

	return evaluation
}

// checkAndRecordStatusChange checks if a quote's status has changed
func (e *Evaluator) checkAndRecordStatusChange(ctx context.Context, eval *models.QuoteEvaluation) bool {
	// Get previous evaluation
	prevStatus, prevTCBStatus, err := e.evaluationStore.GetLastEvaluation(ctx, eval.ServiceAddress, eval.QuoteHash)
	if err != nil {
		// No previous evaluation or error
		return false
	}

	// Check if status changed
	if prevStatus == string(eval.Status) && prevTCBStatus == string(eval.TCBStatus) {
		return false
	}

	// Record status change
	if err := e.evaluationStore.StoreEvaluationHistory(ctx,
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
