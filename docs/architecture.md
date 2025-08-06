# go-tcb-notify - Architecture Document

## Executive Summary

The go-tcb-notify service is a critical component of the Flashtestation protocol that monitors Intel TDX (Trust Domain Extensions) TCB updates and identifies TDX attestations that need revalidation. This service ensures that our registry maintains only TDX attestations that meet current Intel security requirements.

## Problem Statement

Intel periodically updates TCB requirements for TDX as they discover vulnerabilities or release security patches. When these updates occur:
- Previously valid TDX attestations may become outdated
- The Flashtestation registry needs to identify affected TDX attestations
- Affected TDX VMs must re-attest with updated requirements

Without automated monitoring, outdated TDX attestations could remain in the registry, compromising security guarantees.

## High-Level Architecture

```mermaid
graph TB
    subgraph "External Systems"
        PCS[Intel TDX PCS API<br/>Public Endpoints]
        ETH[Ethereum RPC]
    end
    
    subgraph "go-tcb-notify Service"
        F[TDX TCB Fetcher]
        C[TDX Quote Checker]
        A[Alert Publisher]
        M[Metrics Exporter]
    end
    
    subgraph "Data Storage"
        DB[(PostgreSQL)]
    end
    
    subgraph "Integration Points"
        W[Webhook Endpoint]
        R[Flashtestation Registry]
        P[Prometheus]
    end
    
    PCS -->|TDX TCB Info<br/>No Auth Required| F
    F -->|Store TCB| DB
    F -->|TCB Update| C
    ETH -->|Read TDX Quotes| C
    R -->|TDX Quote Data| C
    C -->|Check Status| DB
    C -->|Impact Alert| A
    A -->|Send Alert| W
    M -->|Export| P
    W -->|Trigger| R
```

## 🔧 Core Components

### 1. **TDX TCB Fetcher**
**Purpose**: Monitor Intel PCS for TDX TCB information updates

**Responsibilities**:
- Poll Intel TDX PCS API every hour (configurable)
- Track TCB Evaluation Data Numbers for change detection
- Store complete TDX TCB information with all levels
- Trigger impact analysis for TDX updates

**Key Implementation Details**:
- Uses public Intel TDX endpoints (no authentication required)
- Monitors `/tdx/certification/v4/tcb?fmspc={fmspc}` endpoint
- Tracks `tcbEvaluationDataNumber` for version control

### 2. **TDX Quote Checker**
**Purpose**: Identify TDX quotes impacted by TCB changes

**Responsibilities**:
- Retrieve registered TDX quotes from Flashtestation registry
- Parse TCB components from each TDX quote
- Compare against new TDX TCB level requirements
- Identify status degradation with specific advisory IDs

**TDX TCB Comparison Logic**:
```go
// Check SGX TCB components (underlying platform)
for i := 0; i < 16; i++ {
    if quoteComponent[i] < tcbLevel.SGXTCBComponents[i].SVN {
        return false // Level not met
    }
}
// Check TDX-specific components
for i := 0; i < 16; i++ {
    if tdxComponent[i] < tcbLevel.TDXTCBComponents[i].SVN {
        return false // Level not met
    }
}
```

### 3. **Alert Publisher**
**Purpose**: Notify DevOps systems about impacted TDX quotes

**Enhanced Alert Payload**:
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
    "fmspc": "50806F000000"
  },
  "suggestedAction": "invalidate_attestation"
}
```

## Data Model

### Updated Database Schema

```sql
-- TDX TCB information from Intel
tdx_tcb_info
├── fmspc (varchar)                      -- Platform identifier
├── version (integer)                    -- TCB structure version
├── issue_date (timestamp)    
├── next_update (timestamp)
├── tcb_type (integer)
├── tcb_evaluation_data_number (integer) -- Monotonic version tracker
├── tcb_levels (jsonb)                   -- Array of TDX TCB levels
├── raw_response (jsonb)                 -- Complete API response
└── created_at (timestamp)
    PRIMARY KEY (fmspc, tcb_evaluation_data_number)

-- Monitored TDX quotes from registry
monitored_tdx_quotes
├── address (varchar)         -- TEE-controlled address
├── quote_data (bytea)        -- Raw TDX attestation quote
├── workload_id (varchar)     
├── fmspc (varchar)          
├── tcb_components (jsonb)    -- Extracted TCB values
├── current_status (varchar)  -- Current TCB status
├── needs_update (boolean)    
└── last_checked (timestamp)

-- Alert history for audit
alert_history
├── id (serial)
├── quote_address (varchar)
├── reason (varchar)
├── details (jsonb)           -- Includes advisory IDs
├── sent_at (timestamp)
└── acknowledged (boolean)
```

## Operational Flow

### TDX TCB Update Detection Flow

```mermaid
sequenceDiagram
    participant TCB as TDX TCB Fetcher
    participant PCS as Intel TDX PCS API
    participant DB as PostgreSQL
    participant QC as TDX Quote Checker
    participant AP as Alert Publisher

    loop Every Hour
        TCB->>PCS: GET /tdx/certification/v4/tcb?fmspc={fmspc}
        Note over PCS: No API key required
        PCS-->>TCB: TDX TCB Info with tcbEvaluationDataNumber
        TCB->>DB: Check if tcbEvaluationDataNumber is new
        alt New Evaluation Data Number
            TCB->>DB: Store complete TDX TCB info
            TCB->>QC: Trigger impact analysis
            QC->>DB: Get TDX quotes for FMSPC
            loop For Each TDX Quote
                QC->>QC: Compare components against TDX TCB levels
                alt Status Degraded
                    Note over QC: Extract advisory IDs
                    QC->>DB: Update quote status
                    QC->>AP: Send enhanced alert
                end
            end
        end
    end
```

## API Integration Details

### Intel TDX PCS API v4 Structure

```json
{
  "tcbInfo": {
    "id": "TDX",
    "version": 3,
    "issueDate": "2022-04-13T09:37:45Z",
    "nextUpdate": "2022-05-13T09:37:45Z",
    "fmspc": "50806F000000",
    "pceId": "0000",
    "tcbType": 0,
    "tcbEvaluationDataNumber": 12,
    "tdxModule": {
      "mrsigner": "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
      "attributes": "0000000000000000",
      "attributesMask": "FFFFFFFFFFFFFFFF"
    },
    "tcbLevels": [
      {
        "tcb": {
          "sgxtcbcomponents": [
            {"svn": 1, "category": "BIOS", "type": "Early Microcode Update"},
            // ... 15 more components
          ],
          "pcesvn": 11,
          "tdxtcbcomponents": [
            {"svn": 2, "category": "OS/VMM", "type": "TDX Module"},
            {"svn": 0},
            {"svn": 1, "category": "OS/VMM", "type": "TDX Late Microcode Update"},
            // ... 13 more components
          ]
        },
        "tcbDate": "2021-11-10T00:00:00Z",
        "tcbStatus": "UpToDate",
        "advisoryIDs": ["INTEL-SA-00586"]
      }
    ]
  },
  "signature": "..."
}
```

### Key TDX-Specific Elements
- **`id`**: Always "TDX" for TDX TCB info
- **`tdxModule`**: TDX module identity information
- **`tdxtcbcomponents`**: TDX-specific TCB components (16 components)
- **`sgxtcbcomponents`**: Underlying platform components (still needed for TDX)

## Monitoring & Observability

### Prometheus Metrics
- `tdx_tcb_notify_checks_total` - Total TDX TCB checks performed
- `tdx_tcb_notify_quotes_impacted` - Current impacted TDX quote count
- `tdx_tcb_notify_alerts_sent_total` - Total alerts sent
- `tdx_tcb_notify_api_errors_total` - TDX PCS API error count

### Health Endpoints
- `/health` - Overall service health
- `/metrics` - Prometheus metrics
- `/ready` - Kubernetes readiness

## Deployment

### Environment Configuration
```bash
# Core Configuration
RPC_URL=http://ethereum-rpc:8545
REGISTRY_ADDRESS=0x...
DATABASE_URL=postgres://user:pass@host/go_tcb_notify

# Intel TDX PCS (No API key needed)
PCS_BASE_URL=https://api.trustedservices.intel.com

# Intervals
TCB_CHECK_INTERVAL=1h
QUOTE_CHECK_INTERVAL=5m

# Alerting
ALERT_WEBHOOK_URL=https://webhook.site/...
```

### Kubernetes Resources
```yaml
namespace: flashtestation
├── deployment: go-tcb-notify
├── service: go-tcb-notify-svc
├── configmap: go-tcb-notify-config
├── secret: go-tcb-notify-secrets
└── servicemonitor: go-tcb-notify-metrics
```

## Integration with DevOps

### Alert Processing Flow
```mermaid
graph LR
    A[go-tcb-notify Alert] -->|Webhook| B[Alert Router]
    B --> C{Parse Alert}
    C -->|Extract Address| D[Automation Script]
    C -->|Extract Advisory IDs| E[Security Review]
    D --> F[invalidateAttestation()]
    E --> G[Assess Severity]
    G --> F
```

### Example Automation Handler
```javascript
async function handleTDXTCBAlert(alert) {
  const { quote } = alert;
  
  // Log advisory information
  console.log(`TDX Quote ${quote.address} degraded to ${quote.newStatus}`);
  console.log(`Security advisories: ${quote.advisoryIDs.join(', ')}`);
  
  // Auto-invalidate if critical
  if (quote.newStatus === 'Revoked' || quote.newStatus.includes('OutOfDate')) {
    const tx = await registry.invalidateAttestation(quote.address);
    await tx.wait();
    console.log(`Invalidated TDX attestation for ${quote.address}`);
  }
}
```

## Security Considerations

### Data Protection
- PostgreSQL SSL/TLS required
- No private keys or sensitive data stored
- Quote data stored for verification only

### API Security
- Intel PCS uses HTTPS only
- No authentication tokens to manage
- Rate limiting handled by retry logic

## Scalability Path

### Phase 1 (Current)
- Single instance deployment
- Sequential FMSPC checking
- Suitable for <10,000 quotes

### Phase 2 (Future)
- Parallel FMSPC processing
- Redis caching for TCB data
- Horizontal scaling with leader election

## Success Metrics

- **Detection Time**: <5 minutes from TDX TCB update
- **False Positives**: 0% (exact component matching)
- **Alert Delivery**: 99.9% success rate
- **API Availability**: Handle TDX PCS downtime gracefully

## References

- [Intel TDX PCS API v4 Documentation](https://api.portal.trustedservices.intel.com/content/documentation.html)
- [Intel TDX Architecture Specification](https://www.intel.com/content/www/us/en/developer/tools/trust-domain-extensions/documentation.html)
- [Intel Security Advisories](https://www.intel.com/content/www/us/en/security-center/default.html)
- [Flashtestation Specification](../flashtestation-spec.md)
- [go-tdx-guest Library](https://github.com/google/go-tdx-guest)
