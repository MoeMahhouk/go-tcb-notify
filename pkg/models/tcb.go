package models

import (
	"encoding/json"
	"time"
)

type TCBInfo struct {
	FMSPC                   string          `json:"fmspc" db:"fmspc"`
	Version                 int             `json:"version" db:"version"`
	IssueDate               time.Time       `json:"issueDate" db:"issue_date"`
	NextUpdate              time.Time       `json:"nextUpdate" db:"next_update"`
	TCBType                 int             `json:"tcbType" db:"tcb_type"`
	TCBEvaluationDataNumber int             `json:"tcbEvaluationDataNumber" db:"tcb_evaluation_data_number"`
	TCBLevels               json.RawMessage `json:"tcbLevels" db:"tcb_levels"`
	RawResponse             json.RawMessage `json:"rawResponse" db:"raw_response"`
	CreatedAt               time.Time       `json:"createdAt" db:"created_at"`
}

type TCBLevel struct {
	TCB struct {
		SGXTCBComponents []TCBComponent `json:"sgxtcbcomponents"`
		PCESVN           int            `json:"pcesvn"`
		TDXTCBComponents []TCBComponent `json:"tdxtcbcomponents"`
	} `json:"tcb"`
	TCBDate     time.Time `json:"tcbDate"`
	TCBStatus   string    `json:"tcbStatus"`
	AdvisoryIDs []string  `json:"advisoryIDs"`
}

type TCBComponent struct {
	SVN      int    `json:"svn"`
	Category string `json:"category,omitempty"`
	Type     string `json:"type,omitempty"`
}
