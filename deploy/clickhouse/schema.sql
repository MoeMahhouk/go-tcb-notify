-- Database
CREATE DATABASE IF NOT EXISTS go_tcb_notify;

USE go_tcb_notify;

-- ===================
-- Registry Tables
-- ===================

-- Raw quotes from blockchain events
CREATE TABLE IF NOT EXISTS registry_quotes_raw
(
  service_address String,
  block_number    UInt64,
  block_time      DateTime64(3, 'UTC'),
  tx_hash         String,
  log_index       UInt32,
  quote_bytes     String,  -- hex encoded
  quote_len       UInt32,
  quote_sha256    FixedString(64),
  ingested_at     DateTime64(3, 'UTC') DEFAULT now64()
)
ENGINE = MergeTree
ORDER BY (block_number, log_index)
PARTITION BY toYYYYMM(block_time);

-- Parsed quote data with extracted components
CREATE TABLE IF NOT EXISTS tdx_quotes_parsed
(
  service_address  String,
  block_number     UInt64,
  block_time       DateTime64(3, 'UTC'),
  tx_hash          String,
  log_index        UInt32,
  quote_len        UInt32,
  quote_sha256     FixedString(64),
  
  -- Extracted components
  fmspc            String,
  sgx_components   String,  -- hex encoded 16 bytes
  tdx_components   String,  -- hex encoded 16 bytes
  pce_svn          UInt16,
  
  -- Measurements
  mr_td            String,
  mr_seam          String,
  mr_signer_seam   String,
  report_data      String,
  
  parsed_at        DateTime64(3, 'UTC')
)
ENGINE = MergeTree
ORDER BY (block_number, log_index)
PARTITION BY toYYYYMM(block_time);

-- Quote evaluation results
CREATE TABLE IF NOT EXISTS tdx_quote_evaluations
(
  service_address  String,
  quote_hash       String,
  quote_length     UInt32,
  fmspc            String,
  sgx_components   String,
  tdx_components   String,
  pce_svn          UInt16,
  mr_td            String,
  mr_seam          String,
  mr_signer_seam   String,
  report_data      String,
  status           String,  -- VALID, INVALID, VALID_SIGNATURE
  tcb_status       String,  -- UpToDate, OutOfDate, SWHardeningNeeded, etc.
  error            Nullable(String),
  block_number     UInt64,
  log_index        UInt32,
  block_time       DateTime64(3, 'UTC'),
  evaluated_at     DateTime64(3, 'UTC')
)
ENGINE = MergeTree
ORDER BY (service_address, evaluated_at)
PARTITION BY toYYYYMM(evaluated_at);

-- ===================
-- Intel PCS Global Tracking Tables
-- ===================

-- All FMSPCs available from Intel PCS (global tracking)
CREATE TABLE IF NOT EXISTS pcs_fmspcs
(
  fmspc            String,
  platform         String,  -- SGX, TDX, or ALL
  first_seen       DateTime64(3, 'UTC'),
  last_seen        DateTime64(3, 'UTC'),
  is_active        UInt8 DEFAULT 1
)
ENGINE = ReplacingMergeTree(last_seen)
ORDER BY fmspc;

-- TCB Info for ALL FMSPCs from Intel (not just registered quotes)
CREATE TABLE IF NOT EXISTS pcs_tcb_info
(
  fmspc                        String,
  tcb_evaluation_data_number   UInt32,
  issue_date                   DateTime64(3, 'UTC'),
  next_update                  DateTime64(3, 'UTC'),
  tcb_type                     UInt32,
  tcb_levels_json              String,  -- JSON array of TCB levels
  raw_json                     String,  -- Full response from Intel
  fetched_at                   DateTime64(3, 'UTC')
)
ENGINE = ReplacingMergeTree(fetched_at)
ORDER BY (fmspc, tcb_evaluation_data_number)
PARTITION BY toYYYYMM(fetched_at);

-- Track changes in TCB components over time
CREATE TABLE IF NOT EXISTS tcb_component_changes
(
  fmspc            String,
  component_index  UInt32,
  component_name   String,
  component_type   String,  -- SGX or TDX
  old_version      UInt32,
  new_version      UInt32,
  change_type      String,  -- UPGRADE or DOWNGRADE
  evaluation_number UInt32,
  detected_at      DateTime64(3, 'UTC')
)
ENGINE = MergeTree
ORDER BY (fmspc, detected_at, component_index)
PARTITION BY toYYYYMM(detected_at);

-- ===================
-- Comparison Tables (Registry vs Intel PCS)
-- ===================

-- Track quotes that need re-evaluation due to TCB updates
CREATE TABLE IF NOT EXISTS quotes_needing_reevaluation
(
  service_address     String,
  fmspc              String,
  current_tcb_status String,
  latest_intel_eval  UInt32,
  last_evaluated_eval UInt32,
  reason             String,
  detected_at        DateTime64(3, 'UTC')
)
ENGINE = ReplacingMergeTree(detected_at)
ORDER BY (service_address, fmspc);

-- TCB status comparison between registered quotes and Intel's latest
CREATE TABLE IF NOT EXISTS tcb_status_comparison
(
  fmspc                   String,
  registered_quote_count  UInt32,
  latest_intel_eval_num   UInt32,
  outdated_quote_count    UInt32,
  uptodate_quote_count    UInt32,
  needs_attention_count   UInt32,
  comparison_date         DateTime64(3, 'UTC')
)
ENGINE = ReplacingMergeTree(comparison_date)
ORDER BY fmspc;

-- ===================
-- Alert Management
-- ===================

CREATE TABLE IF NOT EXISTS tcb_alerts
(
  id                    UInt64,
  service_address       Nullable(String),  -- NULL for global alerts
  alert_type            String,  -- TCB_UPDATE, STATUS_CHANGE, NEW_ADVISORY
  severity              String,  -- CRITICAL, HIGH, MEDIUM, LOW
  fmspc                 String,
  tcb_evaluation_number UInt32,
  previous_status       Nullable(String),
  new_status            Nullable(String),
  advisory_ids          Array(String),
  details               String,  -- JSON
  created_at            DateTime64(3, 'UTC'),
  acknowledged          UInt8 DEFAULT 0,
  acknowledged_at       Nullable(DateTime64(3, 'UTC'))
)
ENGINE = MergeTree
ORDER BY (created_at, id)
PARTITION BY toYYYYMM(created_at);

-- ===================
-- Pipeline Management
-- ===================

CREATE TABLE IF NOT EXISTS pipeline_offsets
(
  service         String,
  last_block      UInt64,
  last_log_index  UInt32,
  updated_at      DateTime64(3, 'UTC')
)
ENGINE = ReplacingMergeTree(updated_at)
ORDER BY service;

-- ===================
-- Analytics Views
-- ===================

-- Latest quote status per address
CREATE VIEW IF NOT EXISTS latest_quote_status AS
SELECT 
  service_address,
  argMax(fmspc, evaluated_at) as fmspc,
  argMax(tcb_status, evaluated_at) as tcb_status,
  argMax(status, evaluated_at) as quote_status,
  max(evaluated_at) as last_evaluated
FROM tdx_quote_evaluations
GROUP BY service_address;

-- FMSPC coverage analysis
CREATE VIEW IF NOT EXISTS fmspc_coverage AS
SELECT 
  pf.fmspc,
  pf.platform,
  countDistinct(qp.service_address) as registered_quote_count,
  max(ti.tcb_evaluation_data_number) as latest_tcb_eval,
  max(ti.fetched_at) as last_tcb_update
FROM pcs_fmspcs pf
LEFT JOIN tdx_quotes_parsed qp ON pf.fmspc = qp.fmspc
LEFT JOIN pcs_tcb_info ti ON pf.fmspc = ti.fmspc
GROUP BY pf.fmspc, pf.platform
ORDER BY registered_quote_count DESC;

-- Global TCB health dashboard
CREATE VIEW IF NOT EXISTS tcb_health_dashboard AS
WITH latest_evaluations AS (
  SELECT 
    fmspc,
    argMax(tcb_status, evaluated_at) as current_status,
    max(evaluated_at) as last_check
  FROM tdx_quote_evaluations
  GROUP BY fmspc
)
SELECT 
  current_status,
  count() as count,
  round(count() * 100.0 / sum(count()) OVER (), 2) as percentage
FROM latest_evaluations
GROUP BY current_status
ORDER BY count DESC;

-- Component version distribution across all quotes
CREATE VIEW IF NOT EXISTS component_version_distribution AS
WITH latest_quotes AS (
  SELECT 
    service_address,
    argMax(sgx_components, block_number) as sgx,
    argMax(tdx_components, block_number) as tdx
  FROM tdx_quotes_parsed
  GROUP BY service_address
)
SELECT 
  'SGX' as component_type,
  arrayJoin(range(0, 16)) as component_index,
  hex(substring(sgx, component_index + 1, 1)) as version,
  count() as count
FROM latest_quotes
GROUP BY component_type, component_index, version

UNION ALL

SELECT 
  'TDX' as component_type,
  arrayJoin(range(0, 16)) as component_index,
  hex(substring(tdx, component_index + 1, 1)) as version,
  count() as count
FROM latest_quotes
GROUP BY component_type, component_index, version
ORDER BY component_type, component_index, version;