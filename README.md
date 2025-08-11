# go-tcb-notify

A monitoring and notification service for Intel TDX TCB changes in the Flashtestation protocol ecosystem.

## Overview

go-tcb-notify watches for Intel TDX Trusted Computing Base (TCB) updates and identifies TDX attestations that need revalidation. The service ensures that the Flashtestation registry maintains only TDX attestations that meet current Intel security requirements.

## Architecture

The service consists of four core components:

1. **TDX TCB Fetcher** - Monitors Intel PCS for TCB information updates
2. **TDX Quote Checker** - Identifies quotes impacted by TCB changes  
3. **Alert Publisher** - Notifies DevOps systems about impacted quotes
4. **Metrics Exporter** - Exports Prometheus metrics for monitoring

## Features

- Monitors Intel TDX PCS API for TCB updates
- Compares quotes against new TCB requirements
- Sends detailed alerts with advisory IDs and suggested actions
- PostgreSQL storage for audit trails
- Prometheus metrics integration
- Health check endpoints
- Docker support

## Getting Started

### Prerequisites

- Go 1.24+
- PostgreSQL 12+
- (Optional) Docker / Docker Compose
- An Ethereum RPC endpoint with access to the chain your registry is on
- **ABI JSON** for your deployed Flashtestation Registry contract

### Getting the ABI

Build your contracts with Foundry (`forge build`). Copy the registry artifact to this repo and adjust the right path.

### Installation

1. Clone the repository:
```bash
git clone https://github.com/MoeMahhouk/go-tcb-notify.git
cd go-tcb-notify
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Run database migrations:
```bash
go run cmd/go-tcb-notify/main.go
```

### Configuration

Environment variables:

```bash
# Server
PORT=8080
LOG_LEVEL=info

# Database
DATABASE_URL=postgres://user:pass@localhost/go_tcb_notify?sslmode=disable

# Ethereum
RPC_URL=http://localhost:8545
REGISTRY_ADDRESS=0x...

# Intel PCS
PCS_BASE_URL=https://api.trustedservices.intel.com

# Monitoring
TCB_CHECK_INTERVAL=1h
QUOTE_CHECK_INTERVAL=5m

# Alerting
ALERT_WEBHOOK_URL=https://webhook.site/...

# Metrics
METRICS_ENABLED=true
```

### Running with Docker

```bash
docker-compose up -d
```

## API Endpoints

### Health Checks

- `GET /health` - Service health status
- `GET /ready` - Readiness probe
- `GET /metrics` - Prometheus metrics

### API

- `GET /api/v1/tcb/{fmspc}` - Get TCB information for FMSPC
- `GET /api/v1/quotes` - List monitored quotes

## Database Schema

The service uses PostgreSQL with three main tables:

- `tdx_tcb_info` - Intel TCB information
- `monitored_tdx_quotes` - Tracked quotes from registry
- `alert_history` - Alert audit trail

## Integration

### With google/go-tdx-guest

The service integrates with the `google/go-tdx-guest` library for:

- TCB information parsing
- Quote verification
- Component comparison logic

### With Flashtestation Registry

- Monitors registered TDX quotes
- Retrieves quote data for verification
- Identifies impacted attestations

## Monitoring

### Prometheus Metrics

- `tdx_tcb_notify_checks_total` - Total TCB checks performed
- `tdx_tcb_notify_quotes_impacted` - Current impacted quote count
- `tdx_tcb_notify_alerts_sent_total` - Total alerts sent
- `tdx_tcb_notify_api_errors_total` - TDX PCS API error count

### Alerting

Alerts are sent via webhook with the following structure:

```json
{
  "severity": "warning",
  "source": "go-tcb-notify",
  "timestamp": "2024-01-20T10:30:00Z",
  "quote": {
    "address": "0x...",
    "reason": "TDX TCB status changed",
    "previousStatus": "UpToDate",
    "newStatus": "OutOfDate",
    "workloadId": "0x...",
    "tcbEvaluationDataNumber": 13,
    "advisoryIDs": ["INTEL-SA-00586", "INTEL-SA-00615"],
    "fmspc": "50806F000000",
    "suggestedAction": "invalidate_attestation"
  }
}
```

## Development

### Building

```bash
go build -o go-tcb-notify ./cmd/go-tcb-notify
```

### Testing

```bash
go test ./...
```

### Adding New Features

1. Implement in appropriate service package
2. Add tests
3. Update configuration if needed
4. Update documentation

## Security Considerations

- PostgreSQL connections use SSL/TLS in production
- No private keys or sensitive data stored
- Quote data stored for verification only
- Intel PCS API uses HTTPS only

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

[License information to be added]

## References

- [Flashtestation Specification](https://github.com/flashbots/rollup-boost/blob/main/specs/flashtestations.md)
- [Intel TDX Documentation](https://www.intel.com/content/www/us/en/developer/tools/trust-domain-extensions/documentation.html)
- [google/go-tdx-guest](https://github.com/google/go-tdx-guest)