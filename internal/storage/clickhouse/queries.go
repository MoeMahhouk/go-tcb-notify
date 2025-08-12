package clickhouse

const (
	InsertQuoteRaw = `
		INSERT INTO tdx_quotes_raw
		(service_address, block_number, block_time, tx_hash, log_index, quote_bytes, quote_len, quote_sha256, inserted_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, now64())
	`

	SelectQuotesAfter = `
		SELECT service_address, block_number, block_time, tx_hash, log_index, quote_bytes, quote_len, quote_sha256
		FROM tdx_quotes_raw
		WHERE (block_number > ?) OR (block_number = ? AND log_index > ?)
		ORDER BY block_number ASC, log_index ASC
		LIMIT ?
	`

	InsertQuoteParsed = `
		INSERT INTO tdx_quotes_parsed
		(service_address, block_number, block_time, tx_hash, log_index, quote_len, quote_sha256,
		 tee_tcb_svn, mr_seam, mr_signer_seam, seam_svn, attributes, xfam, mr_td, mr_owner, mr_owner_config, mr_config_id, report_data, parsed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now64())
	`

	// Pipeline offsets: ReplacingMergeTree keyed by service
	GetOffset = `
		SELECT last_block, last_log_index
		FROM pipeline_offsets
		WHERE service = ?
		ORDER BY updated_at DESC
		LIMIT 1
	`

	UpsertOffset = `
		INSERT INTO pipeline_offsets (service, last_block, last_log_index, updated_at)
		VALUES (?, ?, ?, now64())
	`

	// PCS TCB Info: ReplacingMergeTree keyed by fmspc (full JSON for forward-compat)
	UpsertPCSTCBInfo = `
		INSERT INTO pcs_tcb_info (fmspc, tcb_evaluation_data_number, issue_date, next_update, raw_json)
		VALUES (?, ?, parseDateTimeBestEffort(?), parseDateTimeBestEffort(?), ?)
	`

	// Per-quote validation results
	InsertQuoteStatus = `
		INSERT INTO tdx_quote_status
		(service_address, block_number, block_time, tx_hash, log_index, quote_sha256, status, reason, checked_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, now64())
	`
)
