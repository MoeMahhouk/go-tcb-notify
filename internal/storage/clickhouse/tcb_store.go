package clickhouse

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
)

// TCBStore implements storage.TCBStore interface for ClickHouse
type TCBStore struct {
	conn clickhouse.Conn
}

// NewTCBStore creates a new ClickHouse TCB store
func NewTCBStore(conn clickhouse.Conn) *TCBStore {
	return &TCBStore{conn: conn}
}

// StoreTCBInfo stores TCB information from Intel PCS
func (s *TCBStore) StoreTCBInfo(ctx context.Context, info *models.TCBInfo) error {
	// Convert TCB levels to JSON for storage
	tcbLevelsJSON, err := json.Marshal(info.TCBLevels)
	if err != nil {
		return fmt.Errorf("failed to marshal TCB levels: %w", err)
	}

	err = s.conn.Exec(ctx, InsertPCSTCBInfo,
		info.FMSPC,
		info.TCBEvaluationDataNumber,
		info.IssueDate,
		info.NextUpdate,
		info.TCBType,
		string(tcbLevelsJSON),
		info.RawJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to store TCB info for FMSPC %s: %w", info.FMSPC, err)
	}
	return nil
}

// GetLatestTCBInfo returns the most recent TCB info for an FMSPC
func (s *TCBStore) GetLatestTCBInfo(ctx context.Context, fmspc string) (*models.TCBInfo, error) {
	// For now, we only need the evaluation number based on current usage
	// But the interface requires full TCB info, so let's implement it properly
	// even though it might not be used yet

	var evalNum uint32
	err := s.conn.QueryRow(ctx, CheckTCBUpdate, fmspc).Scan(&evalNum)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No TCB info found
		}
		return nil, fmt.Errorf("failed to get TCB info for FMSPC %s: %w", fmspc, err)
	}

	// For now, return minimal TCB info with just the evaluation number
	// This can be expanded later if full info is needed
	return &models.TCBInfo{
		FMSPC:                   fmspc,
		TCBEvaluationDataNumber: evalNum,
	}, nil
}

// GetAllFMSPCs returns all known FMSPCs
func (s *TCBStore) GetAllFMSPCs(ctx context.Context) ([]string, error) {
	rows, err := s.conn.Query(ctx, GetAllFMSPCs)
	if err != nil {
		return nil, fmt.Errorf("failed to query FMSPCs: %w", err)
	}
	defer rows.Close()

	var fmspcs []string
	for rows.Next() {
		var fmspc string
		if err := rows.Scan(&fmspc); err != nil {
			return nil, fmt.Errorf("failed to scan FMSPC: %w", err)
		}
		fmspcs = append(fmspcs, fmspc)
	}

	return fmspcs, rows.Err()
}

// UpdateFMSPCSeen updates the last seen time for an FMSPC
func (s *TCBStore) UpdateFMSPCSeen(ctx context.Context, fmspc string) error {
	err := s.conn.Exec(ctx, UpdateFMSPCSeen, fmspc, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update FMSPC seen time: %w", err)
	}
	return nil
}

// StoreFMSPC stores FMSPC information
func (s *TCBStore) StoreFMSPC(ctx context.Context, fmspc string, platform string) error {
	err := s.conn.Exec(ctx, UpsertPCSFMSPC, fmspc, platform)
	if err != nil {
		return fmt.Errorf("failed to store FMSPC: %w", err)
	}
	return nil
}

// GetCurrentEvalNumber gets the current TCB evaluation number for an FMSPC
func (s *TCBStore) GetCurrentEvalNumber(ctx context.Context, fmspc string) (uint32, error) {
	var evalNum uint32
	err := s.conn.QueryRow(ctx, CheckTCBUpdate, fmspc).Scan(&evalNum)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("no TCB info found for FMSPC %s", fmspc)
		}
		return 0, fmt.Errorf("failed to get current eval number: %w", err)
	}
	return evalNum, nil
}
