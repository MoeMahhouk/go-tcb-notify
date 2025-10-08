package tdx

import (
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/google/go-tdx-guest/pcs"
	pb "github.com/google/go-tdx-guest/proto/tdx"
)

type SGXFromPCK struct {
	FMSPC  [6]byte
	SGXSVN [16]uint8
	PCESVN uint16
}

// ExtractFMSPCFromQuote returns hex(FMSPC) derived from the PCK leaf inside the quote.
func ExtractFMSPCFromQuote(q *pb.QuoteV4) (string, error) {
	s, err := extractSGXFromPCK(q)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(s.FMSPC[:]), nil
}

// ExtractSgxTcbAndPceSvnFromQuote returns SGX 16 component SVNs and PCESVN from PCK leaf.
func ExtractSgxTcbAndPceSvnFromQuote(q *pb.QuoteV4) ([16]uint8, uint16, error) {
	s, err := extractSGXFromPCK(q)
	if err != nil {
		return [16]uint8{}, 0, err
	}
	return s.SGXSVN, s.PCESVN, nil
}

func extractPckChainFromQuote(q *pb.QuoteV4) ([]byte, error) {
	if q == nil || q.GetSignedData() == nil {
		return nil, errors.New("nil quote/signed data")
	}
	cd := q.GetSignedData().GetCertificationData()
	if cd == nil || cd.GetQeReportCertificationData() == nil {
		return nil, errors.New("nil QE report certification data")
	}
	pck := cd.GetQeReportCertificationData().GetPckCertificateChainData()
	if pck == nil || len(pck.GetPckCertChain()) == 0 {
		return nil, errors.New("missing PCK cert chain in quote")
	}
	return pck.GetPckCertChain(), nil
}

func extractSGXFromPCK(q *pb.QuoteV4) (*SGXFromPCK, error) {
	raw, err := extractPckChainFromQuote(q)
	if err != nil {
		return nil, err
	}

	// The PCK certificate chain is PEM-encoded, so we need to decode it first
	// Following the same approach as go-tdx-guest's extractChainFromQuoteV4
	block, _ := pem.Decode(raw)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("parse PCK chain: failed to decode PEM certificate")
	}

	leaf, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse PCK chain: %w", err)
	}

	// Use go-tdx-guest's PckCertificateExtensions to properly extract SGX extensions
	exts, err := pcs.PckCertificateExtensions(leaf)
	if err != nil {
		return nil, fmt.Errorf("extract PCK extensions: %w", err)
	}

	// Parse FMSPC from hex string
	fmspcBytes, err := hex.DecodeString(exts.FMSPC)
	if err != nil || len(fmspcBytes) != 6 {
		return nil, fmt.Errorf("invalid FMSPC: %s", exts.FMSPC)
	}

	var out SGXFromPCK
	copy(out.FMSPC[:], fmspcBytes)

	// Extract TCB components from the TCB structure
	// CPUSvnComponents contains the 16 SGX TCB component SVNs
	if len(exts.TCB.CPUSvnComponents) >= 16 {
		for i := 0; i < 16; i++ {
			out.SGXSVN[i] = exts.TCB.CPUSvnComponents[i]
		}
	}
	out.PCESVN = exts.TCB.PCESvn

	if out.FMSPC == ([6]byte{}) {
		return nil, errors.New("FMSPC not present in PCK leaf")
	}
	return &out, nil
}
