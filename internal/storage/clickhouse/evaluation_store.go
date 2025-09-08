package clickhouse

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/storage"
)

// EvaluationStore implements storage.EvaluationStore for ClickHouse
type EvaluationStore struct {
	conn clickhouse.Conn
}

// NewEvaluationStore creates a new EvaluationStore instance
func NewEvaluationStore(conn clickhouse.Conn) storage.EvaluationStore {
	return &EvaluationStore{conn: conn}
}

// StoreEvaluation stores a quote evaluation result
func (s *EvaluationStore) StoreEvaluation(ctx context.Context, eval *models.QuoteEvaluation) error {
	// Using the existing InsertQuoteEvaluation query from queries.go
	// This matches what the evaluator service is already doing

	// Convert TCB components to hex strings
	sgxComponents := ""
	tdxComponents := ""
	if eval.TCBComponents != (models.TCBComponents{}) {
		sgxComponents = hex.EncodeToString(eval.TCBComponents.SGXComponents[:])
		tdxComponents = hex.EncodeToString(eval.TCBComponents.TDXComponents[:])
	}

	err := s.conn.Exec(ctx, InsertQuoteEvaluation,
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
		// evaluated_at is set by now64() in the query
		eval.MrTd,
		eval.MrSeam,
		eval.MrSignerSeam,
		eval.ReportData,
	)
	if err != nil {
		return fmt.Errorf("failed to store evaluation: %w", err)
	}

	return nil
}

// GetLastEvaluation gets the last evaluation for a quote (for status change detection)
func (s *EvaluationStore) GetLastEvaluation(ctx context.Context, serviceAddress, quoteHash string) (status, tcbStatus string, err error) {
	err = s.conn.QueryRow(ctx, GetLastEvaluation, serviceAddress, quoteHash).Scan(&status, &tcbStatus)
	if err != nil {
		return "", "", fmt.Errorf("failed to get last evaluation: %w", err)
	}
	return status, tcbStatus, nil
}

// StoreEvaluationHistory stores evaluation history for tracking changes
func (s *EvaluationStore) StoreEvaluationHistory(ctx context.Context, serviceAddress, quoteHash, prevStatus, newStatus, prevTCBStatus, newTCBStatus string) error {
	err := s.conn.Exec(ctx, InsertQuoteEvaluationHistory,
		serviceAddress,
		quoteHash,
		prevStatus,
		newStatus,
		prevTCBStatus,
		newTCBStatus,
	)
	if err != nil {
		return fmt.Errorf("failed to store evaluation history: %w", err)
	}
	return nil
}
