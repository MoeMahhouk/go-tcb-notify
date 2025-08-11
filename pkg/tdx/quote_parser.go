package tdx

import (
	"encoding/hex"
	"fmt"

	tdxabi "github.com/google/go-tdx-guest/abi"
	pb "github.com/google/go-tdx-guest/proto/tdx"
	"github.com/google/go-tdx-guest/verify"
	"github.com/sirupsen/logrus"

	"github.com/MoeMahhouk/go-tcb-notify/pkg/tdxutil"
)

// ParsedQuote represents a parsed TDX quote with extracted components
type ParsedQuote struct {
	Quote         *pb.QuoteV4
	FMSPC         string
	TCBComponents TCBComponents
}

// TCBComponents represents the TCB components needed for verification
type TCBComponents struct {
	SGXComponents [16]uint8
	TDXComponents [16]uint8
	PCESVN        uint16
}

type QuoteParser struct{}

func NewQuoteParser() *QuoteParser { return &QuoteParser{} }

// ParseQuote parses the quote and extracts FMSPC + TCB components using PCK
func (p *QuoteParser) ParseQuote(raw []byte) (*ParsedQuote, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("quote data is empty")
	}

	obj, err := tdxabi.QuoteToProto(raw)
	if err != nil {
		return nil, fmt.Errorf("QuoteToProto: %w", err)
	}
	q, ok := obj.(*pb.QuoteV4)
	if !ok {
		return nil, fmt.Errorf("unexpected quote type: %T", obj)
	}

	fmspc, comps, err := tdxutil.ExtractFromQuote(q)
	if err != nil {
		return nil, fmt.Errorf("extract from quote: %w", err)
	}

	out := &ParsedQuote{
		Quote: q,
		FMSPC: fmspc,
		TCBComponents: TCBComponents{
			SGXComponents: comps.SGX,
			TDXComponents: comps.TDX,
			PCESVN:        comps.PCES,
		},
	}
	logrus.WithFields(logrus.Fields{
		"fmspc": fmspc,
		"sgx":   hex.EncodeToString(out.TCBComponents.SGXComponents[:]),
		"tdx":   hex.EncodeToString(out.TCBComponents.TDXComponents[:]),
		"pces":  out.TCBComponents.PCESVN,
	}).Debug("parsed TDX quote")
	return out, nil
}

func (p *QuoteParser) VerifyQuoteSignature(quoteData []byte, getCollateral bool) error {
	opts := verify.Options{
		GetCollateral:    getCollateral,
		CheckRevocations: false,
	}
	err := verify.TdxQuote(quoteData, &opts)
	return err
}

func (p *QuoteParser) GetQuoteInfo(parsed *ParsedQuote) map[string]any {
	info := map[string]any{
		"fmspc":         parsed.FMSPC,
		"tcbComponents": parsed.TCBComponents,
	}
	if parsed.Quote != nil && parsed.Quote.Header != nil {
		info["version"] = parsed.Quote.Header.Version
		info["attestationKeyType"] = parsed.Quote.Header.AttestationKeyType
		info["teeType"] = fmt.Sprintf("0x%x", parsed.Quote.Header.TeeType)
	}
	if parsed.Quote != nil && parsed.Quote.TdQuoteBody != nil {
		if parsed.Quote.TdQuoteBody.MrTd != nil {
			info["mrTd"] = hex.EncodeToString(parsed.Quote.TdQuoteBody.MrTd)
		}
		if parsed.Quote.TdQuoteBody.ReportData != nil {
			info["reportData"] = hex.EncodeToString(parsed.Quote.TdQuoteBody.ReportData)
		}
	}
	return info
}
