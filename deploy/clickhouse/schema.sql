-- ===================
-- Core Tables
-- ===================

-- Raw quotes from blockchain registry
CREATE TABLE IF NOT EXISTS registry_quotes_raw (
    service_address String,
    block_number UInt64,
    block_time DateTime64(3),
    tx_hash String,
    log_index UInt32,
    event_type Enum('Registered' = 1, 'Invalidated' = 2),
    quote_bytes String DEFAULT '',  -- Hex encoded quote (empty for invalidation events)
    quote_len UInt32 DEFAULT 0,
    quote_sha256 String DEFAULT '',
    fmspc String DEFAULT '',  -- Extracted FMSPC for easier querying
    ingested_at DateTime64(3) DEFAULT now64(3)
) ENGINE = MergeTree()
ORDER BY (service_address, block_number, log_index)
PARTITION BY toYYYYMM(block_time)
COMMENT 'All registry events (registrations and invalidations) from blockchain';

-- Quote evaluation results (latest status per quote)
CREATE TABLE IF NOT EXISTS tdx_quote_evaluations (
    service_address String,
    quote_hash String,
    quote_length UInt32,
    status Enum('Valid' = 1, 'Invalid' = 2),
    tcb_status Enum('UpToDate' = 1, 'OutOfDate' = 2, 'Revoked' = 3, 'ConfigurationNeeded' = 4, 'Unknown' = 5),
    error_message String DEFAULT '',
    fmspc String DEFAULT '',
    sgx_components String DEFAULT '',  -- Hex encoded 16 bytes
    tdx_components String DEFAULT '',  -- Hex encoded 16 bytes  
    pce_svn UInt16 DEFAULT 0,
    block_number UInt64,
    log_index UInt32,
    block_time DateTime64(3),
    evaluated_at DateTime64(3) DEFAULT now64(3),
    mr_td String DEFAULT '',
    mr_seam String DEFAULT '',
    mr_signer_seam String DEFAULT '',
    report_data String DEFAULT ''
) ENGINE = ReplacingMergeTree(evaluated_at)
ORDER BY (service_address, quote_hash)
PARTITION BY toYYYYMM(evaluated_at)
COMMENT 'Latest evaluation status for each quote';

-- Quote evaluation history (track all status changes)
CREATE TABLE IF NOT EXISTS tdx_quote_evaluation_history (
    service_address String,
    quote_hash String,
    previous_status String,
    new_status String,
    previous_tcb_status String,
    new_tcb_status String,
    changed_at DateTime64(3) DEFAULT now64(3)
) ENGINE = MergeTree()
ORDER BY (service_address, changed_at)
PARTITION BY toYYYYMM(changed_at)
TTL changed_at + INTERVAL 90 DAY
COMMENT 'History of quote status changes';

-- ===================
-- Intel PCS Tables
-- ===================

-- Store ALL FMSPCs from Intel PCS (global tracking)
CREATE TABLE IF NOT EXISTS pcs_fmspcs (
    fmspc String,
    platform String DEFAULT 'ALL',
    first_seen DateTime64(3) DEFAULT now64(3),
    last_seen DateTime64(3) DEFAULT now64(3),
    is_active Boolean DEFAULT true
) ENGINE = ReplacingMergeTree(last_seen)
ORDER BY fmspc
COMMENT 'All FMSPCs available from Intel PCS';

-- Intel PCS TCB Info (keep all versions for history)
CREATE TABLE IF NOT EXISTS pcs_tcb_info (
    fmspc String,
    tcb_evaluation_data_number UInt32,
    issue_date Date,
    next_update Date,
    tcb_type UInt32 DEFAULT 0,
    tcb_levels_json String DEFAULT '',  -- JSON array of TCB levels
    raw_json String,  -- Complete response from Intel
    fetched_at DateTime64(3, UTC) DEFAULT now64(3)
) ENGINE = MergeTree()  -- Not ReplacingMergeTree - we want history
ORDER BY (fmspc, tcb_evaluation_data_number, fetched_at)
PARTITION BY fmspc
COMMENT 'Historical TCB info from Intel PCS';

-- TCB change alerts (when Intel updates TCB)
CREATE TABLE IF NOT EXISTS tcb_change_alerts (
    id UUID DEFAULT generateUUIDv4(),
    fmspc String,
    old_eval_number UInt32,
    new_eval_number UInt32,
    affected_quotes_count UInt32 DEFAULT 0,
    details String DEFAULT '',
    created_at DateTime64(3) DEFAULT now64(3),
    acknowledged Boolean DEFAULT false
) ENGINE = MergeTree()
ORDER BY (created_at, fmspc)
PARTITION BY toYYYYMM(created_at)
TTL created_at + INTERVAL 30 DAY
COMMENT 'Alerts when Intel PCS TCB info changes';

-- ===================
-- Pipeline Management
-- ===================

-- Pipeline state (for service restart/resume)
CREATE TABLE IF NOT EXISTS pipeline_offsets (
    service String,
    last_block UInt64,
    last_log_index UInt32,
    updated_at DateTime64(3) DEFAULT now64(3)
) ENGINE = ReplacingMergeTree(updated_at)
ORDER BY service
COMMENT 'Tracks last processed block for each service';

-- ===================
-- Monitoring Views
-- ===================

-- Current invalid quotes (for alert systems)
CREATE MATERIALIZED VIEW IF NOT EXISTS invalid_quotes_current
ENGINE = MergeTree()
ORDER BY (service_address, last_evaluated)
POPULATE AS
WITH latest_evaluations AS (
    SELECT
        service_address,
        argMax(quote_hash, evaluated_at) as quote_hash,
        argMax(status, evaluated_at) as status,
        argMax(tcb_status, evaluated_at) as tcb_status,
        argMax(error_message, evaluated_at) as error_message,
        argMax(fmspc, evaluated_at) as fmspc,
        max(evaluated_at) as last_evaluated
    FROM tdx_quote_evaluations
    GROUP BY service_address
)
SELECT * FROM latest_evaluations
WHERE status = 'Invalid';

-- Recent status changes (last 24 hours)
CREATE MATERIALIZED VIEW IF NOT EXISTS recent_status_changes
ENGINE = MergeTree()
ORDER BY changed_at
TTL changed_at + INTERVAL 7 DAY
POPULATE AS
SELECT * FROM tdx_quote_evaluation_history
WHERE changed_at > now() - INTERVAL 24 HOUR;

-- Recent TCB updates from Intel
CREATE MATERIALIZED VIEW IF NOT EXISTS recent_tcb_updates
ENGINE = MergeTree()
ORDER BY fetched_at
TTL fetched_at + INTERVAL 30 DAY
POPULATE AS
SELECT 
    fmspc,
    tcb_evaluation_data_number,
    issue_date,
    next_update,
    fetched_at
FROM pcs_tcb_info
WHERE fetched_at > now() - INTERVAL 7 DAY;

-- View for affected registered quotes (alerts for external systems)
CREATE MATERIALIZED VIEW IF NOT EXISTS affected_quotes_by_tcb_update
ENGINE = AggregatingMergeTree()
ORDER BY (fmspc, last_tcb_update)
POPULATE AS
WITH latest_events AS (
    SELECT 
        service_address,
        argMax(event_type, block_number) as latest_event_type,
        argMax(fmspc, block_number) as fmspc,
        argMax(quote_sha256, block_number) as quote_hash
    FROM registry_quotes_raw
    GROUP BY service_address
)
SELECT
    r.service_address as service_address,
    r.fmspc as fmspc,
    r.quote_hash as quote_hash,
    argMax(e.status, e.evaluated_at) as current_status,
    argMax(e.tcb_status, e.evaluated_at) as current_tcb_status,
    max(t.tcb_evaluation_data_number) as latest_tcb_eval_number,
    max(t.fetched_at) as last_tcb_update
FROM latest_events r
JOIN pcs_tcb_info t ON r.fmspc = t.fmspc
LEFT JOIN tdx_quote_evaluations e ON r.service_address = e.service_address
WHERE r.latest_event_type = 'Registered'  -- Only include non-invalidated quotes
GROUP BY r.service_address, r.fmspc, r.quote_hash;

-- Alert view: Quotes that need re-evaluation due to TCB updates
CREATE VIEW IF NOT EXISTS quotes_needing_reevaluation_due_to_tcb AS
SELECT
    a.service_address,
    a.fmspc,
    a.quote_hash,
    a.current_status,
    a.current_tcb_status,
    a.latest_tcb_eval_number,
    c.old_eval_number as evaluated_with_tcb_version,
    c.new_eval_number as current_tcb_version,
    c.created_at as tcb_updated_at
FROM affected_quotes_by_tcb_update a
JOIN tcb_change_alerts c ON a.fmspc = c.fmspc
WHERE c.created_at > a.last_tcb_update - INTERVAL 1 DAY
  AND c.new_eval_number > c.old_eval_number
  AND a.current_tcb_status != 'UpToDate'
ORDER BY c.created_at DESC, a.service_address;
