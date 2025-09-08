package clickhouse

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/storage"
)

// AlertStore implements storage.AlertStore for ClickHouse
type AlertStore struct {
	conn clickhouse.Conn
}

// NewAlertStore creates a new AlertStore instance
func NewAlertStore(conn clickhouse.Conn) storage.AlertStore {
	return &AlertStore{conn: conn}
}

// CreateAlert creates a new TCB change alert
func (s *AlertStore) CreateAlert(ctx context.Context, alert *models.TCBAlert) error {
	// Using the existing InsertTCBChangeAlert query from queries.go
	err := s.conn.Exec(ctx, InsertTCBChangeAlert,
		alert.FMSPC,
		alert.OldEvalNumber,
		alert.NewEvalNumber,
		alert.AffectedQuotesCount,
		alert.Details,
		// created_at is set by now64() in the query
	)
	if err != nil {
		return fmt.Errorf("failed to create TCB alert: %w", err)
	}

	return nil
}

// GetPendingAlerts returns alerts that haven't been acknowledged
func (s *AlertStore) GetPendingAlerts(ctx context.Context) ([]*models.TCBAlert, error) {
	// Using GetPendingAlerts query from queries.go
	rows, err := s.conn.Query(ctx, GetPendingAlerts)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*models.TCBAlert
	for rows.Next() {
		var alert models.TCBAlert
		err := rows.Scan(
			&alert.FMSPC,
			&alert.OldEvalNumber,
			&alert.NewEvalNumber,
			&alert.AffectedQuotesCount,
			&alert.Details,
			&alert.CreatedAt,
			&alert.Acknowledged,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert row: %w", err)
		}
		alerts = append(alerts, &alert)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alert rows: %w", err)
	}

	return alerts, nil
}

// GetRecentAlerts returns recent alerts (last N alerts)
func (s *AlertStore) GetRecentAlerts(ctx context.Context, limit int) ([]*models.TCBAlert, error) {
	// Using GetRecentAlerts query from queries.go
	rows, err := s.conn.Query(ctx, GetRecentAlerts, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*models.TCBAlert
	for rows.Next() {
		var alert models.TCBAlert
		err := rows.Scan(
			&alert.FMSPC,
			&alert.OldEvalNumber,
			&alert.NewEvalNumber,
			&alert.AffectedQuotesCount,
			&alert.Details,
			&alert.CreatedAt,
			&alert.Acknowledged,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert row: %w", err)
		}
		alerts = append(alerts, &alert)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alert rows: %w", err)
	}

	return alerts, nil
}
