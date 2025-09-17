package clickhouse

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// OffsetStore implements storage.OffsetStore interface for ClickHouse
type OffsetStore struct {
	conn clickhouse.Conn
}

// NewOffsetStore creates a new ClickHouse offset store
func NewOffsetStore(conn clickhouse.Conn) *OffsetStore {
	return &OffsetStore{conn: conn}
}

// LoadOffset loads the last processed position for a service
func (s *OffsetStore) LoadOffset(ctx context.Context, service string) (blockNumber uint64, logIndex uint32, err error) {
	err = s.conn.QueryRow(ctx, GetOffset, service).Scan(&blockNumber, &logIndex)
	if err != nil {
		if err == sql.ErrNoRows {
			// No offset found, start from beginning
			return 0, 0, fmt.Errorf("no offset found for service %s", service)
		}
		return 0, 0, fmt.Errorf("failed to load offset for service %s: %w", service, err)
	}
	return blockNumber, logIndex, nil
}

// SaveOffset saves the current processing position for a service
func (s *OffsetStore) SaveOffset(ctx context.Context, service string, blockNumber uint64, logIndex uint32) error {
	err := s.conn.Exec(ctx, UpsertOffset, service, blockNumber, logIndex)
	if err != nil {
		return fmt.Errorf("failed to save offset for service %s: %w", service, err)
	}
	return nil
}
