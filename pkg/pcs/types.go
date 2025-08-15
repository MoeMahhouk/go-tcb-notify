package pcs

import (
	"time"

	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
)

// TCBInfoResponse represents the full response from Intel PCS TCB API
type TCBInfoResponse struct {
	TcbInfo   models.TCBInfo `json:"tcbInfo"`
	Signature string         `json:"signature"`
}

// TCBInfoData represents the parsed TCB info from the raw response
type TCBInfoData struct {
	NextUpdate time.Time `json:"nextUpdate"`
	IssueDate  time.Time `json:"issueDate"`
}

// TCBChangeAlert represents data for creating TCB change alerts
type TCBChangeAlert struct {
	FMSPC            string
	EvaluationNumber uint32
	TotalChanges     uint32
	Downgrades       uint32
	Severity         string
}
