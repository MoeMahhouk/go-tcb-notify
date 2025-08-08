// internal/services/tcb_verifier.go
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

	logrus.WithFields(logrus.Fields{
		"fmspc":       tcbInfo.FMSPC,
		"numLevels":   len(tcbLevels),
		"quotePCESVN": quoteTCB.PCESVN,
	}).Debug("Starting TCB verification")

	// Find the highest TCB level that the quote meets
	// TCB levels are ordered from highest (most secure) to lowest
	for i, level := range tcbLevels {
		if v.meetsLevel(quoteTCB, &level) {
			logrus.WithFields(logrus.Fields{
				"level":       i,
				"status":      level.TCBStatus,
				"advisoryIDs": level.AdvisoryIDs,
			}).Info("Quote meets TCB level")

			return &TCBVerificationResult{
				Status:      level.TCBStatus,
				Reason:      v.getStatusReason(level.TCBStatus),
				AdvisoryIDs: level.AdvisoryIDs,
				Level:       i,
			}, nil
		}
	}

	// If no level is met, the quote is completely out of date
	logrus.WithField("fmspc", tcbInfo.FMSPC).Warn("Quote does not meet any TCB level")

	return &TCBVerificationResult{
		Status:      "OutOfDate",
		Reason:      "Quote does not meet any current TCB level",
		AdvisoryIDs: []string{},
		Level:       -1,
	}, nil
}

func (v *TCBVerifier) meetsLevel(quoteTCB *QuoteTCBComponents, level *models.TCBLevel) bool {
	// Check SGX TCB components (underlying platform)
	// All quote components must be >= the level's required components
	for i := 0; i < 16; i++ {
		if i < len(quoteTCB.SGXTCBComponents) && i < len(level.TCB.SGXTCBComponents) {
			quoteValue := uint8(quoteTCB.SGXTCBComponents[i])
			requiredValue := uint8(level.TCB.SGXTCBComponents[i].SVN)

			if quoteValue < requiredValue {
				logrus.WithFields(logrus.Fields{
					"component":     i,
					"type":          "SGX",
					"quoteValue":    quoteValue,
					"requiredValue": requiredValue,
				}).Debug("Component does not meet required level")
				return false
			}
		}
	}

	// Check TDX-specific components
	for i := 0; i < 16; i++ {
		if i < len(quoteTCB.TDXTCBComponents) && i < len(level.TCB.TDXTCBComponents) {
			quoteValue := uint8(quoteTCB.TDXTCBComponents[i])
			requiredValue := uint8(level.TCB.TDXTCBComponents[i].SVN)

			if quoteValue < requiredValue {
				logrus.WithFields(logrus.Fields{
					"component":     i,
					"type":          "TDX",
					"quoteValue":    quoteValue,
					"requiredValue": requiredValue,
				}).Debug("Component does not meet required level")
				return false
			}
		}
	}

	// Check PCE SVN
	if quoteTCB.PCESVN < level.TCB.PCESVN {
		logrus.WithFields(logrus.Fields{
			"quotePCESVN":    quoteTCB.PCESVN,
			"requiredPCESVN": level.TCB.PCESVN,
		}).Debug("PCE SVN does not meet required level")
		return false
	}

	logrus.Debug("Quote meets all requirements for this TCB level")
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
	case "ConfigurationAndSWHardeningNeeded":
		return "Platform configuration and software hardening needed"
	case "SWHardeningNeeded":
		return "Software hardening needed for TCB recovery"
	case "Revoked":
		return "TCB level has been revoked due to security issues"
	default:
		return fmt.Sprintf("TCB status: %s", status)
	}
}
