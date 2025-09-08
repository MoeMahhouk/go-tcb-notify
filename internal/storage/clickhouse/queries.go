package clickhouse

const (
	// ===================
	// Registry Ingestion (used by ingest-registry)
	// ===================

	InsertQuoteRaw = `
		INSERT INTO registry_quotes_raw
		(service_address, block_number, block_time, tx_hash, log_index, event_type,
		 quote_bytes, quote_len, quote_sha256, fmspc, ingested_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now64())
	`

	InsertInvalidationEvent = `
		INSERT INTO registry_quotes_raw
		(service_address, block_number, block_time, tx_hash, log_index, event_type, ingested_at)
		VALUES (?, ?, ?, ?, ?, 'Invalidated', now64())
	`

	// ===================
	// Quote Evaluation (used by evaluate-quotes)
	// ===================

	// Get ALL currently active (non-invalidated) quotes for evaluation
	GetActiveQuotes = `
		WITH latest_events AS (
			SELECT 
				service_address,
				argMax(event_type, block_number) as latest_event_type,
				argMax(block_number, block_number) as block_number,
				argMax(block_time, block_number) as block_time,
				argMax(tx_hash, block_number) as tx_hash,
				argMax(log_index, block_number) as log_index,
				argMax(quote_bytes, block_number) as quote_bytes,
				argMax(quote_len, block_number) as quote_len,
				argMax(quote_sha256, block_number) as quote_sha256
			FROM registry_quotes_raw
			GROUP BY service_address
		)
		SELECT 
			service_address, 
			block_number, 
			block_time, 
			tx_hash, 
			log_index, 
			quote_bytes, 
			quote_len, 
			quote_sha256
		FROM latest_events
		WHERE latest_event_type = 'Registered'
		ORDER BY service_address, block_number DESC
	`
	// Count registered quotes affected by FMSPC change (only active quotes)
	CountAffectedRegisteredQuotes = `
		WITH latest_events AS (
			SELECT 
				service_address,
				argMax(event_type, block_number) as latest_event_type,
				argMax(fmspc, block_number) as fmspc
			FROM registry_quotes_raw
			WHERE fmspc = ?
			GROUP BY service_address
		)
		SELECT COUNT(*)
		FROM latest_events
		WHERE latest_event_type = 'Registered'
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

	// Get all active FMSPCs
	GetAllFMSPCs = `
		SELECT DISTINCT fmspc 
		FROM pcs_fmspcs 
		WHERE is_active = 1
	`

	// Update FMSPC last seen time
	UpdateFMSPCSeen = `
		INSERT INTO pcs_fmspcs (fmspc, last_seen)
		VALUES (?, ?)
	`

	// ===================
	// Alert Management Queries
	// ===================

	// Get pending alerts (recent alerts that need attention)
	GetPendingAlerts = `
		SELECT fmspc, old_eval_number, new_eval_number, 
			   affected_quotes_count, details, created_at, acknowledged
		FROM tcb_change_alerts
		WHERE acknowledged = false
		ORDER BY created_at DESC
	`

	// Get recent alerts (last 24 hours)
	GetRecentAlerts = `
		SELECT fmspc, old_eval_number, new_eval_number, 
			   affected_quotes_count, details, created_at, acknowledged
		FROM tcb_change_alerts
		WHERE created_at > now() - INTERVAL 24 HOUR
		ORDER BY created_at DESC
		LIMIT ?
	`

	// Acknowledge an alert by FMSPC and timestamp
	AcknowledgeAlert = `
		ALTER TABLE tcb_change_alerts 
		UPDATE acknowledged = true 
		WHERE fmspc = ? AND created_at = ?
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
