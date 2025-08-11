package tdxutil

import (
	pb "github.com/google/go-tdx-guest/proto/tdx"
)

// TCBComponents aggregates what we compare against PCS TCB levels.
type TCBComponents struct {
	SGX  [16]uint8
	TDX  [16]uint8
	PCES uint16
}

// ExtractFromQuote gathers (FMSPC, SGX[16], PCESVN) from PCK and TDX[16] from TD report.
func ExtractFromQuote(q *pb.QuoteV4) (fmspcHex string, comps TCBComponents, err error) {
	fmspcHex, err = ExtractFMSPCFromQuote(q)
	if err != nil {
		return "", comps, err
	}
	sgx16, pces, err := ExtractSgxTcbAndPceSvnFromQuote(q)
	if err != nil {
		return "", comps, err
	}
	comps.SGX = sgx16
	comps.PCES = pces

	tdxBytes := q.GetTdQuoteBody().GetTeeTcbSvn() // 16 bytes
	for i := 0; i < 16 && i < len(tdxBytes); i++ {
		comps.TDX[i] = tdxBytes[i]
	}
	return fmspcHex, comps, nil
}
