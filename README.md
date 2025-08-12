# go-tcb-notify (ETL/Pipeline Edition)

A small set of Go services that ingest TDX attestation registrations from the **FlashtestationRegistry** smart contract, fetch Intel **TDX TCB** information from the public PCS API, and evaluate quotes against the latest TCB levels. Results are written to **ClickHouse** for downstream consumers (dashboards, alert workers, etc).

> This refactor removes direct webhooks and moves to a storage‑centric pipeline. Each step is its own process.

## Components

- **ingest-registry**: scans Ethereum logs from `FlashtestationRegistry` and stores raw events in ClickHouse.
- **fetch-pcs**: periodically fetches FMSPCs and TDX TCB info from Intel PCS and stores raw JSON.
- **evaluate-quotes**: computes each address’s latest status by parsing quotes and comparing against the newest TCB.

## Why ClickHouse?

ClickHouse offers fast, columnar analytics and easy upserts with `ReplacingMergeTree`. We keep raw, append‑only logs and derive “latest” views using simple queries (`argMax`) or periodic materialization.

## Build

- Requires Go **1.24+**, `abigen`, `forge`, and `jq` if you want to (re)generate bindings.

```bash
# generate bindings from your flashtestations repo (submodule or sibling)
make bindings

# build binaries
make build
```

The binaries will be placed in `./bin`.

## Run (Docker)

```bash
docker compose up --build -d
```

This starts ClickHouse and runs all three processes.

### Services & env

- **ingest-registry**
  - `ETHEREUM_RPC_URL` – your RPC (we set the experimental endpoint in compose).
  - `REGISTRY_ADDRESS` – FlashtestationRegistry address.
  - `START_BLOCK`, `BATCH_SIZE`, `POLL_INTERVAL` – scan controls.
- **fetch-pcs**
  - `PCS_BASE_URL` – defaults to Intel's public PCS (`https://api.trustedservices.intel.com`).
  - `FMSPC_REFRESH`, `TCB_REFRESH` – how often to refresh lists.
- **evaluate-quotes**
  - `EVALUATE_INTERVAL` – evaluation cadence.

All services read ClickHouse from `CLICKHOUSE_DSN` (e.g. `clickhouse://user:pass@host:9000/db?secure=false`).

> DSN is the simplest approach. The driver also supports programmatic options; we keep all config centralized in `internal/config`.

## Database layout

Created automatically by the services on startup:

- `registry_events_raw` — append‑only on‑chain events:
  - `contract_address`, `event`, `tee_address`, `quote_hex`, `override`, `block_number`, `log_index`, `tx_hash`, `ts`
- `registry_quotes_current` — “latest quote per address” view maintained by the evaluator.
- `tdx_tcb_info` — Intel PCS TCB JSON (versioned by `eval_number`).
- `tdx_quote_status` — computed status per address over time.
- `pipeline_state` — last processed block for the registry ingestor.

Using `ReplacingMergeTree` keeps things simple and avoids accidental duplicates during retries.

## Bindings from your contracts

We don’t hardcode an ABI. Instead, we generate type‑safe Go bindings from your **flashtestations** repo via `abigen`:

```bash
make bindings
```

The target extracts the `.abi` section from the Foundry artifact and writes `internal/registry/bindings/registry.go`.

## Configuration

All env vars are loaded via `internal/config` and shared across commands:

- ClickHouse: `CLICKHOUSE_DSN` **(preferred)** or `CLICKHOUSE_ADDR`, `CLICKHOUSE_DATABASE`, `CLICKHOUSE_USERNAME`, `CLICKHOUSE_PASSWORD`, `CLICKHOUSE_SECURE`, `CLICKHOUSE_SKIP_VERIFY`.
- Ethereum: `ETHEREUM_RPC_URL` (or `RPC_URL`), `REGISTRY_ADDRESS`, `START_BLOCK`, `BATCH_SIZE`, `POLL_INTERVAL`.
- Intel PCS: `PCS_BASE_URL`, `FMSPC_REFRESH`, `TCB_REFRESH`.
- Evaluator: `EVALUATE_INTERVAL`.

## Program flow

1. **ingest-registry** uses the generated filterer to read:
   - `TEEServiceRegistered(address addr, bytes quote, bool override)`
   - `TEEServiceInvalidated(address addr)`
   and stores them in `registry_events_raw` (with `block_number`, `log_index`, `tx_hash` for provenance & dedup).
2. **fetch-pcs** reads FMSPC list from `/sgx/certification/v4/fmspcs?platform=all` and for each value fetches `/tdx/certification/v4/tcb?fmspc=…`. It stores the whole JSON document.
3. **evaluate-quotes** picks the latest quote per address with an `argMax` trick, parses the quote (using your existing `pkg/tdx` utilities), compares components against the latest TCB levels, and writes a status row + the latest quote view.

## Notes

- Leave alerting/automation to downstream consumers watching `tdx_quote_status` (or create a separate alert worker).
- If you need more normalization later (e.g., expanding TCB levels into columns), ClickHouse can ingest from the raw JSON with JSON functions or use a Materialized View.
