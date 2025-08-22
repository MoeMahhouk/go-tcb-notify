# go-tcb-notify

> **⚠️ DRAFT STATUS**: This project is currently in draft mode and under active development. The architecture, APIs, and implementation details may change drastically before being ready for production use.

An ETL pipeline for monitoring Intel TDX TCB updates and evaluating TDX attestation quotes from the Flashtestation protocol. The system provides automated monitoring and alerting when Intel updates TCB requirements that affect previously valid attestations.

## Overview

When Intel updates TDX TCB requirements due to security vulnerabilities or patches, previously valid TDX attestations may become outdated. This system monitors these changes and evaluates quote validity in real-time using a storage-centric architecture built on ClickHouse.

## Components

- **ingest-registry**: Monitors FlashtestationRegistry smart contract events and stores raw quote data
- **fetch-pcs**: Retrieves TCB information from Intel PCS API and detects updates
- **evaluate-quotes**: Validates TDX quotes against current TCB requirements

## Quick Start

### Prerequisites
- Go 1.24+
- Docker and Docker Compose
- `abigen`, `forge`, and `jq` (for contract bindings)

### Run with Docker

```bash
# Clone and setup
git clone <repository-url>
cd go-tcb-notify
cp .env.example .env
# Edit .env with your configuration

# Start all services
docker compose up --build -d
```

This starts ClickHouse and all pipeline services with proper dependencies.

### Development Build

```bash
# Generate contract bindings (if needed)
make bindings

# Build all services
make build

# Run individual services
./bin/ingest-registry
./bin/fetch-pcs
./bin/evaluate-quotes
```

## Configuration

Configuration is managed through environment variables. Key settings:

```bash
# Ethereum
ETHEREUM_RPC_URL=http://your-rpc-endpoint:8545
REGISTRY_ADDRESS=0x0000000000000000000000000000000000000000
START_BLOCK=0

# ClickHouse (Native Protocol)
CH_ADDRS=clickhouse:9000
CH_DATABASE=default
CH_USERNAME=default
CH_PASSWORD=

# Service intervals
INGEST_POLL_INTERVAL=15s
INGEST_BATCH_BLOCKS=128
EVAL_POLL_INTERVAL=60s
EVAL_GET_COLLATERAL=true
EVAL_CHECK_REVOCATIONS=true
PCS_POLL_INTERVAL=1h
```

See `.env.example` for complete configuration options.

## Monitoring

The system stores all data in ClickHouse, enabling flexible monitoring and alerting:

```sql
-- Check quotes needing attention
SELECT service_address, current_tcb_status, tcb_updated_at
FROM quotes_needing_reevaluation_due_to_tcb
WHERE tcb_updated_at > now() - INTERVAL 1 HOUR;

-- Monitor processing rates
SELECT toStartOfHour(ingested_at) as hour, count() as quotes_processed
FROM registry_quotes_raw
WHERE ingested_at > now() - INTERVAL 24 HOUR
GROUP BY hour;
```

## Architecture

For detailed architecture information, database schema, API specifications, and deployment guidance, see [docs/architecture.md](docs/architecture.md).

## Development

### Project Structure
```
├── cmd/                    # Service entry points
├── internal/               # Internal packages
├── pkg/                   # Public packages
├── deploy/                # Deployment configurations
└── docs/                  # Documentation
```

### Contract Bindings

When the FlashtestationRegistry contract changes:
```bash
make bindings
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## References

- [Architecture Documentation](docs/architecture.md)
- [Intel TDX PCS API](https://api.portal.trustedservices.intel.com/content/documentation.html)
- [Flashtestation Specification](https://github.com/flashbots/rollup-boost/blob/main/specs/flashtestations.md)
- [go-tdx-guest Library](https://github.com/google/go-tdx-guest)
