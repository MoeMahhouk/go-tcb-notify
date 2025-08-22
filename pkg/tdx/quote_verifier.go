package tdx

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
	"github.com/google/go-tdx-guest/abi"
	pb "github.com/google/go-tdx-guest/proto/tdx"
	"github.com/google/go-tdx-guest/verify"
	"github.com/sirupsen/logrus"
)

// QuoteVerifier provides clean interface for TDX quote operations
type QuoteVerifier struct {
	getCollateral    bool
	checkRevocations bool
}

// NewQuoteVerifier creates a new quote verifier
func NewQuoteVerifier(getCollateral, checkRevocations bool) *QuoteVerifier {
	return &QuoteVerifier{
		getCollateral:    getCollateral,
		checkRevocations: checkRevocations,
	}
}

// ParseQuote parses raw quote bytes into structured format
func (v *QuoteVerifier) ParseQuote(quoteBytes []byte) (*models.ParsedQuote, error) {
	// Use abi.QuoteToProto to parse the quote
	quoteObj, err := abi.QuoteToProto(quoteBytes)
	if err != nil {
		return nil, fmt.Errorf("parse quote: %w", err)
	}

	quote, ok := quoteObj.(*pb.QuoteV4)
	if !ok {
		return nil, fmt.Errorf("unsupported quote type: %T", quoteObj)
	}

	// Use tdxutil to extract FMSPC and components
	fmspc, comps, err := ExtractFromQuote(quote)
	if err != nil {
		logrus.WithError(err).Debug("Failed to extract components from quote")
		// Continue with partial data
		return &models.ParsedQuote{
			Quote: quote,
		}, nil
	}

	parsed := &models.ParsedQuote{
		Quote: quote,
		FMSPC: fmspc,
		TCBComponents: models.TCBComponents{
			SGXComponents: comps.SGXComponents,
			TDXComponents: comps.TDXComponents,
			PCESVN:        comps.PCESVN,
		},
	}

	// Extract measurements
	if quote.TdQuoteBody != nil {
		if quote.TdQuoteBody.MrTd != nil {
			parsed.MrTd = hex.EncodeToString(quote.TdQuoteBody.MrTd)
		}
		if quote.TdQuoteBody.MrSeam != nil {
			parsed.MrSeam = hex.EncodeToString(quote.TdQuoteBody.MrSeam)
		}
		if quote.TdQuoteBody.MrSignerSeam != nil {
			parsed.MrSignerSeam = hex.EncodeToString(quote.TdQuoteBody.MrSignerSeam)
		}
		if quote.TdQuoteBody.ReportData != nil {
			parsed.ReportData = hex.EncodeToString(quote.TdQuoteBody.ReportData)
		}
	}

	return parsed, nil
}

// VerifyAndEvaluateQuote verifies a quote and evaluates its TCB status
func (v *QuoteVerifier) VerifyAndEvaluateQuote(quoteBytes []byte) (*models.QuoteEvaluation, error) {
	// Calculate hash
	hash := sha256.Sum256(quoteBytes)

	evaluation := &models.QuoteEvaluation{
		QuoteLength: uint32(len(quoteBytes)),
		QuoteHash:   hex.EncodeToString(hash[:]),
	}

	// Parse the quote first
	parsedQuote, err := v.ParseQuote(quoteBytes)
	if err != nil {
		evaluation.Status = models.StatusInvalidFormat
		evaluation.ErrorMessage = err.Error()
		return evaluation, nil
	}

	// Extract components
	evaluation.FMSPC = parsedQuote.FMSPC
	evaluation.TCBComponents = parsedQuote.TCBComponents
	evaluation.MrTd = parsedQuote.MrTd
	evaluation.MrSeam = parsedQuote.MrSeam
	evaluation.MrSignerSeam = parsedQuote.MrSignerSeam
	evaluation.ReportData = parsedQuote.ReportData

	// Verify using go-tdx-guest
	verifyOpts := verify.Options{
		GetCollateral:    v.getCollateral,
		CheckRevocations: v.checkRevocations,
	}

	// Parse the quote for verification
	quoteObj, err := abi.QuoteToProto(quoteBytes)
	if err != nil {
		evaluation.Status = models.StatusInvalidFormat
		evaluation.ErrorMessage = err.Error()
		return evaluation, nil
	}

	// Use verify.TdxQuote (takes the quote proto as interface{})
	err = verify.TdxQuote(quoteObj, &verifyOpts)
	if err != nil {
		// Parse error to determine status
		evaluation.Status, evaluation.TCBStatus = v.parseVerificationError(err)
		evaluation.ErrorMessage = err.Error()
	} else {
		evaluation.Status = models.StatusValid
		evaluation.TCBStatus = models.TCBStatusUpToDate
	}

	logrus.WithFields(logrus.Fields{
		"fmspc":     evaluation.FMSPC,
		"status":    evaluation.Status,
		"tcbStatus": evaluation.TCBStatus,
	}).Debug("Evaluated quote")

	return evaluation, nil
}

// parseVerificationError determines status from verification error
func (v *QuoteVerifier) parseVerificationError(err error) (models.QuoteStatus, models.TCBStatus) {
	if err == nil {
		return models.StatusValid, models.TCBStatusUpToDate
	}

	errStr := err.Error()

	// Check for TCB-related errors
	if strings.Contains(errStr, "OutOfDate") {
		if strings.Contains(errStr, "ConfigurationNeeded") {
			return models.StatusValidSignature, models.TCBStatusOutOfDateConfigurationNeeded
		}
		return models.StatusValidSignature, models.TCBStatusOutOfDate
	}
	if strings.Contains(errStr, "SWHardeningNeeded") {
		if strings.Contains(errStr, "Configuration") {
			return models.StatusValidSignature, models.TCBStatusConfigurationAndSWHardeningNeeded
		}
		return models.StatusValidSignature, models.TCBStatusSWHardeningNeeded
	}
	if strings.Contains(errStr, "ConfigurationNeeded") {
		return models.StatusValidSignature, models.TCBStatusConfigurationNeeded
	}
	if strings.Contains(errStr, "Revoked") {
		return models.StatusValidSignature, models.TCBStatusRevoked
	}
	if strings.Contains(errStr, "TCB") {
		return models.StatusValidSignature, models.TCBStatusUnknown
	}

	// Signature or structure errors
	if strings.Contains(errStr, "signature") || strings.Contains(errStr, "certificate") {
		return models.StatusInvalidSignature, models.TCBStatusNotApplicable
	}

	return models.StatusInvalid, models.TCBStatusNotApplicable
}

// ExtractComponentDetails provides detailed component information using centralized names
func ExtractComponentDetails(components models.TCBComponents) []models.ComponentDetail {
	details := make([]models.ComponentDetail, 0, 32)

	// SGX components (positions 0-15)
	for i := 0; i < 16; i++ {
		details = append(details, models.ComponentDetail{
			Index:   i,
			Name:    GetComponentName(i),
			Version: int(components.SGXComponents[i]),
			Type:    "SGX",
		})
	}

	// TDX components (positions 16-31)
	for i := 0; i < 16; i++ {
		details = append(details, models.ComponentDetail{
			Index:   16 + i,
			Name:    GetComponentName(16 + i),
			Version: int(components.TDXComponents[i]),
			Type:    "TDX",
		})
	}

	return details
}

// ExtractFromQuote gathers (FMSPC, SGX[16], PCESVN) from PCK and TDX[16] from TD report.
func ExtractFromQuote(q *pb.QuoteV4) (fmspcHex string, comps models.TCBComponents, err error) {
	fmspcHex, err = ExtractFMSPCFromQuote(q)
	if err != nil {
		return "", comps, err
	}
	sgx16, pcesvn, err := ExtractSgxTcbAndPceSvnFromQuote(q)
	if err != nil {
		return "", comps, err
	}
	comps.SGXComponents = sgx16
	comps.PCESVN = pcesvn

	tdxBytes := q.GetTdQuoteBody().GetTeeTcbSvn() // 16 bytes
	for i := 0; i < 16 && i < len(tdxBytes); i++ {
		comps.TDXComponents[i] = tdxBytes[i]
	}
	return fmspcHex, comps, nil
}
