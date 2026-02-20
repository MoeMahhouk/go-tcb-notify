package ingest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	"github.com/MoeMahhouk/go-tcb-notify/internal/registry/bindings"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/storage"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/tdx"
)

const ServiceName = "ingest-registry"

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

// QuoteProcessor handles the business logic of processing quotes
type QuoteProcessor struct {
	parser *tdx.QuoteParser
}

// NewQuoteProcessor creates a new quote processor
func NewQuoteProcessor() *QuoteProcessor {
	return &QuoteProcessor{
		parser: tdx.NewQuoteParser(),
	}
}

// ProcessQuote processes a raw quote and extracts relevant information
func (p *QuoteProcessor) ProcessQuote(raw []byte) (*models.RegistryQuote, error) {
	// Calculate quote hash
	sum := sha256.Sum256(raw)
	quoteHash := hex.EncodeToString(sum[:])

	// Extract FMSPC from quote
	fmspc := ""
	parsed, err := p.parser.ParseQuote(raw)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"quote_len": len(raw),
			"error":     err,
		}).Warn("Failed to parse quote for FMSPC extraction")
	} else {
		fmspc = parsed.FMSPC
		logrus.WithFields(logrus.Fields{
			"fmspc":     fmspc,
			"quote_len": len(raw),
		}).Debug("Successfully extracted FMSPC from quote")
	}

	return &models.RegistryQuote{
		QuoteBytes:  raw,
		QuoteLength: uint32(len(raw)),
		QuoteHash:   quoteHash,
		FMSPC:       fmspc,
	}, nil
}

// Ingester handles registry event ingestion with clean separation of concerns
type Ingester struct {
	// Storage interfaces
	quoteStore  storage.QuoteStore
	offsetStore storage.OffsetStore

	// External dependencies
	ethClient *ethclient.Client
	registry  *bindings.FlashtestationRegistry

	// Business logic
	processor *QuoteProcessor

	// Configuration and logging
	config     *config.IngestRegistry
	startBlock uint64
	logger     *logrus.Entry

	// State
	lastBlock uint64
	lastIdx   uint
}

// NewIngester creates a new registry ingester service with dependency injection
func NewIngester(
	quoteStore storage.QuoteStore,
	offsetStore storage.OffsetStore,
	ethClient *ethclient.Client,
	registryAddr common.Address,
	cfg *config.IngestRegistry,
	startBlock uint64,
) (*Ingester, error) {
	registry, err := bindings.NewFlashtestationRegistry(registryAddr, ethClient)
	if err != nil {
		return nil, fmt.Errorf("create registry binding: %w", err)
	}

	return &Ingester{
		quoteStore:  quoteStore,
		offsetStore: offsetStore,
		ethClient:   ethClient,
		registry:    registry,
		processor:   NewQuoteProcessor(),
		config:      cfg,
		startBlock:  startBlock,
		logger:      logrus.WithField("service", ServiceName),
	}, nil
}

// Run starts the registry ingestion service
func (i *Ingester) Run(ctx context.Context) error {
	i.logger.WithFields(logrus.Fields{
		"poll_interval": i.config.PollInterval,
		"batch_blocks":  i.config.BatchBlocks,
		"start_block":   i.startBlock,
	}).Info("Starting registry ingestion service")

	// Get last processed checkpoint
	if err := i.loadCheckpoint(ctx); err != nil {
		i.logger.WithError(err).Warn("Failed to load checkpoint, starting from configured start block")
	}

	for {
		select {
		case <-ctx.Done():
			i.logger.Info("Registry ingestion service stopped")
			return ctx.Err()
		default:
		}

		caughtUp, err := i.processBatch(ctx)
		if err != nil {
			i.logger.WithError(err).Error("Failed to process batch")
		}

		// Wait before next batch: use poll interval when at tip or on error, skip when catching up
		if caughtUp || err != nil {
			select {
			case <-ctx.Done():
				i.logger.Info("Registry ingestion service stopped")
				return ctx.Err()
			case <-time.After(i.config.PollInterval):
			}
		}
	}
}

// processBatch processes a batch of registry events.
// Returns true when the ingester is caught up to the chain tip.
func (i *Ingester) processBatch(ctx context.Context) (caughtUp bool, err error) {
	// Get latest block
	head, err := i.ethClient.HeaderByNumber(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("get latest block: %w", err)
	}

	// Safety margin: the RPC may report a head via eth_blockNumber that
	// eth_getLogs can't serve yet. Subtract a few blocks to avoid this race.
	const tipSafetyMargin = 4
	headNum := head.Number.Uint64()
	if headNum > tipSafetyMargin {
		headNum -= tipSafetyMargin
	}

	// Calculate batch range
	from := i.lastBlock + 1
	to := min(from+i.config.BatchBlocks-1, headNum)
	if from > headNum {
		return true, nil // At the tip, nothing to process
	}

	// Query events
	events, err := i.fetchEvents(ctx, from, to)
	if err != nil {
		return false, fmt.Errorf("fetch events: %w", err)
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

	// Advance checkpoint to end of batch
	i.lastBlock = to
	i.lastIdx = 0

	if processed > 0 {
		i.logger.WithFields(logrus.Fields{
			"processed": processed,
			"from":      from,
			"to":        to,
		}).Info("Processed registry events")
	} else {
		i.logger.WithFields(logrus.Fields{
			"from": from,
			"to":   to,
		}).Debug("No events in batch, advancing checkpoint")
	}

	if err := i.saveCheckpoint(ctx); err != nil {
		return false, fmt.Errorf("save checkpoint: %w", err)
	}

	return to >= headNum, nil
}

// fetchEvents fetches events from the registry
func (i *Ingester) fetchEvents(ctx context.Context, from, to uint64) ([]Event, error) {
	var events []Event
	headerCache := make(map[uint64]time.Time)

	// Fetch TEEServiceRegistered events
	regEvents, err := i.fetchRegistrationEvents(ctx, from, to, headerCache)
	if err != nil {
		return nil, fmt.Errorf("fetch registration events: %w", err)
	}
	events = append(events, regEvents...)

	// Fetch TEEServiceInvalidated events
	invEvents, err := i.fetchInvalidationEvents(ctx, from, to, headerCache)
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

	return events, nil
}

// getBlockTime returns the timestamp for a block, using a cache to avoid redundant RPC calls.
func (i *Ingester) getBlockTime(ctx context.Context, blockNumber uint64, cache map[uint64]time.Time) (time.Time, error) {
	if t, ok := cache[blockNumber]; ok {
		return t, nil
	}
	header, err := i.ethClient.HeaderByNumber(ctx, big.NewInt(int64(blockNumber)))
	if err != nil {
		return time.Time{}, fmt.Errorf("get block header %d: %w", blockNumber, err)
	}
	t := time.Unix(int64(header.Time), 0).UTC()
	cache[blockNumber] = t
	return t, nil
}

// fetchRegistrationEvents fetches TEEServiceRegistered events
func (i *Ingester) fetchRegistrationEvents(ctx context.Context, from, to uint64, headerCache map[uint64]time.Time) ([]Event, error) {
	iter, err := i.registry.FilterTEEServiceRegistered(&bind.FilterOpts{
		Start:   from,
		End:     &to,
		Context: ctx,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("filter registration events: %w", err)
	}
	defer iter.Close()

	var events []Event
	for iter.Next() {
		ev := iter.Event
		lg := ev.Raw

		blockTime, err := i.getBlockTime(ctx, lg.BlockNumber, headerCache)
		if err != nil {
			return nil, err
		}

		events = append(events, Event{
			Type:        "TEEServiceRegistered",
			TeeAddress:  ev.TeeAddress,
			RawQuote:    ev.RawQuote,
			BlockNumber: lg.BlockNumber,
			LogIndex:    uint(lg.Index),
			TxHash:      lg.TxHash,
			BlockTime:   blockTime,
		})
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("registration iterator error: %w", err)
	}

	return events, nil
}

// fetchInvalidationEvents fetches TEEServiceInvalidated events
func (i *Ingester) fetchInvalidationEvents(ctx context.Context, from, to uint64, headerCache map[uint64]time.Time) ([]Event, error) {
	iter, err := i.registry.FilterTEEServiceInvalidated(&bind.FilterOpts{
		Start:   from,
		End:     &to,
		Context: ctx,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("filter invalidation events: %w", err)
	}
	defer iter.Close()

	var events []Event
	for iter.Next() {
		ev := iter.Event
		lg := ev.Raw

		blockTime, err := i.getBlockTime(ctx, lg.BlockNumber, headerCache)
		if err != nil {
			return nil, err
		}

		events = append(events, Event{
			Type:        "TEEServiceInvalidated",
			TeeAddress:  ev.TeeAddress,
			RawQuote:    nil,
			BlockNumber: lg.BlockNumber,
			LogIndex:    uint(lg.Index),
			TxHash:      lg.TxHash,
			BlockTime:   blockTime,
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
	// DEBUG: Log raw quote info
	i.logger.WithFields(logrus.Fields{
		"address":       event.TeeAddress.Hex(),
		"block":         event.BlockNumber,
		"raw_quote_len": len(event.RawQuote),
		"first_bytes":   hex.EncodeToString(event.RawQuote[:min(20, len(event.RawQuote))]),
	}).Debug("Processing registration event")

	// Process the quote (business logic)
	processed, err := i.processor.ProcessQuote(event.RawQuote)
	if err != nil {
		return fmt.Errorf("process quote: %w", err)
	}

	// Prepare the quote for storage
	quote := &models.RegistryQuote{
		ServiceAddress: event.TeeAddress.Hex(),
		BlockNumber:    event.BlockNumber,
		BlockTime:      event.BlockTime,
		TxHash:         event.TxHash.Hex(),
		LogIndex:       uint32(event.LogIndex),
		QuoteBytes:     processed.QuoteBytes,
		QuoteLength:    processed.QuoteLength,
		QuoteHash:      processed.QuoteHash,
		FMSPC:          processed.FMSPC,
		IngestedAt:     time.Now().UTC(),
	}

	// Store using the interface (no SQL here!)
	return i.quoteStore.StoreRawQuote(ctx, quote)
}

// processInvalidationEvent processes a TEEServiceInvalidated event
func (i *Ingester) processInvalidationEvent(ctx context.Context, event Event) error {
	// Store invalidation using the interface
	return i.quoteStore.StoreInvalidation(
		ctx,
		event.TeeAddress.Hex(),
		event.BlockNumber,
		event.BlockTime,
		event.TxHash.Hex(),
		uint32(event.LogIndex),
	)
}

// loadCheckpoint loads the last processed position
func (i *Ingester) loadCheckpoint(ctx context.Context) error {
	blockNumber, logIndex, err := i.offsetStore.LoadOffset(ctx, ServiceName)
	if err != nil {
		// No checkpoint found — set lastBlock so that from = lastBlock + 1 = startBlock
		if i.startBlock > 0 {
			i.lastBlock = i.startBlock - 1
		} else {
			i.lastBlock = 0
		}
		i.lastIdx = 0
		return err
	}

	i.lastBlock = blockNumber
	i.lastIdx = uint(logIndex)
	return nil
}

// saveCheckpoint saves the current position
func (i *Ingester) saveCheckpoint(ctx context.Context) error {
	return i.offsetStore.SaveOffset(ctx, ServiceName, i.lastBlock, uint32(i.lastIdx))
}

