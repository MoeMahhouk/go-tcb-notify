package tdx

import (
	"encoding/hex"
	"fmt"

	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
	tdxabi "github.com/google/go-tdx-guest/abi"
	pb "github.com/google/go-tdx-guest/proto/tdx"
	"github.com/sirupsen/logrus"
)

type QuoteParser struct{}

func NewQuoteParser() *QuoteParser { return &QuoteParser{} }

// ParseQuote parses the quote and extracts FMSPC + TCB components using PCK
func (p *QuoteParser) ParseQuote(raw []byte) (*models.ParsedQuote, error) {
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

	fmspc, comps, err := ExtractFromQuote(q)
	if err != nil {
		return nil, fmt.Errorf("extract from quote: %w", err)
	}

	out := &models.ParsedQuote{
		Quote: q,
		FMSPC: fmspc,
		TCBComponents: models.TCBComponents{
			SGXComponents: comps.SGXComponents,
			TDXComponents: comps.TDXComponents,
			PCESVN:        comps.PCESVN,
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

func (p *QuoteParser) GetQuoteInfo(parsed *models.ParsedQuote) map[string]any {
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
