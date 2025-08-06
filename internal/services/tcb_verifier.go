package services

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
	"github.com/sirupsen/logrus"
)

type TCBVerifier struct {
	db *sql.DB
}

type TCBVerificationResult struct {
	Status      string   `json:"status"`
	Reason      string   `json:"reason"`
	AdvisoryIDs []string `json:"advisoryIds"`
	Level       int      `json:"level"`
}

func NewTCBVerifier(db *sql.DB) *TCBVerifier {
	return &TCBVerifier{db: db}
}

func (v *TCBVerifier) VerifyQuote(quoteTCB *QuoteTCBComponents, tcbInfo *models.TCBInfo) (*TCBVerificationResult, error) {
	// Parse TCB levels from stored JSON
	var tcbLevels []models.TCBLevel
	if err := json.Unmarshal(tcbInfo.TCBLevels, &tcbLevels); err != nil {
		return nil, fmt.Errorf("failed to parse TCB levels: %w", err)
	}

	// Find the highest TCB level that the quote meets
	for i, level := range tcbLevels {
		if v.meetsLevel(quoteTCB, &level) {
			return &TCBVerificationResult{
				Status:      level.TCBStatus,
				Reason:      v.getStatusReason(level.TCBStatus),
				AdvisoryIDs: level.AdvisoryIDs,
				Level:       i,
			}, nil
		}
	}

	// If no level is met, the quote is completely out of date
	return &TCBVerificationResult{
		Status:      "OutOfDate",
		Reason:      "Quote does not meet any current TCB level",
		AdvisoryIDs: []string{},
		Level:       -1,
	}, nil
}

func (v *TCBVerifier) meetsLevel(quoteTCB *QuoteTCBComponents, level *models.TCBLevel) bool {
	// Check SGX TCB components (underlying platform)
	for i := 0; i < 16 && i < len(quoteTCB.SGXTCBComponents) && i < len(level.TCB.SGXTCBComponents); i++ {
		if quoteTCB.SGXTCBComponents[i] < level.TCB.SGXTCBComponents[i].SVN {
			logrus.WithFields(logrus.Fields{
				"component": i,
				"quote_svn": quoteTCB.SGXTCBComponents[i],
				"level_svn": level.TCB.SGXTCBComponents[i].SVN,
			}).Debug("SGX TCB component below required level")
			return false
		}
	}

	// Check TDX-specific components
	for i := 0; i < 16 && i < len(quoteTCB.TDXTCBComponents) && i < len(level.TCB.TDXTCBComponents); i++ {
		if quoteTCB.TDXTCBComponents[i] < level.TCB.TDXTCBComponents[i].SVN {
			logrus.WithFields(logrus.Fields{
				"component": i,
				"quote_svn": quoteTCB.TDXTCBComponents[i],
				"level_svn": level.TCB.TDXTCBComponents[i].SVN,
			}).Debug("TDX TCB component below required level")
			return false
		}
	}

	// Check PCE SVN
	if quoteTCB.PCESVN < level.TCB.PCESVN {
		logrus.WithFields(logrus.Fields{
			"quote_pcesvn": quoteTCB.PCESVN,
			"level_pcesvn": level.TCB.PCESVN,
		}).Debug("PCE SVN below required level")
		return false
	}

	return true
}

func (v *TCBVerifier) getStatusReason(status string) string {
	switch status {
	case "UpToDate":
		return "Quote meets current TCB requirements"
	case "OutOfDate":
		return "Quote TCB components are below current requirements"
	case "ConfigurationNeeded":
		return "Platform configuration update required"
	case "Revoked":
		return "TCB level has been revoked due to security issues"
	default:
		return "Unknown TCB status"
	}
}
