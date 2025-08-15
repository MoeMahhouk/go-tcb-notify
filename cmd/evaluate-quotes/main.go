package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sirupsen/logrus"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	clickdb "github.com/MoeMahhouk/go-tcb-notify/internal/storage/clickhouse"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/tdx"
)

const serviceName = "evaluate-quotes"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logrus.Info("Shutting down...")
		cancel()
	}()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// Setup logging
	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	ch, err := clickdb.Open(ctx, &cfg.ClickHouse)
	if err != nil {
		log.Fatalf("clickhouse: %v", err)
	}
	defer ch.Close()

	// Note: GetCollateral and CheckRevocations should be part of a verification config
	// For now, we'll use default values
	getCollateral := false    // Default to false for performance
	checkRevocations := false // Default to false for performance

	log.Printf("[eval] Starting: batch_size=%d, get_collateral=%v, check_revocations=%v",
		cfg.EvaluateQuotes.BatchSize, getCollateral, checkRevocations)

	// Get last checkpoint
	lastBlock, lastIdx, err := getOffset(ctx, ch)
	if err != nil {
		log.Printf("WARN get offset: %v (starting from 0)", err)
		lastBlock = 0
		lastIdx = 0
	}

	// Main processing loop
	for {
		if err := processQuotes(ctx, ch, cfg.EvaluateQuotes.BatchSize, getCollateral, checkRevocations, &lastBlock, &lastIdx); err != nil {
			log.Printf("ERROR process: %v", err)
			time.Sleep(5 * time.Second)
		}

		// Check for shutdown
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
			// Continue processing
		}
	}
}

func getOffset(ctx context.Context, ch clickhouse.Conn) (uint64, uint, error) {
	var lastBlock uint64
	var lastIdx uint32
	row := ch.QueryRow(ctx, clickdb.GetOffset, serviceName)
	if err := row.Scan(&lastBlock, &lastIdx); err != nil {
		// No prior offset yet; start from zero (handled in caller)
		return 0, 0, nil
	}
	return lastBlock, uint(lastIdx), nil
}

func processQuotes(ctx context.Context, ch clickhouse.Conn, batchSize int, getCollateral, checkRevocations bool, lastBlock *uint64, lastIdx *uint) error {
	// Query for new events using the query from queries.go
	rows, err := ch.Query(ctx, clickdb.GetUnprocessedQuotes, *lastBlock, *lastBlock, *lastIdx, batchSize)
	if err != nil {
		return fmt.Errorf("query quotes: %w", err)
	}
	defer rows.Close()

	verifier := tdx.NewQuoteVerifier(getCollateral, checkRevocations)
	batch := make([]models.QuoteEvaluation, 0, batchSize)

	for rows.Next() {
		var (
			teeAddress  string
			quoteBytes  []byte
			quoteHash   string
			blockNumber uint64
			logIndex    uint32
			blockTime   time.Time
		)

		if err := rows.Scan(&teeAddress, &quoteBytes, &quoteHash, &blockNumber, &logIndex, &blockTime); err != nil {
			return fmt.Errorf("scan row: %w", err)
		}

		// Verify and evaluate the quote
		evaluation, err := verifier.VerifyAndEvaluateQuote(quoteBytes)
		if err != nil {
			logrus.WithError(err).WithField("hash", quoteHash).Error("Failed to evaluate quote")
			// Create failed evaluation
			evaluation = &models.QuoteEvaluation{
				ServiceAddress: teeAddress,
				QuoteHash:      quoteHash,
				QuoteLength:    len(quoteBytes),
				Status:         models.StatusInvalid,
				TCBStatus:      models.TCBStatusUnknown,
				Error:          err.Error(),
				BlockNumber:    blockNumber,
				LogIndex:       logIndex,
				BlockTime:      blockTime,
				EvaluatedAt:    time.Now(),
			}
		} else {
			// Set additional fields
			evaluation.ServiceAddress = teeAddress
			evaluation.BlockNumber = blockNumber
			evaluation.LogIndex = logIndex
			evaluation.BlockTime = blockTime
			evaluation.EvaluatedAt = time.Now()
		}

		batch = append(batch, *evaluation)

		// Update checkpoint
		*lastBlock = blockNumber
		*lastIdx = uint(logIndex)
	}

	if len(batch) > 0 {
		// Insert evaluations
		if err := insertEvaluations(ctx, ch, batch); err != nil {
			return fmt.Errorf("insert evaluations: %w", err)
		}

		// Update offset
		if err := upsertOffset(ctx, ch, *lastBlock, uint32(*lastIdx)); err != nil {
			return fmt.Errorf("update offset: %w", err)
		}

		log.Printf("[eval] Processed %d quotes, last: block=%d idx=%d", len(batch), *lastBlock, *lastIdx)
	}

	return nil
}

func insertEvaluations(ctx context.Context, ch clickhouse.Conn, evaluations []models.QuoteEvaluation) error {
	batch, err := ch.PrepareBatch(ctx, clickdb.InsertEvaluation)
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	for _, eval := range evaluations {
		// Convert TCBComponents to individual arrays for storage
		sgxComponents := make([]uint8, 16)
		tdxComponents := make([]uint8, 16)
		for i := 0; i < 16; i++ {
			sgxComponents[i] = eval.TCBComponents.SGXComponents[i]
			tdxComponents[i] = eval.TCBComponents.TDXComponents[i]
		}

		err := batch.Append(
			eval.ServiceAddress,
			eval.QuoteHash,
			eval.QuoteLength,
			eval.FMSPC,
			sgxComponents,
			tdxComponents,
			eval.TCBComponents.PCESVN,
			eval.MrTd,
			eval.MrSeam,
			eval.MrSignerSeam,
			eval.ReportData,
			string(eval.Status),
			string(eval.TCBStatus),
			eval.Error,
			eval.BlockNumber,
			eval.LogIndex,
			eval.BlockTime,
			eval.EvaluatedAt,
		)
		if err != nil {
			return fmt.Errorf("append to batch: %w", err)
		}
	}

	return batch.Send()
}

func upsertOffset(ctx context.Context, ch clickhouse.Conn, block uint64, idx uint32) error {
	return ch.Exec(ctx, clickdb.UpsertOffset, serviceName, block, idx)
}
