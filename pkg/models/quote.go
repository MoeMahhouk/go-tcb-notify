package models

import (
	"encoding/json"
	"time"
)

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

type Alert struct {
	Severity  string     `json:"severity"`
	Source    string     `json:"source"`
	Timestamp time.Time  `json:"timestamp"`
	Quote     QuoteAlert `json:"quote"`
}

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

type AlertHistory struct {
	ID           int             `json:"id" db:"id"`
	QuoteAddress string          `json:"quoteAddress" db:"quote_address"`
	Reason       string          `json:"reason" db:"reason"`
	Details      json.RawMessage `json:"details" db:"details"`
	SentAt       time.Time       `json:"sentAt" db:"sent_at"`
	Acknowledged bool            `json:"acknowledged" db:"acknowledged"`
}