package clickhouse

const (
	// ===================
	// Registry Ingestion (used by ingest-registry)
	// ===================

	InsertQuoteRaw = `
		INSERT INTO registry_quotes_raw
		(service_address, block_number, block_time, tx_hash, log_index, 
		 quote_bytes, quote_len, quote_sha256, fmspc, ingested_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, now64())
	`

	// ===================
	// Quote Evaluation (used by evaluate-quotes)
	// ===================

	// Get ALL currently registered quotes for evaluation
	GetAllRegisteredQuotes = `
		SELECT 
			service_address, 
			block_number, 
			block_time, 
			tx_hash, 
			log_index, 
			quote_bytes, 
			quote_len, 
			quote_sha256
		FROM registry_quotes_raw
		ORDER BY service_address, block_number DESC
	`

	// Insert or update evaluation result
	InsertQuoteEvaluation = `
		INSERT INTO tdx_quote_evaluations
		(service_address, quote_hash, quote_length, 
		 status, tcb_status, error_message,
		 fmspc, sgx_components, tdx_components, pce_svn,
		 block_number, log_index, block_time, evaluated_at,
		 mr_td, mr_seam, mr_signer_seam, report_data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now64(), ?, ?, ?, ?)
	`

	// Track evaluation history (for tracking status changes)
	InsertQuoteEvaluationHistory = `
		INSERT INTO tdx_quote_evaluation_history
		(service_address, quote_hash, previous_status, new_status, 
		 previous_tcb_status, new_tcb_status, changed_at)
		VALUES (?, ?, ?, ?, ?, ?, now64())
	`

	// Get the last evaluation for a quote (to detect status changes)
	GetLastEvaluation = `
		SELECT status, tcb_status
		FROM tdx_quote_evaluations
		WHERE service_address = ? AND quote_hash = ?
		ORDER BY evaluated_at DESC
		LIMIT 1
	`

	// ===================
	// Intel PCS Management (used by fetch-pcs)
	// ===================

	// Store ALL FMSPCs from Intel PCS
	UpsertPCSFMSPC = `
		INSERT INTO pcs_fmspcs 
		(fmspc, platform, first_seen, last_seen, is_active)
		VALUES (?, ?, now64(), now64(), 1)
	`

	// Insert new TCB info from Intel
	InsertPCSTCBInfo = `
		INSERT INTO pcs_tcb_info 
		(fmspc, tcb_evaluation_data_number, issue_date, next_update, 
		 tcb_type, tcb_levels_json, raw_json, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, now64())
	`

	// Check if TCB has been updated for an FMSPC
	CheckTCBUpdate = `
		SELECT tcb_evaluation_data_number
		FROM pcs_tcb_info
		WHERE fmspc = ?
		ORDER BY tcb_evaluation_data_number DESC
		LIMIT 1
	`

	// Insert TCB change alert
	InsertTCBChangeAlert = `
		INSERT INTO tcb_change_alerts
		(fmspc, old_eval_number, new_eval_number, 
		 affected_quotes_count, details, created_at)
		VALUES (?, ?, ?, ?, ?, now64())
	`

	// Count registered quotes affected by FMSPC change
	CountAffectedRegisteredQuotes = `
		SELECT COUNT(DISTINCT service_address)
		FROM registry_quotes_raw
		WHERE fmspc = ?
	`

	// ===================
	// Pipeline State Management
	// ===================

	// Get offset for a service (used by ingest-registry to resume)
	GetOffset = `
		SELECT last_block, last_log_index
		FROM pipeline_offsets
		WHERE service = ?
		ORDER BY updated_at DESC
		LIMIT 1
	`

	// Update offset for a service
	UpsertOffset = `
		INSERT INTO pipeline_offsets 
		(service, last_block, last_log_index, updated_at)
		VALUES (?, ?, ?, now64())
	`
)
