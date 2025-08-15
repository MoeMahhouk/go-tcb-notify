package models

import (
	"encoding/json"
	"time"

	pb "github.com/google/go-tdx-guest/proto/tdx"
)

// Quote Status types
type QuoteStatus string

const (
	StatusValid            QuoteStatus = "VALID"
	StatusValidSignature   QuoteStatus = "VALID_SIGNATURE"
	StatusInvalid          QuoteStatus = "INVALID"
	StatusInvalidSignature QuoteStatus = "INVALID_SIGNATURE"
	StatusInvalidFormat    QuoteStatus = "INVALID_FORMAT"
)

// TCB Status types
type TCBStatus string

const (
	TCBStatusUpToDate                          TCBStatus = "UpToDate"
	TCBStatusSWHardeningNeeded                 TCBStatus = "SWHardeningNeeded"
	TCBStatusConfigurationNeeded               TCBStatus = "ConfigurationNeeded"
	TCBStatusConfigurationAndSWHardeningNeeded TCBStatus = "ConfigurationAndSWHardeningNeeded"
	TCBStatusOutOfDate                         TCBStatus = "OutOfDate"
	TCBStatusOutOfDateConfigurationNeeded      TCBStatus = "OutOfDateConfigurationNeeded"
	TCBStatusRevoked                           TCBStatus = "Revoked"
	TCBStatusUnknown                           TCBStatus = "Unknown"
	TCBStatusNotApplicable                     TCBStatus = "N/A"
)

// RegistryEvent represents an event from the FlashtestationRegistry
type RegistryEvent struct {
	EventType     string    `db:"event_type"`
	TeeAddress    string    `db:"tee_address"`
	QuoteBytes    []byte    `db:"quote_bytes"`
	QuoteHash     string    `db:"quote_hash"`
	QuoteLength   uint32    `db:"quote_length"`
	AlreadyExists bool      `db:"already_exists"`
	BlockNumber   uint64    `db:"block_number"`
	LogIndex      uint32    `db:"log_index"`
	TxHash        string    `db:"tx_hash"`
	BlockTime     time.Time `db:"block_time"`
	IngestedAt    time.Time `db:"ingested_at"`
}

// ParsedQuote contains parsed TDX quote data
type ParsedQuote struct {
	Quote         *pb.QuoteV4
	FMSPC         string
	TCBComponents TCBComponents
	MrTd          string
	MrSeam        string
	MrSignerSeam  string
	ReportData    string
}

// TCBComponents contains the TCB component values
type TCBComponents struct {
	SGXComponents [16]byte `json:"sgx_components"`
	TDXComponents [16]byte `json:"tdx_components"`
	PCESVN        uint16   `json:"pce_svn"`
}

// QuoteEvaluation contains the complete evaluation result
type QuoteEvaluation struct {
	ServiceAddress string        `json:"service_address" db:"service_address"`
	QuoteHash      string        `json:"quote_hash" db:"quote_hash"`
	QuoteLength    int           `json:"quote_length" db:"quote_length"`
	FMSPC          string        `json:"fmspc" db:"fmspc"`
	TCBComponents  TCBComponents `json:"tcb_components" db:"tcb_components"`
	MrTd           string        `json:"mr_td" db:"mr_td"`
	MrSeam         string        `json:"mr_seam" db:"mr_seam"`
	MrSignerSeam   string        `json:"mr_signer_seam" db:"mr_signer_seam"`
	ReportData     string        `json:"report_data" db:"report_data"`
	Status         QuoteStatus   `json:"status" db:"status"`
	TCBStatus      TCBStatus     `json:"tcb_status" db:"tcb_status"`
	Error          string        `json:"error,omitempty" db:"error"`
	BlockNumber    uint64        `json:"block_number" db:"block_number"`
	LogIndex       uint32        `json:"log_index" db:"log_index"`
	BlockTime      time.Time     `json:"block_time" db:"block_time"`
	EvaluatedAt    time.Time     `json:"evaluated_at" db:"evaluated_at"`
}

// ComponentDetail provides detailed information about a TCB component
type ComponentDetail struct {
	Index   int    `json:"index"`
	Name    string `json:"name"`
	Version int    `json:"version"`
	Type    string `json:"type"` // "SGX" or "TDX"
}

// TCBInfo represents TCB information from Intel PCS
type TCBInfo struct {
	ID                      string      `json:"id" db:"id"`
	Version                 int         `json:"version" db:"version"`
	IssueDate               time.Time   `json:"issueDate" db:"issue_date"`
	NextUpdate              time.Time   `json:"nextUpdate" db:"next_update"`
	FMSPC                   string      `json:"fmspc" db:"fmspc"`
	PCEId                   string      `json:"pceId" db:"pce_id"`
	TCBType                 int         `json:"tcbType" db:"tcb_type"`
	TCBEvaluationDataNumber uint32      `json:"tcbEvaluationDataNumber" db:"tcb_evaluation_data_number"`
	TDXModule               interface{} `json:"tdxModule,omitempty" db:"tdx_module"`
	TcbLevels               []TCBLevel  `json:"tcbLevels" db:"tcb_levels"`
	Status                  TCBStatus   `json:"status,omitempty"`
	AdvisoryIDs             []string    `json:"advisoryIDs,omitempty"`
	RawJSON                 string      `json:"-" db:"raw_json"`
	FetchedAt               time.Time   `json:"fetchedAt" db:"fetched_at"`
}

// TCBLevel represents a single TCB level from Intel PCS
type TCBLevel struct {
	TCB struct {
		SGXComponents []ComponentSVN `json:"sgxtcbcomponents"`
		TDXComponents []ComponentSVN `json:"tdxtcbcomponents,omitempty"`
		PCESVN        int            `json:"pcesvn"`
	} `json:"tcb"`
	TCBDate     string    `json:"tcbDate"`
	TCBStatus   TCBStatus `json:"tcbStatus"`
	AdvisoryIDs []string  `json:"advisoryIDs,omitempty"`
}

// ComponentSVN represents a component's security version
type ComponentSVN struct {
	SVN      int    `json:"svn"`
	Category string `json:"category,omitempty"`
	Type     string `json:"type,omitempty"`
}

// TCBChange represents a change in TCB component versions
type TCBChange struct {
	FMSPC            string    `json:"fmspc" db:"fmspc"`
	ComponentIndex   int       `json:"componentIndex" db:"component_index"`
	ComponentName    string    `json:"componentName" db:"component_name"`
	ComponentType    string    `json:"componentType" db:"component_type"`
	OldVersion       int       `json:"oldVersion" db:"old_version"`
	NewVersion       int       `json:"newVersion" db:"new_version"`
	ChangeType       string    `json:"changeType" db:"change_type"` // "UPGRADE" or "DOWNGRADE"
	EvaluationNumber uint32    `json:"evaluationNumber" db:"evaluation_number"`
	DetectedAt       time.Time `json:"detectedAt" db:"detected_at"`
}

// FMSPC represents FMSPC information from Intel PCS
type FMSPC struct {
	FMSPC     string    `json:"fmspc" db:"fmspc"`
	Platform  string    `json:"platform" db:"platform"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// FMSPCResponse for PCS API response
type FMSPCResponse struct {
	FMSPC    string `json:"fmspc"`
	Platform string `json:"platform"`
}

// MonitoredQuote represents a quote being monitored for changes
type MonitoredQuote struct {
	Address       string          `json:"address" db:"address"`
	QuoteData     []byte          `json:"quoteData" db:"quote_data"`
	WorkloadID    string          `json:"workloadId" db:"workload_id"`
	FMSPC         string          `json:"fmspc" db:"fmspc"`
	TCBComponents json.RawMessage `json:"tcbComponents" db:"tcb_components"`
	CurrentStatus string          `json:"currentStatus" db:"current_status"`
	NeedsUpdate   bool            `json:"needsUpdate" db:"needs_update"`
	LastChecked   time.Time       `json:"lastChecked" db:"last_checked"`
}

// QuoteAlert represents an alert for a quote status change
type QuoteAlert struct {
	Address                 string   `json:"address"`
	Reason                  string   `json:"reason"`
	PreviousStatus          string   `json:"previousStatus"`
	NewStatus               string   `json:"newStatus"`
	WorkloadID              string   `json:"workloadId"`
	TCBEvaluationDataNumber int      `json:"tcbEvaluationDataNumber"`
	AdvisoryIDs             []string `json:"advisoryIDs"`
	FMSPC                   string   `json:"fmspc"`
	SuggestedAction         string   `json:"suggestedAction"`
}

// AlertHistory tracks historical alerts
type AlertHistory struct {
	ID           int             `json:"id" db:"id"`
	QuoteAddress string          `json:"quoteAddress" db:"quote_address"`
	Reason       string          `json:"reason" db:"reason"`
	Details      json.RawMessage `json:"details" db:"details"`
	SentAt       time.Time       `json:"sentAt" db:"sent_at"`
	Acknowledged bool            `json:"acknowledged" db:"acknowledged"`
}
