package ingest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	"github.com/MoeMahhouk/go-tcb-notify/internal/registry/bindings"
	clickdb "github.com/MoeMahhouk/go-tcb-notify/internal/storage/clickhouse"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/tdx"
)

const ServiceName = "ingest-registry"

// Ingester handles registry event ingestion
type Ingester struct {
	db        clickhouse.Conn
	ethClient *ethclient.Client
	registry  *bindings.FlashtestationRegistry
	parser    *tdx.QuoteParser
	config    *config.IngestRegistry
	logger    *logrus.Entry
	lastBlock uint64
	lastIdx   uint
}

// NewIngester creates a new registry ingester service
func NewIngester(db clickhouse.Conn, ethClient *ethclient.Client, registryAddr common.Address, cfg *config.IngestRegistry) (*Ingester, error) {
	registry, err := bindings.NewFlashtestationRegistry(registryAddr, ethClient)
	if err != nil {
		return nil, fmt.Errorf("create registry binding: %w", err)
	}

	return &Ingester{
		db:        db,
		ethClient: ethClient,
		registry:  registry,
		parser:    tdx.NewQuoteParser(),
		config:    cfg,
		logger:    logrus.WithField("service", ServiceName),
	}, nil
}

// Run starts the registry ingestion service
func (i *Ingester) Run(ctx context.Context) error {
	i.logger.WithFields(logrus.Fields{
		"poll_interval": i.config.PollInterval,
		"batch_blocks":  i.config.BatchBlocks,
	}).Info("Starting registry ingestion service")

	// Get last processed checkpoint
	if err := i.loadCheckpoint(ctx); err != nil {
		i.logger.WithError(err).Warn("Failed to load checkpoint, starting from beginning")
	}

	ticker := time.NewTicker(i.config.PollInterval)
	defer ticker.Stop()

	// Process immediately
	i.processOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			i.logger.Info("Registry ingestion service stopped")
			return ctx.Err()
		case <-ticker.C:
			i.processOnce(ctx)
		}
	}
}

// processOnce processes one batch of events
func (i *Ingester) processOnce(ctx context.Context) {
	if err := i.processBatch(ctx); err != nil {
		i.logger.WithError(err).Error("Failed to process batch")
	}
}

// processBatch processes a batch of registry events
func (i *Ingester) processBatch(ctx context.Context) error {
	// Get latest block
	head, err := i.ethClient.HeaderByNumber(ctx, nil)
	if err != nil {
		return fmt.Errorf("get latest block: %w", err)
	}

	// Calculate batch range
	from := i.lastBlock + 1
	if i.lastBlock == 0 && i.lastIdx == 0 {
		from = 0 // First run
	}

	to := minU64(from+i.config.BatchBlocks-1, head.Number.Uint64())
	if from > to {
		return nil // Nothing to process
	}

	// Query events
	events, err := i.fetchEvents(ctx, from, to)
	if err != nil {
		return fmt.Errorf("fetch events: %w", err)
	}

	if len(events) == 0 {
		return nil
	}

	// Process and store events
	processed := 0
	for _, event := range events {
		if err := i.processEvent(ctx, event); err != nil {
			i.logger.WithError(err).WithField("block", event.BlockNumber).Error("Failed to process event")
			continue
		}
		processed++
	}

	// Update checkpoint
	if processed > 0 {
		if err := i.saveCheckpoint(ctx); err != nil {
			return fmt.Errorf("save checkpoint: %w", err)
		}

		i.logger.WithFields(logrus.Fields{
			"processed": processed,
			"from":      from,
			"to":        to,
			"last":      fmt.Sprintf("(%d,%d)", i.lastBlock, i.lastIdx),
		}).Info("Processed registry events")
	}

	return nil
}

// Event represents a registry event
type Event struct {
	Type        string
	TeeAddress  common.Address
	RawQuote    []byte
	BlockNumber uint64
	LogIndex    uint
	TxHash      common.Hash
	BlockTime   time.Time
}

// fetchEvents fetches events from the registry
func (i *Ingester) fetchEvents(ctx context.Context, from, to uint64) ([]Event, error) {
	var events []Event

	// Fetch TEEServiceRegistered events
	regEvents, err := i.fetchRegistrationEvents(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("fetch registration events: %w", err)
	}
	events = append(events, regEvents...)

	// Fetch TEEServiceInvalidated events
	invEvents, err := i.fetchInvalidationEvents(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("fetch invalidation events: %w", err)
	}
	events = append(events, invEvents...)

	// Sort events by block number and log index for proper ordering
	sort.Slice(events, func(i, j int) bool {
		if events[i].BlockNumber != events[j].BlockNumber {
			return events[i].BlockNumber < events[j].BlockNumber
		}
		return events[i].LogIndex < events[j].LogIndex
	})

	// Update last processed position from the last event
	if len(events) > 0 {
		lastEvent := events[len(events)-1]
		i.lastBlock = lastEvent.BlockNumber
		i.lastIdx = lastEvent.LogIndex
	}

	return events, nil
}

// fetchRegistrationEvents fetches TEEServiceRegistered events
func (i *Ingester) fetchRegistrationEvents(ctx context.Context, from, to uint64) ([]Event, error) {
	iter, err := i.registry.FilterTEEServiceRegistered(&bind.FilterOpts{
		Start:   from,
		End:     &to,
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("filter registration events: %w", err)
	}
	defer iter.Close()

	var events []Event
	for iter.Next() {
		ev := iter.Event
		lg := ev.Raw

		// Get block timestamp
		header, err := i.ethClient.HeaderByNumber(ctx, big.NewInt(int64(lg.BlockNumber)))
		if err != nil {
			return nil, fmt.Errorf("get block header: %w", err)
		}

		events = append(events, Event{
			Type:        "TEEServiceRegistered",
			TeeAddress:  ev.TeeAddress,
			RawQuote:    ev.RawQuote,
			BlockNumber: lg.BlockNumber,
			LogIndex:    uint(lg.Index),
			TxHash:      lg.TxHash,
			BlockTime:   time.Unix(int64(header.Time), 0).UTC(),
		})
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("registration iterator error: %w", err)
	}

	return events, nil
}

// fetchInvalidationEvents fetches TEEServiceInvalidated events
func (i *Ingester) fetchInvalidationEvents(ctx context.Context, from, to uint64) ([]Event, error) {
	iter, err := i.registry.FilterTEEServiceInvalidated(&bind.FilterOpts{
		Start:   from,
		End:     &to,
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("filter invalidation events: %w", err)
	}
	defer iter.Close()

	var events []Event
	for iter.Next() {
		ev := iter.Event
		lg := ev.Raw

		// Get block timestamp
		header, err := i.ethClient.HeaderByNumber(ctx, big.NewInt(int64(lg.BlockNumber)))
		if err != nil {
			return nil, fmt.Errorf("get block header: %w", err)
		}

		events = append(events, Event{
			Type:        "TEEServiceInvalidated",
			TeeAddress:  ev.TeeAddress,
			RawQuote:    nil, // No quote data for invalidation events
			BlockNumber: lg.BlockNumber,
			LogIndex:    uint(lg.Index),
			TxHash:      lg.TxHash,
			BlockTime:   time.Unix(int64(header.Time), 0).UTC(),
		})
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("invalidation iterator error: %w", err)
	}

	return events, nil
}

// processEvent processes a single registry event
func (i *Ingester) processEvent(ctx context.Context, event Event) error {
	switch event.Type {
	case "TEEServiceRegistered":
		return i.processRegistrationEvent(ctx, event)
	case "TEEServiceInvalidated":
		return i.processInvalidationEvent(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

// processRegistrationEvent processes a TEEServiceRegistered event
func (i *Ingester) processRegistrationEvent(ctx context.Context, event Event) error {
	// Calculate quote hash
	sum := sha256.Sum256(event.RawQuote)
	quoteHash := hex.EncodeToString(sum[:])

	// Extract FMSPC from quote
	fmspc := ""
	if parsed, err := i.parser.ParseQuote(event.RawQuote); err == nil {
		fmspc = parsed.FMSPC
	}

	// Store in database with event_type = 'Registered'
	return i.db.Exec(ctx, clickdb.InsertQuoteRaw,
		event.TeeAddress.Hex(),
		event.BlockNumber,
		event.BlockTime,
		event.TxHash.Hex(),
		uint32(event.LogIndex),
		"Registered",                       // event_type
		hex.EncodeToString(event.RawQuote), // Store as hex string
		uint32(len(event.RawQuote)),
		quoteHash,
		fmspc,
	)
}

// processInvalidationEvent processes a TEEServiceInvalidated event
func (i *Ingester) processInvalidationEvent(ctx context.Context, event Event) error {
	// Store invalidation event in database
	return i.db.Exec(ctx, clickdb.InsertInvalidationEvent,
		event.TeeAddress.Hex(),
		event.BlockNumber,
		event.BlockTime,
		event.TxHash.Hex(),
		uint32(event.LogIndex),
	)
}

// loadCheckpoint loads the last processed position
func (i *Ingester) loadCheckpoint(ctx context.Context) error {
	var lastBlock uint64
	var lastIdx uint32

	row := i.db.QueryRow(ctx, clickdb.GetOffset, ServiceName)
	if err := row.Scan(&lastBlock, &lastIdx); err != nil {
		i.lastBlock = 0
		i.lastIdx = 0
		return err
	}

	i.lastBlock = lastBlock
	i.lastIdx = uint(lastIdx)
	return nil
}

// saveCheckpoint saves the current position
func (i *Ingester) saveCheckpoint(ctx context.Context) error {
	return i.db.Exec(ctx, clickdb.UpsertOffset, ServiceName, i.lastBlock, uint32(i.lastIdx))
}

func minU64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
