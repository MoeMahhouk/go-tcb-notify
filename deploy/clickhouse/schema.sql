-- Database (create if not exists)
CREATE DATABASE IF NOT EXISTS go_tcb_notify;

-- Landing zone for on-chain quotes
CREATE TABLE IF NOT EXISTS tdx_quotes_raw
(
  service_address String,                      -- hex-encoded 0x…
  block_number    UInt64,
  block_time      DateTime64(3, 'UTC'),
  tx_hash         String,
  log_index       UInt32,
  quote_bytes     String,                      -- opaque, raw bytes
  quote_len       UInt32,
  quote_sha256    FixedString(64),
  inserted_at     DateTime64(3, 'UTC')
)
ENGINE = MergeTree
ORDER BY (block_number, tx_hash, log_index);

-- Lightly parsed TDX quote metadata (names align with common TDX v4 fields)
-- We keep them Nullable for forward compatibility; parsing can improve over time.
CREATE TABLE IF NOT EXISTS tdx_quotes_parsed
(
  service_address  String,
  block_number     UInt64,
  block_time       DateTime64(3, 'UTC'),
  tx_hash          String,
  log_index        UInt32,
  quote_len        UInt32,
  quote_sha256     FixedString(64),

  tee_tcb_svn      Nullable(String),
  mr_seam          Nullable(String),
  mr_signer_seam   Nullable(String),
  seam_svn         Nullable(String),
  attributes       Nullable(String),
  xfam             Nullable(String),

  mr_td            Nullable(String),
  mr_owner         Nullable(String),
  mr_owner_config  Nullable(String),
  mr_config_id     Nullable(String),
  report_data      Nullable(String),

  parsed_at        DateTime64(3, 'UTC')
)
ENGINE = MergeTree
ORDER BY (block_number, tx_hash, log_index);

-- PCS TCB Info by FMSPC (raw JSON + key dates)
CREATE TABLE IF NOT EXISTS pcs_tcb_info
(
  fmspc                           String,
  tcb_evaluation_data_number      UInt32,
  issue_date                      Date,
  next_update                     Date,
  raw_json                        String,
  fetched_at                      DateTime64(3, 'UTC')
)
ENGINE = ReplacingMergeTree(fetched_at)
ORDER BY fmspc;

-- Service checkpointing (pipeline offsets)
CREATE TABLE IF NOT EXISTS pipeline_offsets
(
  service         String,
  last_block      UInt64,
  last_log_index  UInt32,
  updated_at      DateTime64(3, 'UTC')
)
ENGINE = ReplacingMergeTree(updated_at)
ORDER BY service;

-- Per-quote validation results
CREATE TABLE IF NOT EXISTS tdx_quote_status
(
  service_address  String,
  block_number     UInt64,
  block_time       DateTime64(3, 'UTC'),
  tx_hash          String,
  log_index        UInt32,
  quote_sha256     FixedString(64),
  status           String,              -- "Verified" | "Invalid" (extend as needed)
  reason           Nullable(String),    -- error/why invalid
  checked_at       DateTime64(3, 'UTC')
)
ENGINE = ReplacingMergeTree(checked_at)
ORDER BY (service_address, block_number, log_index);
