package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/go-tdx-guest/verify"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	clickdb "github.com/MoeMahhouk/go-tcb-notify/internal/storage/clickhouse"
)

const serviceName = "validate-quotes"

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

	// Flags via env (simple, no change to config struct required)
	getCollateral := os.Getenv("VALIDATOR_GET_COLLATERAL") != "0"
	checkRevocations := os.Getenv("VALIDATOR_CHECK_REVOCATIONS") == "1"
	batch := cfg.EvaluateQuotes.BatchSize
	if batch <= 0 {
		batch = 200
	}

	for {
		lastBlock, lastIdx, _ := getOffset(ctx, ch)
		n, maxBlock, maxIdx, err := processBatch(ctx, ch, lastBlock, lastIdx, batch, getCollateral, checkRevocations)
		if err != nil {
			log.Printf("validate: %v", err)
		}
		if n > 0 {
			_ = upsertOffset(ctx, ch, maxBlock, maxIdx)
		}
		time.Sleep(3 * time.Second)
	}
}

func processBatch(ctx context.Context, ch clickhouse.Conn, lastBlock uint64, lastIdx uint32, limit int, getColl, checkRev bool) (int, uint64, uint32, error) {
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

		opts := verify.Options{
			GetCollateral:    getColl,
			CheckRevocations: checkRev,
		}
		status := "Verified"
		var reason *string

		if err := verify.TdxQuote(quoteBytes, &opts); err != nil {
			status = "Invalid"
			r := err.Error()
			reason = &r
		}

		if err := ch.Exec(ctx, clickdb.InsertQuoteStatus,
			addr, blockNum, blockTime, txHash, logIndex, qhashStr, status, reason,
		); err != nil {
			return count, maxBlock, maxIdx, err
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
