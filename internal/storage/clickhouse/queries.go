package clickhouse

const (
	// ===================
	// Registry Ingestion
	// ===================

	InsertQuoteRaw = `
		INSERT INTO registry_quotes_raw
		(service_address, block_number, block_time, tx_hash, log_index, 
		 quote_bytes, quote_len, quote_sha256, ingested_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, now64())
	`

	// ===================
	// Quote Evaluation
	// ===================

	GetUnprocessedQuotes = `
		SELECT 
			service_address, block_number, block_time, tx_hash, log_index, 
			quote_bytes, quote_len, quote_sha256
		FROM registry_quotes_raw
		WHERE (block_number > ?) OR (block_number = ? AND log_index > ?)
		ORDER BY block_number ASC, log_index ASC
		LIMIT ?
	`

	InsertQuoteParsed = `
		INSERT INTO tdx_quotes_parsed
		(service_address, block_number, block_time, tx_hash, log_index, 
		 quote_len, quote_sha256, fmspc, sgx_components, tdx_components, pce_svn,
		 mr_td, mr_seam, mr_signer_seam, report_data, parsed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now64())
	`

	InsertEvaluation = `
		INSERT INTO tdx_quote_evaluations
		(service_address, quote_hash, quote_length, fmspc, 
		 sgx_components, tdx_components, pce_svn,
		 mr_td, mr_seam, mr_signer_seam, report_data,
		 status, tcb_status, error,
		 block_number, log_index, block_time, evaluated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// ===================
	// Intel PCS FMSPC Management (Global)
	// ===================

	// Upsert all FMSPCs from Intel PCS
	UpsertPCSFMSPC = `
		INSERT INTO pcs_fmspcs 
		(fmspc, platform, first_seen, last_seen, is_active)
		VALUES (?, ?, now64(), now64(), 1)
	`

	// Get all active FMSPCs from Intel (not just from quotes)
	GetAllPCSFMSPCs = `
		SELECT fmspc, platform
		FROM pcs_fmspcs
		WHERE is_active = 1
		ORDER BY fmspc
	`

	// Mark FMSPCs as inactive if not returned by Intel
	DeactivateStaleFMSPCs = `
		ALTER TABLE pcs_fmspcs
		UPDATE is_active = 0
		WHERE fmspc NOT IN (?)
		  AND is_active = 1
	`

	// ===================
	// Intel PCS TCB Management (Global)
	// ===================

	UpsertPCSTCBInfo = `
		INSERT INTO pcs_tcb_info 
		(fmspc, tcb_evaluation_data_number, issue_date, next_update, 
		 tcb_type, tcb_levels_json, raw_json, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, now64())
	`

	// Get latest TCB info from our cached Intel data (used by evaluate-quotes)
	GetLatestTCBInfo = `
		SELECT 
			fmspc, tcb_evaluation_data_number, tcb_levels_json, raw_json
		FROM pcs_tcb_info
		WHERE fmspc = ?
		ORDER BY tcb_evaluation_data_number DESC
		LIMIT 1
	`

	// Get all latest TCB info for all FMSPCs
	GetAllLatestTCBInfo = `
		SELECT 
			fmspc, 
			argMax(tcb_evaluation_data_number, tcb_evaluation_data_number) as latest_eval_num,
			argMax(tcb_levels_json, tcb_evaluation_data_number) as tcb_levels_json,
			argMax(raw_json, tcb_evaluation_data_number) as raw_json
		FROM pcs_tcb_info
		GROUP BY fmspc
	`

	// ===================
	// Quote-specific FMSPCs
	// ===================

	// Get unique FMSPCs from registered quotes only
	GetRegisteredQuoteFMSPCs = `
		SELECT DISTINCT fmspc
		FROM tdx_quotes_parsed
		WHERE fmspc != ''
		ORDER BY fmspc
	`

	// ===================
	// TCB Change Detection
	// ===================

	GetTCBChanges = `
		WITH latest_tcb AS (
			SELECT 
				fmspc,
				argMax(tcb_evaluation_data_number, fetched_at) as current_eval,
				argMax(tcb_levels_json, fetched_at) as current_levels
			FROM pcs_tcb_info
			GROUP BY fmspc
		),
		previous_tcb AS (
			SELECT 
				t1.fmspc,
				max(t2.tcb_evaluation_data_number) as previous_eval,
				argMax(t2.tcb_levels_json, t2.tcb_evaluation_data_number) as previous_levels
			FROM latest_tcb t1
			JOIN pcs_tcb_info t2 ON t1.fmspc = t2.fmspc
			WHERE t2.tcb_evaluation_data_number < t1.current_eval
			GROUP BY t1.fmspc
		)
		SELECT 
			l.fmspc,
			l.current_eval,
			p.previous_eval,
			l.current_levels,
			p.previous_levels
		FROM latest_tcb l
		JOIN previous_tcb p ON l.fmspc = p.fmspc
		WHERE l.current_eval > p.previous_eval
	`

	InsertTCBComponentChange = `
		INSERT INTO tcb_component_changes
		(fmspc, component_index, component_name, component_type,
		 old_version, new_version, change_type, evaluation_number, detected_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, now64())
	`

	// Get recent TCB changes for alerting
	GetRecentTCBChanges = `
		SELECT 
			fmspc,
			evaluation_number,
			count() as change_count,
			countIf(change_type = 'DOWNGRADE') as downgrade_count
		FROM tcb_component_changes
		WHERE detected_at > now() - INTERVAL 1 HOUR
		GROUP BY fmspc, evaluation_number
		HAVING change_count > 0
	`

	// ===================
	// Comparison Queries (Registry vs Intel)
	// ===================

	// Find quotes that need re-evaluation based on cached Intel data
	GetQuotesNeedingReevaluation = `
		WITH latest_intel AS (
			SELECT 
				fmspc,
				max(tcb_evaluation_data_number) as latest_eval
			FROM pcs_tcb_info
			GROUP BY fmspc
		),
		latest_evaluations AS (
			SELECT 
				service_address,
				fmspc,
				argMax(tcb_status, evaluated_at) as current_status,
				argMax(block_number, evaluated_at) as block_num,
				max(evaluated_at) as last_evaluated
			FROM tdx_quote_evaluations
			GROUP BY service_address, fmspc
		)
		SELECT 
			e.service_address,
			e.fmspc,
			e.current_status,
			i.latest_eval as latest_intel_eval,
			e.last_evaluated
		FROM latest_evaluations e
		JOIN latest_intel i ON e.fmspc = i.fmspc
		WHERE e.current_status != 'UpToDate'
		   OR e.last_evaluated < now() - INTERVAL 1 DAY
	`

	// Insert quote needing re-evaluation
	InsertQuoteNeedingReevaluation = `
		INSERT INTO quotes_needing_reevaluation
		(service_address, fmspc, current_tcb_status, 
		 latest_intel_eval, last_evaluated_eval, reason, detected_at)
		VALUES (?, ?, ?, ?, ?, ?, now64())
	`

	// Update TCB status comparison statistics
	UpdateTCBStatusComparison = `
		INSERT INTO tcb_status_comparison
		(fmspc, registered_quote_count, latest_intel_eval_num,
		 outdated_quote_count, uptodate_quote_count, needs_attention_count,
		 comparison_date)
		SELECT 
			f.fmspc,
			count(DISTINCT q.service_address) as registered_quote_count,
			max(t.tcb_evaluation_data_number) as latest_intel_eval_num,
			countIf(e.tcb_status = 'OutOfDate') as outdated_quote_count,
			countIf(e.tcb_status = 'UpToDate') as uptodate_quote_count,
			countIf(e.tcb_status NOT IN ('UpToDate', 'OutOfDate')) as needs_attention_count,
			now64() as comparison_date
		FROM pcs_fmspcs f
		LEFT JOIN tdx_quotes_parsed q ON f.fmspc = q.fmspc
		LEFT JOIN pcs_tcb_info t ON f.fmspc = t.fmspc
		LEFT JOIN (
			SELECT 
				service_address,
				fmspc,
				argMax(tcb_status, evaluated_at) as tcb_status
			FROM tdx_quote_evaluations
			GROUP BY service_address, fmspc
		) e ON q.service_address = e.service_address
		WHERE f.is_active = 1
		GROUP BY f.fmspc
	`

	// ===================
	// Alert Management
	// ===================

	InsertAlert = `
		INSERT INTO tcb_alerts
		(id, service_address, alert_type, severity, fmspc,
		 tcb_evaluation_number, previous_status, new_status,
		 advisory_ids, details, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now64())
	`

	// Insert global TCB update alert (not tied to specific service)
	InsertGlobalTCBAlert = `
		INSERT INTO tcb_alerts
		(id, alert_type, severity, fmspc, tcb_evaluation_number,
		 details, created_at)
		VALUES (?, 'TCB_UPDATE', ?, ?, ?, ?, now64())
	`

	GetUnacknowledgedAlerts = `
		SELECT 
			id, service_address, alert_type, severity,
			previous_status, new_status, fmspc,
			tcb_evaluation_number, advisory_ids, details,
			created_at
		FROM tcb_alerts
		WHERE acknowledged = 0
		ORDER BY created_at DESC
		LIMIT 100
	`

	// ===================
	// Pipeline State
	// ===================

	GetOffset = `
		SELECT last_block, last_log_index
		FROM pipeline_offsets
		WHERE service = ?
		ORDER BY updated_at DESC
		LIMIT 1
	`

	UpsertOffset = `
		INSERT INTO pipeline_offsets 
		(service, last_block, last_log_index, updated_at)
		VALUES (?, ?, ?, now64())
	`

	// ===================
	// Analytics Queries
	// ===================

	GetTCBStatusDistribution = `
		SELECT 
			tcb_status,
			count() as count,
			count() * 100.0 / sum(count()) OVER () as percentage
		FROM (
			SELECT 
				service_address,
				argMax(tcb_status, evaluated_at) as tcb_status
			FROM tdx_quote_evaluations
			GROUP BY service_address
		)
		GROUP BY tcb_status
		ORDER BY count DESC
	`

	GetFMSPCCoverage = `
		SELECT 
			pf.fmspc,
			pf.platform,
			countDistinct(qp.service_address) as registered_quotes,
			max(ti.tcb_evaluation_data_number) as latest_tcb_eval,
			max(ti.fetched_at) as last_update
		FROM pcs_fmspcs pf
		LEFT JOIN tdx_quotes_parsed qp ON pf.fmspc = qp.fmspc
		LEFT JOIN pcs_tcb_info ti ON pf.fmspc = ti.fmspc
		WHERE pf.is_active = 1
		GROUP BY pf.fmspc, pf.platform
		ORDER BY registered_quotes DESC
	`
)
