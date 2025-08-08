package tdx

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/google/go-tdx-guest/abi"
	"github.com/google/go-tdx-guest/proto/tdx"
	"github.com/google/go-tdx-guest/validate"
	"github.com/google/go-tdx-guest/verify"
	"github.com/sirupsen/logrus"
)

// QuoteParser provides TDX quote parsing functionality
type QuoteParser struct {
	// We can add configuration options here if needed
}

// NewQuoteParser creates a new QuoteParser instance
func NewQuoteParser() *QuoteParser {
	return &QuoteParser{}
}

// ParsedQuote represents a parsed TDX quote with extracted components
type ParsedQuote struct {
	Quote         *tdx.QuoteV4
	FMSPC         string
	TCBComponents TCBComponents
}

// TCBComponents represents the TCB components needed for verification
type TCBComponents struct {
	SGXComponents []uint8 // CPU SVN components (16 components)
	TDXComponents []uint8 // TDX-specific components (16 components)
	PCESVN        uint16
}

// ParseQuote parses a raw TDX quote and extracts relevant components
func (p *QuoteParser) ParseQuote(quoteData []byte) (*ParsedQuote, error) {
	if len(quoteData) == 0 {
		return nil, fmt.Errorf("quote data is empty")
	}

	logrus.WithField("size", len(quoteData)).Debug("Parsing quote")

	// First try to validate the raw quote structure
	// This might give us hints about what's wrong
	opts := &validate.Options{}
	if err := validate.RawTdxQuote(quoteData, opts); err != nil {
		logrus.WithError(err).Warn("Quote validation warning")
		// Don't fail here, continue trying to parse
	}

	// Try different parsing methods
	var quote *tdx.QuoteV4
	var err error

	// Method 1: Try abi.QuoteToProto
	quoteInterface, err := abi.QuoteToProto(quoteData)
	if err == nil {
		if q, ok := quoteInterface.(*tdx.QuoteV4); ok {
			quote = q
			logrus.Debug("Successfully parsed with abi.QuoteToProto")
		} else {
			logrus.Warn("abi.QuoteToProto returned non-QuoteV4 type")
		}
	} else {
		logrus.WithError(err).Debug("abi.QuoteToProto failed, trying manual parsing")

		// Method 2: Try manual parsing
		quote, err = p.parseQuoteManually(quoteData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse quote: %w", err)
		}
		logrus.Debug("Successfully parsed with manual parser")
	}

	// Extract components
	fmspc := p.extractFMSPC(quote, quoteData)
	tcbComponents := p.extractTCBComponents(quote, quoteData)

	return &ParsedQuote{
		Quote:         quote,
		FMSPC:         fmspc,
		TCBComponents: tcbComponents,
	}, nil
}

// parseQuoteManually parses the quote structure manually
func (p *QuoteParser) parseQuoteManually(data []byte) (*tdx.QuoteV4, error) {
	if len(data) < 48 {
		return nil, fmt.Errorf("quote too small: %d bytes (minimum 48)", len(data))
	}

	// Check the header to understand the format
	version := binary.LittleEndian.Uint16(data[0:2])
	attestKeyType := binary.LittleEndian.Uint16(data[2:4])
	teeType := binary.LittleEndian.Uint32(data[4:8])

	logrus.WithFields(logrus.Fields{
		"version":       version,
		"attestKeyType": attestKeyType,
		"teeType":       fmt.Sprintf("0x%x", teeType),
	}).Debug("Quote header info")

	// Check if this is a TDX quote
	if teeType != 0x00000081 {
		logrus.WithField("teeType", fmt.Sprintf("0x%x", teeType)).Warn("Not a TDX quote (expected 0x81)")
	}

	// Check version
	if version != 4 {
		logrus.WithField("version", version).Warn("Not a version 4 quote")
		// Continue anyway as the structure might still be parseable
	}

	quote := &tdx.QuoteV4{
		Header:      &tdx.Header{},
		TdQuoteBody: &tdx.TDQuoteBody{},
	}

	// Parse header (48 bytes)
	quote.Header.Version = uint32(version)
	quote.Header.AttestationKeyType = uint32(attestKeyType)
	quote.Header.TeeType = teeType

	// QE SVN and PCE SVN are in bytes 8-12
	quote.Header.QeSvn = data[8:10]
	quote.Header.PceSvn = data[10:12]

	// QE Vendor ID (16 bytes at offset 12)
	quote.Header.QeVendorId = data[12:28]

	// User Data (20 bytes at offset 28)
	quote.Header.UserData = data[28:48]

	// Check if we have TD Quote Body (584 bytes)
	if len(data) < 48+584 {
		return nil, fmt.Errorf("quote too small for TD Quote Body: %d bytes (need %d)", len(data), 48+584)
	}

	// Parse TD Quote Body
	bodyStart := 48
	bodyEnd := bodyStart + 584
	tdBody := data[bodyStart:bodyEnd]

	offset := 0

	// TEE TCB SVN (16 bytes)
	quote.TdQuoteBody.TeeTcbSvn = tdBody[offset : offset+16]
	offset += 16

	// MR SEAM (48 bytes)
	quote.TdQuoteBody.MrSeam = tdBody[offset : offset+48]
	offset += 48

	// MR SIGNER SEAM (48 bytes)
	quote.TdQuoteBody.MrSignerSeam = tdBody[offset : offset+48]
	offset += 48

	// SEAM Attributes (8 bytes)
	quote.TdQuoteBody.SeamAttributes = tdBody[offset : offset+8]
	offset += 8

	// TD Attributes (8 bytes)
	quote.TdQuoteBody.TdAttributes = tdBody[offset : offset+8]
	offset += 8

	// XFAM (8 bytes)
	quote.TdQuoteBody.Xfam = tdBody[offset : offset+8]
	offset += 8

	// MR TD (48 bytes)
	quote.TdQuoteBody.MrTd = tdBody[offset : offset+48]
	offset += 48

	// MR CONFIG ID (48 bytes)
	quote.TdQuoteBody.MrConfigId = tdBody[offset : offset+48]
	offset += 48

	// MR OWNER (48 bytes)
	quote.TdQuoteBody.MrOwner = tdBody[offset : offset+48]
	offset += 48

	// MR OWNER CONFIG (48 bytes)
	quote.TdQuoteBody.MrOwnerConfig = tdBody[offset : offset+48]
	offset += 48

	// RTMRs (4 x 48 bytes)
	quote.TdQuoteBody.Rtmrs = make([][]byte, 4)
	for i := 0; i < 4; i++ {
		quote.TdQuoteBody.Rtmrs[i] = tdBody[offset : offset+48]
		offset += 48
	}

	// Report Data (64 bytes)
	quote.TdQuoteBody.ReportData = tdBody[offset : offset+64]

	// Parse signature data if present
	if len(data) > bodyEnd {
		// Signature data size (4 bytes)
		if len(data) >= bodyEnd+4 {
			sigDataSize := binary.LittleEndian.Uint32(data[bodyEnd : bodyEnd+4])
			quote.SignedDataSize = sigDataSize

			logrus.WithField("signatureSize", sigDataSize).Debug("Quote has signature data")

			// Parse signature data if we have enough bytes
			if len(data) >= bodyEnd+4+int(sigDataSize) {
				sigDataStart := bodyEnd + 4
				sigDataEnd := sigDataStart + int(sigDataSize)

				// For now, just store the raw signature data
				// A full implementation would parse the ECDSA signature structure
				quote.SignedData = &tdx.Ecdsa256BitQuoteV4AuthData{
					Signature: data[sigDataStart:min(sigDataEnd, len(data))],
				}
			}
		}
	}

	return quote, nil
}

// extractFMSPC extracts the FMSPC from the parsed quote or raw data
func (p *QuoteParser) extractFMSPC(quote *tdx.QuoteV4, rawData []byte) string {
	// Method 1: Try to extract from parsed quote
	if quote != nil && quote.TdQuoteBody != nil && len(quote.TdQuoteBody.TeeTcbSvn) >= 6 {
		fmspc := hex.EncodeToString(quote.TdQuoteBody.TeeTcbSvn[0:6])
		logrus.WithField("fmspc", fmspc).Debug("Extracted FMSPC from parsed TEE TCB SVN")
		return fmspc
	}

	// Method 2: Extract directly from raw data
	// TEE TCB SVN starts at offset 48 (after header) in the TD Quote Body
	if len(rawData) >= 48+16 {
		fmspc := hex.EncodeToString(rawData[48:54])
		logrus.WithField("fmspc", fmspc).Debug("Extracted FMSPC from raw data")
		return fmspc
	}

	logrus.Warn("Could not extract FMSPC")
	return ""
}

// extractTCBComponents extracts TCB components from the quote
func (p *QuoteParser) extractTCBComponents(quote *tdx.QuoteV4, rawData []byte) TCBComponents {
	components := TCBComponents{
		SGXComponents: make([]uint8, 16),
		TDXComponents: make([]uint8, 16),
		PCESVN:        0,
	}

	// Try to extract from parsed quote first
	if quote != nil && quote.TdQuoteBody != nil && quote.TdQuoteBody.TeeTcbSvn != nil {
		// TEE TCB SVN contains the SGX TCB components
		for i := 0; i < 16 && i < len(quote.TdQuoteBody.TeeTcbSvn); i++ {
			components.SGXComponents[i] = quote.TdQuoteBody.TeeTcbSvn[i]
		}

		// Extract PCE SVN from header
		if quote.Header != nil && len(quote.Header.PceSvn) >= 2 {
			components.PCESVN = binary.LittleEndian.Uint16(quote.Header.PceSvn)
		}
	} else if len(rawData) >= 48+16 {
		// Fallback: extract from raw data
		// TEE TCB SVN is at offset 48
		for i := 0; i < 16; i++ {
			components.SGXComponents[i] = rawData[48+i]
		}

		// PCE SVN is at offset 10-12 in header
		if len(rawData) >= 12 {
			components.PCESVN = binary.LittleEndian.Uint16(rawData[10:12])
		}
	}

	logrus.WithFields(logrus.Fields{
		"sgxComponents": hex.EncodeToString(components.SGXComponents),
		"pcesvn":        components.PCESVN,
	}).Debug("Extracted TCB components")

	return components
}

// VerifyQuoteSignature verifies the signature of a TDX quote
func (p *QuoteParser) VerifyQuoteSignature(quoteData []byte, getCollateral bool) error {
	// Use the go-tdx-guest verify package
	opts := verify.Options{
		GetCollateral:    getCollateral,
		CheckRevocations: false,
	}

	// Verify the quote
	if err := verify.TdxQuote(quoteData, &opts); err != nil {
		return fmt.Errorf("quote signature verification failed: %w", err)
	}

	return nil
}

// GetQuoteInfo returns basic information about the quote
func (p *QuoteParser) GetQuoteInfo(parsedQuote *ParsedQuote) map[string]interface{} {
	info := make(map[string]interface{})

	if parsedQuote.Quote != nil && parsedQuote.Quote.Header != nil {
		info["version"] = parsedQuote.Quote.Header.Version
		info["attestationKeyType"] = parsedQuote.Quote.Header.AttestationKeyType
		info["teeType"] = fmt.Sprintf("0x%x", parsedQuote.Quote.Header.TeeType)
		info["qeSvn"] = hex.EncodeToString(parsedQuote.Quote.Header.QeSvn)
		info["pceSvn"] = hex.EncodeToString(parsedQuote.Quote.Header.PceSvn)
	}

	if parsedQuote.Quote != nil && parsedQuote.Quote.TdQuoteBody != nil {
		if parsedQuote.Quote.TdQuoteBody.MrTd != nil {
			info["mrTd"] = hex.EncodeToString(parsedQuote.Quote.TdQuoteBody.MrTd)
		}
		if parsedQuote.Quote.TdQuoteBody.ReportData != nil {
			info["reportData"] = hex.EncodeToString(parsedQuote.Quote.TdQuoteBody.ReportData)
		}
	}

	info["fmspc"] = parsedQuote.FMSPC
	info["tcbComponents"] = parsedQuote.TCBComponents

	return info
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
