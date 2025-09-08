package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
)

// QuoteStore implements storage.QuoteStore interface for ClickHouse
type QuoteStore struct {
	conn clickhouse.Conn
}

// NewQuoteStore creates a new ClickHouse quote store
func NewQuoteStore(conn clickhouse.Conn) *QuoteStore {
	return &QuoteStore{conn: conn}
}

// StoreRawQuote stores a raw quote from the registry
func (s *QuoteStore) StoreRawQuote(ctx context.Context, quote *models.RegistryQuote) error {
	return s.conn.Exec(ctx, InsertQuoteRaw,
		quote.ServiceAddress,
		quote.BlockNumber,
		quote.BlockTime,
		quote.TxHash,
		quote.LogIndex,
		"Registered", // EventType is always "Registered" for new quotes
		quote.QuoteBytes,
		quote.QuoteLength,
		quote.QuoteHash,
		quote.FMSPC,
	)
}

// StoreInvalidation stores an invalidation event
func (s *QuoteStore) StoreInvalidation(ctx context.Context, serviceAddr string, blockNumber uint64, blockTime time.Time, txHash string, logIndex uint32) error {
	return s.conn.Exec(ctx, InsertInvalidationEvent,
		serviceAddr,
		blockNumber,
		blockTime,
		txHash,
		logIndex,
	)
}

// GetActiveQuotes returns all currently active (non-invalidated) quotes
func (s *QuoteStore) GetActiveQuotes(ctx context.Context) ([]*models.RegistryQuote, error) {
	rows, err := s.conn.Query(ctx, GetActiveQuotes)
	if err != nil {
		return nil, fmt.Errorf("failed to query active quotes: %w", err)
	}
	defer rows.Close()

	var quotes []*models.RegistryQuote
	for rows.Next() {
		var q models.RegistryQuote
		err := rows.Scan(
			&q.ServiceAddress,
			&q.BlockNumber,
			&q.BlockTime,
			&q.TxHash,
			&q.LogIndex,
			&q.QuoteBytes,
			&q.QuoteLength,
			&q.QuoteHash,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quote: %w", err)
		}
		quotes = append(quotes, &q)
	}

	return quotes, rows.Err()
}

// CountAffectedQuotes counts quotes affected by an FMSPC change
func (s *QuoteStore) CountAffectedQuotes(ctx context.Context, fmspc string) (int64, error) {
	var count int64
	err := s.conn.QueryRow(ctx, CountAffectedRegisteredQuotes, fmspc).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count affected quotes: %w", err)
	}
	return count, nil
}
