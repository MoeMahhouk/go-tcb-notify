package models

import (
	"time"
)

// RegistryQuote represents a quote from the registry
type RegistryQuote struct {
	ServiceAddress string    `db:"service_address"`
	BlockNumber    uint64    `db:"block_number"`
	BlockTime      time.Time `db:"block_time"`
	TxHash         string    `db:"tx_hash"`
	LogIndex       uint32    `db:"log_index"`
	QuoteBytes     []byte    `db:"quote_bytes"`
	QuoteLength    uint32    `db:"quote_length"`
	QuoteHash      string    `db:"quote_hash"`
	FMSPC          string    `db:"fmspc"`
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
