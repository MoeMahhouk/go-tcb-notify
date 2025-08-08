// internal/services/registry_service.go
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
	"github.com/MoeMahhouk/go-tcb-notify/pkg/tdx"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
)

type RegistryService struct {
	config      *config.Config
	db          *sql.DB
	ethClient   *ethclient.Client
	client      *http.Client
	quoteParser *tdx.QuoteParser
}

type QuoteRegisteredEvent struct {
	Address     common.Address `json:"address"`
	QuoteData   []byte         `json:"quoteData"`
	WorkloadID  string         `json:"workloadId"`
	BlockNumber uint64         `json:"blockNumber"`
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
		quoteParser: tdx.NewQuoteParser(),
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
	// Note: You'll need to get the actual event signature from your contract
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(start)),
		ToBlock:   big.NewInt(int64(end)),
		Addresses: []common.Address{common.HexToAddress(r.config.RegistryAddress)},
		// Topics would include the event signature for QuoteRegistered
	}

	logs, err := r.ethClient.FilterLogs(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to filter logs: %w", err)
	}

	for _, log := range logs {
		// Parse the event based on your contract's ABI
		// This is a simplified example - you'll need to decode based on your actual event structure
		event, err := r.parseQuoteRegisteredEvent(&log)
		if err != nil {
			logrus.WithError(err).WithField("txHash", log.TxHash.Hex()).Error("Failed to parse event")
			continue
		}

		if err := r.processQuoteEvent(event); err != nil {
			logrus.WithError(err).WithField("address", event.Address.Hex()).Error("Failed to process quote event")
		}
	}

	logrus.WithFields(logrus.Fields{
		"start":  start,
		"end":    end,
		"events": len(logs),
	}).Info("Processed block range")

	return nil
}

func (r *RegistryService) parseQuoteRegisteredEvent(log *types.Log) (*QuoteRegisteredEvent, error) {
	// This is where you would decode the log data based on your contract's ABI
	// For now, returning a placeholder

	// You would typically use go-ethereum's abi package to decode the event
	// Example:
	// contractAbi, _ := abi.JSON(strings.NewReader(YourContractABI))
	// event := new(QuoteRegisteredEvent)
	// err := contractAbi.UnpackIntoInterface(event, "QuoteRegistered", log.Data)

	// For now, we'll extract basic information from the log
	// Assuming the address is in the first topic after the event signature
	var address common.Address
	if len(log.Topics) > 1 {
		address = common.BytesToAddress(log.Topics[1].Bytes())
	}

	return &QuoteRegisteredEvent{
		Address:     address,
		QuoteData:   log.Data,
		WorkloadID:  "", // Extract from event data
		BlockNumber: log.BlockNumber,
	}, nil
}

func (r *RegistryService) processQuoteEvent(event *QuoteRegisteredEvent) error {
	// Parse the quote data using the TDX quote parser
	parsedQuote, err := r.quoteParser.ParseQuote(event.QuoteData)
	if err != nil {
		return fmt.Errorf("failed to parse quote data: %w", err)
	}

	// Extract FMSPC from the parsed quote
	fmspc := parsedQuote.FMSPC
	if fmspc == "" {
		logrus.WithField("address", event.Address.Hex()).Warn("No FMSPC found in quote, will need to determine later")
	}

	// Convert TCB components to JSON
	tcbComponents := map[string]interface{}{
		"sgxTcbComponents": parsedQuote.TCBComponents.SGXComponents,
		"tdxTcbComponents": parsedQuote.TCBComponents.TDXComponents,
		"pcesvn":           parsedQuote.TCBComponents.PCESVN,
	}

	tcbBytes, err := json.Marshal(tcbComponents)
	if err != nil {
		return fmt.Errorf("failed to marshal TCB components: %w", err)
	}

	// Store the monitored quote
	quote := &models.MonitoredQuote{
		Address:       event.Address.Hex(),
		QuoteData:     event.QuoteData,
		WorkloadID:    event.WorkloadID,
		FMSPC:         fmspc,
		TCBComponents: json.RawMessage(tcbBytes),
		CurrentStatus: "Unknown", // Will be determined during verification
		NeedsUpdate:   true,      // Mark for initial verification
		LastChecked:   time.Now(),
	}

	return r.storeMonitoredQuote(quote)
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
			fmspc = CASE 
				WHEN EXCLUDED.fmspc != '' THEN EXCLUDED.fmspc 
				ELSE monitored_tdx_quotes.fmspc 
			END,
			tcb_components = EXCLUDED.tcb_components,
			needs_update = true,
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

	logrus.WithFields(logrus.Fields{
		"address": quote.Address,
		"fmspc":   quote.FMSPC,
	}).Debug("Stored monitored quote")

	return nil
}

func (r *RegistryService) fetchQuotesFromAPI(ctx context.Context) error {
	// This would be the implementation for fetching quotes via REST API
	// if Ethereum RPC is not available

	logrus.Info("Fetching quotes from registry API...")

	// Example API endpoint - adjust based on your actual API
	apiURL := fmt.Sprintf("%s/api/quotes", r.config.RegistryAddress)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch quotes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Parse response and process quotes
	// This would depend on your API response format

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
