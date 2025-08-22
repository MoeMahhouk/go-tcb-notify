package models

import (
	"encoding/json"
	"time"

	pb "github.com/google/go-tdx-guest/proto/tdx"
)

// ===================
// Core Types
// ===================

// QuoteStatus represents the validation status of a quote
type QuoteStatus string

const (
	StatusValid            QuoteStatus = "VALID"
	StatusValidSignature   QuoteStatus = "VALID_SIGNATURE"
	StatusInvalid          QuoteStatus = "INVALID"
	StatusInvalidSignature QuoteStatus = "INVALID_SIGNATURE"
	StatusInvalidFormat    QuoteStatus = "INVALID_FORMAT"
)

// TCBStatus represents the status of TCB components
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

// TCBComponents contains the TCB component values
type TCBComponents struct {
	SGXComponents [16]byte `json:"sgx_components"`
	TDXComponents [16]byte `json:"tdx_components"`
	PCESVN        uint16   `json:"pce_svn"`
}

// ComponentDetail represents detailed information about a TCB component
type ComponentDetail struct {
	Index   int    `json:"index"`
	Name    string `json:"name"`
	Version int    `json:"version"`
	Type    string `json:"type"` // "SGX" or "TDX"
}

// ===================
// Registry/Blockchain Models
// ===================

// RegistryQuote represents a quote from the blockchain registry
// Maps to: registry_quotes_raw table
type RegistryQuote struct {
	ServiceAddress string    `db:"service_address"`
	BlockNumber    uint64    `db:"block_number"`
	BlockTime      time.Time `db:"block_time"`
	TxHash         string    `db:"tx_hash"`
	LogIndex       uint32    `db:"log_index"`
	QuoteBytes     []byte    `db:"quote_bytes"`
	QuoteLength    uint32    `db:"quote_len"`    // Note: DB field is quote_len
	QuoteHash      string    `db:"quote_sha256"` // Note: DB field is quote_sha256
	FMSPC          string    `db:"fmspc"`
	IngestedAt     time.Time `db:"ingested_at"`
}

// ===================
// Quote Evaluation Models
// ===================

// QuoteEvaluation represents the evaluation result of a quote
// Maps to: tdx_quote_evaluations table
type QuoteEvaluation struct {
	ServiceAddress string        `json:"service_address" db:"service_address"`
	QuoteHash      string        `json:"quote_hash" db:"quote_hash"`
	QuoteLength    uint32        `json:"quote_length" db:"quote_length"`
	Status         QuoteStatus   `json:"status" db:"status"`
	TCBStatus      TCBStatus     `json:"tcb_status" db:"tcb_status"`
	ErrorMessage   string        `json:"error_message" db:"error_message"` // Alternative error field
	FMSPC          string        `json:"fmspc" db:"fmspc"`
	TCBComponents  TCBComponents `json:"tcb_components" db:"-"` // Not stored directly in DB
	PCESVN         uint16        `json:"pce_svn" db:"pce_svn"`
	BlockNumber    uint64        `json:"block_number" db:"block_number"`
	LogIndex       uint32        `json:"log_index" db:"log_index"`
	BlockTime      time.Time     `json:"block_time" db:"block_time"`
	EvaluatedAt    time.Time     `json:"evaluated_at" db:"evaluated_at"`
	// Additional fields from ParsedQuote
	MrTd         string `json:"mr_td" db:"mr_td"`
	MrSeam       string `json:"mr_seam" db:"mr_seam"`
	MrSignerSeam string `json:"mr_signer_seam" db:"mr_signer_seam"`
	ReportData   string `json:"report_data" db:"report_data"`
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

// ===================
// Intel PCS Models
// ===================

// FMSPC represents FMSPC information from Intel PCS
// Maps to: pcs_fmspcs table
type FMSPC struct {
	FMSPC     string    `json:"fmspc" db:"fmspc"`
	Platform  string    `json:"platform" db:"platform"`
	FirstSeen time.Time `json:"first_seen" db:"first_seen"`
	LastSeen  time.Time `json:"last_seen" db:"last_seen"`
	IsActive  bool      `json:"is_active" db:"is_active"`
}

// TCBInfo represents TCB information from Intel PCS
// Maps to: pcs_tcb_info table
type TCBInfo struct {
	FMSPC                   string    `json:"fmspc" db:"fmspc"`
	TCBEvaluationDataNumber uint32    `json:"tcbEvaluationDataNumber" db:"tcb_evaluation_data_number"`
	IssueDate               time.Time `json:"issueDate" db:"issue_date"`
	NextUpdate              time.Time `json:"nextUpdate" db:"next_update"`
	TCBType                 uint32    `json:"tcbType" db:"tcb_type"`
	TCBLevelsJSON           string    `json:"-" db:"tcb_levels_json"` // JSON in DB
	RawJSON                 string    `json:"-" db:"raw_json"`        // Complete response
	FetchedAt               time.Time `json:"fetchedAt" db:"fetched_at"`

	// Parsed fields (not stored in DB directly)
	TCBLevels []TCBLevel `json:"tcbLevels"`
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

// TCBChangeAlert represents an alert for TCB changes
// Maps to: tcb_change_alerts table
type TCBChangeAlert struct {
	ID                  string    `json:"id" db:"id"`
	FMSPC               string    `json:"fmspc" db:"fmspc"`
	OldEvalNumber       uint32    `json:"old_eval_number" db:"old_eval_number"`
	NewEvalNumber       uint32    `json:"new_eval_number" db:"new_eval_number"`
	ChangeType          string    `json:"change_type" db:"change_type"` // Minor, Major, Critical
	AffectedQuotesCount uint32    `json:"affected_quotes_count" db:"affected_quotes_count"`
	Details             string    `json:"details" db:"details"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	Acknowledged        bool      `json:"acknowledged" db:"acknowledged"`
}

// ===================
// Pipeline Management
// ===================

// PipelineOffset tracks the last processed position for each service
// Maps to: pipeline_offsets table
type PipelineOffset struct {
	Service      string    `db:"service"`
	LastBlock    uint64    `db:"last_block"`
	LastLogIndex uint32    `db:"last_log_index"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// ===================
// API Response Types
// ===================

// FMSPCResponse represents the response from Intel PCS FMSPC API
type FMSPCResponse struct {
	FMSPC    string `json:"fmspc"`
	Platform string `json:"platform"`
}

// TCBInfoResponse represents the full response from Intel PCS TCB API
type TCBInfoResponse struct {
	TCBInfo   json.RawMessage `json:"tcbInfo"`
	Signature string          `json:"signature"`
}
