package tdxutil

import (
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"errors"
	"fmt"

	pb "github.com/google/go-tdx-guest/proto/tdx"
)

// OIDs per Intel SGX PCK Certificate & CRL Profile.
var (
	oidSGXExtensions = asn1.ObjectIdentifier{1, 2, 840, 113741, 1, 13, 1}
	oidFMSPC         = asn1.ObjectIdentifier{1, 2, 840, 113741, 1, 13, 1, 4}
	oidTCBBase       = asn1.ObjectIdentifier{1, 2, 840, 113741, 1, 13, 1, 2}
)

type sgxExtensions []struct {
	ID    asn1.ObjectIdentifier
	Value asn1.RawValue
}

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
	certs, err := x509.ParseCertificates(raw)
	if err != nil || len(certs) == 0 {
		return nil, fmt.Errorf("parse PCK chain: %w", err)
	}
	leaf := certs[0]

	var extRaw []byte
	for _, e := range leaf.Extensions {
		if e.Id.Equal(oidSGXExtensions) {
			extRaw = e.Value
			break
		}
	}
	if len(extRaw) == 0 {
		return nil, errors.New("SGXExtensions OID not found in PCK leaf")
	}

	var seq sgxExtensions
	if _, err := asn1.Unmarshal(extRaw, &seq); err != nil {
		return nil, fmt.Errorf("unmarshal SGXExtensions: %w", err)
	}

	var out SGXFromPCK
	for _, kv := range seq {
		switch {
		case kv.ID.Equal(oidFMSPC):
			var oct asn1.RawValue
			if _, err := asn1.Unmarshal(kv.Value.Bytes, &oct); err == nil && oct.Tag == asn1.TagOctetString {
				if len(oct.Bytes) == 6 {
					copy(out.FMSPC[:], oct.Bytes[:6])
				}
			}
		case hasPrefix(kv.ID, oidTCBBase) && len(kv.ID) == len(oidTCBBase)+1:
			last := int(kv.ID[len(kv.ID)-1]) // 1..18
			switch {
			case 1 <= last && last <= 16:
				var v int
				if _, err := asn1.Unmarshal(kv.Value.Bytes, &v); err == nil && v >= 0 && v <= 0xFF {
					out.SGXSVN[last-1] = uint8(v)
				}
			case last == 17:
				var v int
				if _, err := asn1.Unmarshal(kv.Value.Bytes, &v); err == nil && v >= 0 && v <= 0xFFFF {
					out.PCESVN = uint16(v)
				}
			}
		}
	}
	if out.FMSPC == ([6]byte{}) {
		return nil, errors.New("FMSPC not present in PCK leaf")
	}
	return &out, nil
}

func hasPrefix(a, b asn1.ObjectIdentifier) bool {
	if len(a) < len(b) {
		return false
	}
	for i := range b {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
