package storage

import (
	"context"
	"time"

	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
)

// QuoteStore handles persistence of TDX quotes
type QuoteStore interface {
	// StoreRawQuote stores a raw quote from registry events
	StoreRawQuote(ctx context.Context, quote *models.RegistryQuote) error

	// GetActiveQuotes returns all quotes that haven't been invalidated
	GetActiveQuotes(ctx context.Context) ([]*models.RegistryQuote, error)

	// StoreInvalidation records a quote invalidation event
	StoreInvalidation(ctx context.Context, address string, blockNumber uint64, blockTime time.Time, txHash string, logIndex uint32) error

	// CountAffectedQuotes counts quotes affected by an FMSPC change
	CountAffectedQuotes(ctx context.Context, fmspc string) (int64, error)
}

// TCBStore handles TCB information persistence
type TCBStore interface {
	// StoreTCBInfo stores TCB information from Intel PCS
	StoreTCBInfo(ctx context.Context, info *models.TCBInfo) error

	// GetLatestTCBInfo returns the most recent TCB info for an FMSPC
	GetLatestTCBInfo(ctx context.Context, fmspc string) (*models.TCBInfo, error)

	// GetAllFMSPCs returns all known FMSPCs
	GetAllFMSPCs(ctx context.Context) ([]string, error)

	// UpdateFMSPCSeen updates the last seen time for an FMSPC
	UpdateFMSPCSeen(ctx context.Context, fmspc string) error

	// StoreFMSPC stores FMSPC information
	StoreFMSPC(ctx context.Context, fmspc string, platform string) error

	// GetCurrentEvalNumber gets the current TCB evaluation number for an FMSPC
	GetCurrentEvalNumber(ctx context.Context, fmspc string) (uint32, error)
}

// EvaluationStore handles quote evaluation results
type EvaluationStore interface {
	// StoreEvaluation stores a quote evaluation result
	StoreEvaluation(ctx context.Context, eval *models.QuoteEvaluation) error

	// GetLastEvaluation gets the last evaluation for a quote (for status change detection)
	GetLastEvaluation(ctx context.Context, serviceAddress, quoteHash string) (status, tcbStatus string, err error)

	// StoreEvaluationHistory stores evaluation history for tracking changes
	StoreEvaluationHistory(ctx context.Context, serviceAddress, quoteHash, prevStatus, newStatus, prevTCBStatus, newTCBStatus string) error
}

// OffsetStore handles service checkpoint/offset persistence
type OffsetStore interface {
	// LoadOffset loads the last processed position for a service
	LoadOffset(ctx context.Context, service string) (blockNumber uint64, logIndex uint32, err error)

	// SaveOffset saves the current processing position for a service
	SaveOffset(ctx context.Context, service string, blockNumber uint64, logIndex uint32) error
}

// AlertStore handles TCB change alerts
type AlertStore interface {
	// CreateAlert creates a new TCB change alert
	CreateAlert(ctx context.Context, alert *models.TCBAlert) error

	// GetPendingAlerts returns alerts that haven't been acknowledged
	GetPendingAlerts(ctx context.Context) ([]*models.TCBAlert, error)

	// GetRecentAlerts returns recent alerts (last N alerts)
	GetRecentAlerts(ctx context.Context, limit int) ([]*models.TCBAlert, error)
}
