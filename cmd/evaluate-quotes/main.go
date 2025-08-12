package main

import (
	"context"
	"encoding/hex"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	clickdb "github.com/MoeMahhouk/go-tcb-notify/internal/storage/clickhouse"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/tdx"
)

const serviceName = "evaluate-quotes"

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	ch, err := clickdb.Open(ctx, &cfg.ClickHouse)
	if err != nil {
		log.Fatalf("clickhouse: %v", err)
	}

	parser := tdx.NewQuoteParser()

	for {
		lastBlock, lastIdx, _ := getOffset(ctx, ch)
		n, maxBlock, maxIdx, err := processBatch(ctx, ch, parser, lastBlock, lastIdx, cfg.EvaluateQuotes.BatchSize)
		if err != nil {
			log.Printf("evaluate: %v", err)
		}
		if n > 0 {
			_ = upsertOffset(ctx, ch, maxBlock, maxIdx)
		}
		time.Sleep(3 * time.Second)
	}
}

func processBatch(ctx context.Context, ch clickhouse.Conn, parser *tdx.QuoteParser, lastBlock uint64, lastIdx uint32, limit int) (int, uint64, uint32, error) {
	rows, err := ch.Query(ctx, clickdb.SelectQuotesAfter, lastBlock, lastBlock, lastIdx, limit)
	if err != nil {
		return 0, lastBlock, lastIdx, err
	}
	defer rows.Close()

	count := 0
	maxBlock := lastBlock
	maxIdx := lastIdx

	for rows.Next() {
		var (
			addr       string
			blockNum   uint64
			blockTime  time.Time
			txHash     string
			logIndex   uint32
			quoteBytes []byte
			qlen       uint32
			qhashStr   string
		)
		if err := rows.Scan(&addr, &blockNum, &blockTime, &txHash, &logIndex, &quoteBytes, &qlen, &qhashStr); err != nil {
			return count, maxBlock, maxIdx, err
		}

		// Parse via pkg/tdx -> tdxutil (FMSPC, TDX/SGX SVNs, MR_TD, ReportData…)
		pq, err := parser.ParseQuote(quoteBytes)
		if err != nil {
			// Still record a parsed row with minimal fields for observability
			errStr := err.Error()
			_ = ch.Exec(ctx, clickdb.InsertQuoteParsed,
				addr, blockNum, blockTime, txHash, logIndex, qlen, qhashStr,
				nil, nil, nil, nil, nil, nil,
				nil, nil, nil, nil,
				&errStr,
			)
		} else {
			// Map to columns we currently have
			var (
				teeTCBSVN     *string
				mrSeam        *string
				mrSignerSeam  *string
				seamSVN       *string
				attributes    *string
				xfam          *string
				mrTD          *string
				mrOwner       *string
				mrOwnerConfig *string
				mrConfigID    *string
				reportData    *string
			)

			if b := pq.Quote.GetTdQuoteBody().GetTeeTcbSvn(); len(b) > 0 {
				s := hex.EncodeToString(b)
				teeTCBSVN = &s
			}
			if b := pq.Quote.GetTdQuoteBody().GetMrTd(); len(b) > 0 {
				s := hex.EncodeToString(b)
				mrTD = &s
			}
			if b := pq.Quote.GetTdQuoteBody().GetReportData(); len(b) > 0 {
				s := hex.EncodeToString(b)
				reportData = &s
			}
			// (seam/mr_seam fields are not always present in TD10 report body; keep them nil for now)

			if err := ch.Exec(ctx, clickdb.InsertQuoteParsed,
				addr, blockNum, blockTime, txHash, logIndex, qlen, qhashStr,
				teeTCBSVN, mrSeam, mrSignerSeam, seamSVN, attributes, xfam,
				mrTD, mrOwner, mrOwnerConfig, mrConfigID, reportData,
			); err != nil {
				return count, maxBlock, maxIdx, err
			}
		}

		if blockNum > maxBlock || (blockNum == maxBlock && logIndex > maxIdx) {
			maxBlock = blockNum
			maxIdx = logIndex
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return count, maxBlock, maxIdx, err
	}
	return count, maxBlock, maxIdx, nil
}

func getOffset(ctx context.Context, ch clickhouse.Conn) (uint64, uint32, error) {
	var lastBlock uint64
	var lastIdx uint32
	row := ch.QueryRow(ctx, clickdb.GetOffset, serviceName)
	if err := row.Scan(&lastBlock, &lastIdx); err != nil {
		return 0, 0, nil
	}
	return lastBlock, lastIdx, nil
}

func upsertOffset(ctx context.Context, ch clickhouse.Conn, block uint64, idx uint32) error {
	return ch.Exec(ctx, clickdb.UpsertOffset, serviceName, block, idx)
}
