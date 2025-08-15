package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	"github.com/MoeMahhouk/go-tcb-notify/internal/registry/bindings"
	clickdb "github.com/MoeMahhouk/go-tcb-notify/internal/storage/clickhouse"
)

const serviceName = "ingest-registry"

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// ClickHouse
	ch, err := clickdb.Open(ctx, &cfg.ClickHouse)
	if err != nil {
		log.Fatalf("clickhouse: %v", err)
	}

	// Ethereum client
	ec, err := ethclient.DialContext(ctx, cfg.Ethereum.RPCURL)
	if err != nil {
		log.Fatalf("ethclient: %v", err)
	}

	reg, err := bindings.NewFlashtestationRegistry(cfg.Ethereum.RegistryAddress, ec)
	if err != nil {
		log.Fatalf("registry binding: %v", err)
	}

	log.Printf("[ingest] start, registry=%s, poll=%s, batch_blocks=%d",
		cfg.Ethereum.RegistryAddress.Hex(), cfg.IngestRegistry.PollInterval, cfg.IngestRegistry.BatchBlocks)

	// Last processed checkpoint (block, log index)
	lastBlock, lastIdx, _ := getOffset(ctx, ch)

	ticker := time.NewTicker(cfg.IngestRegistry.PollInterval)
	defer ticker.Stop()

	for {
		if err := runOnce(ctx, ch, ec, reg, &lastBlock, &lastIdx, cfg.IngestRegistry.BatchBlocks); err != nil {
			log.Printf("WARN runOnce: %v", err)
			time.Sleep(3 * time.Second)
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}

func runOnce(ctx context.Context, ch clickhouse.Conn, ec *ethclient.Client, reg *bindings.FlashtestationRegistry,
	lastBlock *uint64, lastIdx *uint, batchBlocks uint64) error {

	// Determine latest chain head
	head, err := ec.HeaderByNumber(ctx, nil)
	if err != nil {
		return fmt.Errorf("headerByNumber latest: %w", err)
	}
	latest := head.Number.Uint64()

	// Compute query window
	from := *lastBlock
	to := minU64(latest, from+batchBlocks)
	if from == 0 {
		// If no offset was stored yet, scan just the most recent batch window
		to = latest
		if latest > batchBlocks {
			from = latest - batchBlocks
		}
	}
	if from > to {
		return nil
	}

	opts := &bind.FilterOpts{
		Start:   from,
		End:     &to,
		Context: ctx,
	}

	iter, err := reg.FilterTEEServiceRegistered(opts)
	if err != nil {
		return fmt.Errorf("filter events [%d..%d]: %w", from, to, err)
	}
	defer iter.Close()

	batch := 0
	for iter.Next() {
		ev := iter.Event
		lg := ev.Raw // types.Log

		// Block timestamp (portable): fetch the header for this block
		h, err := ec.HeaderByNumber(ctx, big.NewInt(int64(lg.BlockNumber)))
		if err != nil {
			return fmt.Errorf("fetch header for ts: %w", err)
		}
		blockTime := time.Unix(int64(h.Time), 0).UTC()

		q := ev.RawQuote
		sum := sha256.Sum256(q)

		if err := ch.Exec(ctx, clickdb.InsertQuoteRaw,
			ev.TeeAddress.Hex(),
			uint64(lg.BlockNumber),
			blockTime,
			lg.TxHash.Hex(),
			uint32(lg.Index),
			q,
			uint32(len(q)),
			hex.EncodeToString(sum[:]),
		); err != nil {
			return fmt.Errorf("insert raw: %w", err)
		}

		*lastBlock = lg.BlockNumber
		*lastIdx = uint(lg.Index)
		batch++
	}
	if err := iter.Error(); err != nil {
		return fmt.Errorf("iterator: %w", err)
	}
	if batch > 0 {
		if err := upsertOffset(ctx, ch, *lastBlock, uint32(*lastIdx)); err != nil {
			return err
		}
		log.Printf("[ingest] %d events [%d..%d], last=(%d,%d)", batch, from, to, *lastBlock, *lastIdx)
	}
	return nil
}

func getOffset(ctx context.Context, ch clickhouse.Conn) (uint64, uint, error) {
	var lastBlock uint64
	var lastIdx uint32
	row := ch.QueryRow(ctx, clickdb.GetOffset, serviceName)
	if err := row.Scan(&lastBlock, &lastIdx); err != nil {
		// No prior offset yet; start from zero (handled in runOnce)
		return 0, 0, nil
	}
	return lastBlock, uint(lastIdx), nil
}

func upsertOffset(ctx context.Context, ch clickhouse.Conn, block uint64, idx uint32) error {
	return ch.Exec(ctx, clickdb.UpsertOffset, serviceName, block, idx)
}

func minU64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
