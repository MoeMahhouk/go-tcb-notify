package database

import (
	"database/sql"
	"fmt"
)

const (
	createTCBInfoTable = `
	CREATE TABLE IF NOT EXISTS tdx_tcb_info (
		fmspc VARCHAR(16) NOT NULL,
		version INTEGER NOT NULL,
		issue_date TIMESTAMP NOT NULL,
		next_update TIMESTAMP NOT NULL,
		tcb_type INTEGER NOT NULL,
		tcb_evaluation_data_number INTEGER NOT NULL,
		tcb_levels JSONB NOT NULL,
		raw_response JSONB NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (fmspc, tcb_evaluation_data_number)
	);`

	createMonitoredQuotesTable = `
	CREATE TABLE IF NOT EXISTS monitored_tdx_quotes (
		address VARCHAR(42) PRIMARY KEY,
		quote_data BYTEA NOT NULL,
		workload_id VARCHAR(64) NOT NULL,
		fmspc VARCHAR(16) NOT NULL,
		tcb_components JSONB NOT NULL,
		current_status VARCHAR(32) NOT NULL,
		needs_update BOOLEAN DEFAULT FALSE,
		last_checked TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	createAlertHistoryTable = `
	CREATE TABLE IF NOT EXISTS alert_history (
		id SERIAL PRIMARY KEY,
		quote_address VARCHAR(42) NOT NULL,
		reason VARCHAR(255) NOT NULL,
		details JSONB NOT NULL,
		sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		acknowledged BOOLEAN DEFAULT FALSE
	);`

	createIndexes = `
	CREATE INDEX IF NOT EXISTS idx_tcb_info_fmspc ON tdx_tcb_info(fmspc);
	CREATE INDEX IF NOT EXISTS idx_quotes_fmspc ON monitored_tdx_quotes(fmspc);
	CREATE INDEX IF NOT EXISTS idx_quotes_needs_update ON monitored_tdx_quotes(needs_update);
	CREATE INDEX IF NOT EXISTS idx_alert_history_address ON alert_history(quote_address);
	`
)

func Migrate(db *sql.DB) error {
	migrations := []string{
		createTCBInfoTable,
		createMonitoredQuotesTable,
		createAlertHistoryTable,
		createIndexes,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to run migration: %w", err)
		}
	}

	return nil
}