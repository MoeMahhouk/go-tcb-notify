package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
)

type RegistryService struct {
	config    *config.Config
	db        *sql.DB
	ethClient *ethclient.Client
	client    *http.Client
}

type QuoteRegisteredEvent struct {
	Address     string `json:"address"`
	QuoteData   string `json:"quoteData"`
	WorkloadID  string `json:"workloadId"`
	BlockNumber uint64 `json:"blockNumber"`
}

func NewRegistryService(cfg *config.Config, db *sql.DB) (*RegistryService, error) {
	var ethClient *ethclient.Client
	var err error

	if cfg.EthereumRPCURL != "" {
		ethClient, err = ethclient.Dial(cfg.EthereumRPCURL)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to Ethereum: %w", err)
		}
	}

	return &RegistryService{
		config:    cfg,
		db:        db,
		ethClient: ethClient,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (r *RegistryService) FetchQuotesFromRegistry(ctx context.Context) error {
	if r.ethClient == nil {
		return r.fetchQuotesFromAPI(ctx)
	}
	return r.fetchQuotesFromEthereum(ctx)
}

func (r *RegistryService) fetchQuotesFromEthereum(ctx context.Context) error {
	// Get the latest processed block
	lastBlock, err := r.getLastProcessedBlock()
	if err != nil {
		logrus.WithError(err).Error("Failed to get last processed block")
		lastBlock = r.config.StartBlock
	}

	// Get current block number
	currentBlock, err := r.ethClient.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current block number: %w", err)
	}

	if lastBlock >= currentBlock {
		logrus.Debug("No new blocks to process")
		return nil
	}

	// Process blocks in batches
	for start := lastBlock + 1; start <= currentBlock; start += r.config.BatchSize {
		end := start + r.config.BatchSize - 1
		if end > currentBlock {
			end = currentBlock
		}

		if err := r.processBlockRange(ctx, start, end); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"start": start,
				"end":   end,
			}).Error("Failed to process block range")
			return err
		}

		// Update last processed block
		if err := r.updateLastProcessedBlock(end); err != nil {
			logrus.WithError(err).Error("Failed to update last processed block")
		}
	}

	return nil
}

func (r *RegistryService) processBlockRange(ctx context.Context, start, end uint64) error {
	// Query for QuoteRegistered events
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(start)),
		ToBlock:   big.NewInt(int64(end)),
		Addresses: []common.Address{common.HexToAddress(r.config.RegistryAddress)},
		Topics: [][]common.Hash{
			{common.HexToHash("0x1234567890abcdef")}, // QuoteRegistered event signature
		},
	}

	logs, err := r.ethClient.FilterLogs(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to filter logs: %w", err)
	}

	for _, log := range logs {
		if err := r.processQuoteEvent(log.Data, log.TxHash.Hex()); err != nil {
			logrus.WithError(err).WithField("txHash", log.TxHash.Hex()).Error("Failed to process quote event")
		}
	}

	logrus.WithFields(logrus.Fields{
		"start":  start,
		"end":    end,
		"events": len(logs),
	}).Info("Processed block range")

	return nil
}

func (r *RegistryService) processQuoteEvent(data []byte, txHash string) error {
	// Parse the event data to extract quote information
	// This would decode the actual event structure from your registry contract

	// For now, we'll simulate the parsing
	// quoteData := hex.EncodeToString(data)

	// Extract FMSPC and TCB components from quote
	fmspc, tcbComponents, err := r.parseQuoteData(data)
	if err != nil {
		return fmt.Errorf("failed to parse quote data: %w", err)
	}

	// Store the monitored quote
	quote := &models.MonitoredQuote{
		Address:       txHash, // Using txHash as address for now
		QuoteData:     data,
		WorkloadID:    "", // Extract from quote
		FMSPC:         fmspc,
		TCBComponents: tcbComponents,
		CurrentStatus: "UpToDate",
		NeedsUpdate:   false,
		LastChecked:   time.Now(),
	}

	return r.storeMonitoredQuote(quote)
}

func (r *RegistryService) parseQuoteData(data []byte) (string, json.RawMessage, error) {
	// This would use the go-tdx-guest library to parse the quote
	// For now, return placeholder values

	fmspc := "00906EA10000" // Example FMSPC

	// Example TCB components structure
	tcbComponents := map[string]interface{}{
		"sgxTcbComponents": make([]int, 16),
		"tdxTcbComponents": make([]int, 16),
		"pcesvn":           0,
	}

	tcbBytes, err := json.Marshal(tcbComponents)
	if err != nil {
		return "", nil, err
	}

	return fmspc, json.RawMessage(tcbBytes), nil
}

func (r *RegistryService) storeMonitoredQuote(quote *models.MonitoredQuote) error {
	query := `
		INSERT INTO monitored_tdx_quotes 
		(address, quote_data, workload_id, fmspc, tcb_components, current_status, needs_update, last_checked)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (address) 
		DO UPDATE SET 
			quote_data = EXCLUDED.quote_data,
			workload_id = EXCLUDED.workload_id,
			fmspc = EXCLUDED.fmspc,
			tcb_components = EXCLUDED.tcb_components,
			last_checked = EXCLUDED.last_checked`

	_, err := r.db.Exec(query,
		quote.Address,
		quote.QuoteData,
		quote.WorkloadID,
		quote.FMSPC,
		quote.TCBComponents,
		quote.CurrentStatus,
		quote.NeedsUpdate,
		quote.LastChecked,
	)

	if err != nil {
		return fmt.Errorf("failed to store monitored quote: %w", err)
	}

	logrus.WithField("address", quote.Address).Debug("Stored monitored quote")
	return nil
}

func (r *RegistryService) fetchQuotesFromAPI(ctx context.Context) error {
	// Fallback to direct API calls if Ethereum RPC is not available
	logrus.Info("Fetching quotes from registry API...")

	// Implementation would depend on your registry's API structure
	// This is a placeholder for the API-based approach

	return nil
}

func (r *RegistryService) getLastProcessedBlock() (uint64, error) {
	var block uint64
	err := r.db.QueryRow("SELECT COALESCE(MAX(last_processed_block), 0) FROM registry_state").Scan(&block)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return block, nil
}

func (r *RegistryService) updateLastProcessedBlock(block uint64) error {
	_, err := r.db.Exec(`
		INSERT INTO registry_state (id, last_processed_block, updated_at) 
		VALUES (1, $1, CURRENT_TIMESTAMP)
		ON CONFLICT (id) 
		DO UPDATE SET last_processed_block = EXCLUDED.last_processed_block, updated_at = EXCLUDED.updated_at`,
		block)
	return err
}

func (r *RegistryService) GetMonitoredQuotes() ([]*models.MonitoredQuote, error) {
	query := `
		SELECT address, quote_data, workload_id, fmspc, tcb_components, 
		       current_status, needs_update, last_checked
		FROM monitored_tdx_quotes
		ORDER BY last_checked DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quotes []*models.MonitoredQuote
	for rows.Next() {
		quote := &models.MonitoredQuote{}
		err := rows.Scan(
			&quote.Address,
			&quote.QuoteData,
			&quote.WorkloadID,
			&quote.FMSPC,
			&quote.TCBComponents,
			&quote.CurrentStatus,
			&quote.NeedsUpdate,
			&quote.LastChecked,
		)
		if err != nil {
			return nil, err
		}
		quotes = append(quotes, quote)
	}

	return quotes, nil
}
