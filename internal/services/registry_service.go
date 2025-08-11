package services

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"

	ethereum "github.com/ethereum/go-ethereum"
	gethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	tdxabi "github.com/google/go-tdx-guest/abi"
	pb "github.com/google/go-tdx-guest/proto/tdx"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	"github.com/MoeMahhouk/go-tcb-notify/internal/eth"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/tdxutil"
	"github.com/sirupsen/logrus"
)

type RegistryService struct {
	cfg        *config.Config
	db         *sql.DB
	regABI     gethabi.ABI
	regAddress common.Address
	event      gethabi.Event
}

func NewRegistryService(cfg *config.Config, db *sql.DB) (*RegistryService, error) {
	if cfg.RegistryAddress == "" {
		return nil, fmt.Errorf("REGISTRY_ADDRESS is required")
	}
	addr := common.HexToAddress(cfg.RegistryAddress)
	ab, err := eth.LoadABIFromFile(cfg.RegistryABIPath)
	if err != nil {
		return nil, fmt.Errorf("load registry ABI: %w", err)
	}
	name := cfg.RegistryEventName
	ev, ok := ab.Events[name]
	if !ok {
		return nil, fmt.Errorf("%s event not found in ABI (check REGISTRY_ABI_PATH / REGISTRY_EVENT_NAME)", name)
	}
	return &RegistryService{
		cfg:        cfg,
		db:         db,
		regABI:     ab,
		regAddress: addr,
		event:      ev,
	}, nil
}

// GetMonitoredQuotes returns all rows from monitored_tdx_quotes
func (s *RegistryService) GetMonitoredQuotes() ([]*models.MonitoredQuote, error) {
	rows, err := s.db.Query(`
		SELECT address, quote_data, workload_id, fmspc, tcb_components, current_status, needs_update, last_checked
		FROM monitored_tdx_quotes
		ORDER BY address`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*models.MonitoredQuote
	for rows.Next() {
		m := &models.MonitoredQuote{}
		if err := rows.Scan(&m.Address, &m.QuoteData, &m.WorkloadID, &m.FMSPC, &m.TCBComponents, &m.CurrentStatus, &m.NeedsUpdate, &m.LastChecked); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// FetchQuotesFromRegistry scans QuoteRegistered logs and upserts rows.
func (s *RegistryService) FetchQuotesFromRegistry(ctx context.Context) error {
	if s.cfg.EthereumRPCURL == "" {
		return fmt.Errorf("ETHEREUM_RPC_URL / RPC_URL not configured")
	}
	cli, err := ethclient.DialContext(ctx, s.cfg.EthereumRPCURL)
	if err != nil {
		return fmt.Errorf("dial RPC: %w", err)
	}
	defer cli.Close()

	head, err := cli.HeaderByNumber(ctx, nil)
	if err != nil {
		return fmt.Errorf("get latest header: %w", err)
	}
	latest := head.Number.Uint64()

	from := s.cfg.StartBlock
	if dbLast, err := s.getLastProcessedBlock(); err == nil && dbLast > 0 {
		if dbLast+1 > from {
			from = dbLast + 1
		}
	}
	if from == 0 {
		from = 0
	}
	if from > latest {
		logrus.WithFields(logrus.Fields{"from": from, "latest": latest}).Info("registry up to date")
		return nil
	}

	batch := s.cfg.BatchSize
	if batch == 0 {
		batch = 1000
	}
	var processed uint64
	for start := from; start <= latest; {
		end := start + batch - 1
		if end > latest {
			end = latest
		}

		logrus.WithFields(logrus.Fields{
			"from": start,
			"to":   end,
		}).Info("scanning QuoteRegistered logs")

		logs, err := cli.FilterLogs(ctx, buildQuery(s.regAddress, s.event, start, end))
		if err != nil {
			return fmt.Errorf("FilterLogs(%d-%d): %w", start, end, err)
		}

		for _, lg := range logs {
			if lg.Topics[0] != s.event.ID {
				continue
			}
			addr, workloadID, quote, err := s.decodeEvent(lg)
			if err != nil {
				logrus.WithError(err).Warn("failed to decode QuoteRegistered")
				continue
			}
			if err := s.ingestQuote(addr, workloadID, quote); err != nil {
				logrus.WithError(err).Warn("failed to ingest quote")
				continue
			}
			processed = lg.BlockNumber
		}

		if processed > 0 {
			if err := s.setLastProcessedBlock(processed); err != nil {
				logrus.WithError(err).Warn("failed to persist last processed block")
			}
		}

		if end == math.MaxUint64 || end == latest {
			break
		}
		start = end + 1
	}
	return nil
}

func buildQuery(addr common.Address, ev gethabi.Event, from, to uint64) (q ethereum.FilterQuery) {
	q.FromBlock = new(big.Int).SetUint64(from)
	q.ToBlock = new(big.Int).SetUint64(to)
	q.Addresses = []common.Address{addr}
	q.Topics = [][]common.Hash{{ev.ID}}
	return
}

func (s *RegistryService) decodeEvent(lg types.Log) (addr common.Address, workloadID [32]byte, quote []byte, err error) {
	// Unpack non-indexed data into a map so we can find the 'bytes' payload regardless of its field name.
	m := map[string]any{}
	if err = s.regABI.UnpackIntoMap(m, s.event.Name, lg.Data); err != nil {
		return addr, workloadID, nil, fmt.Errorf("unpack data: %w", err)
	}
	for _, in := range s.event.Inputs {
		if in.Indexed {
			// none in the current ABI, but keep for completeness
			continue
		}
		switch in.Type.String() {
		case "address":
			if v, ok := m[in.Name].(common.Address); ok {
				addr = v
			} else if b, ok := m[in.Name].([20]byte); ok {
				addr = common.BytesToAddress(b[:])
			}
		case "bytes":
			if b, ok := m[in.Name].([]byte); ok {
				quote = b
			}
		case "bytes32":
			if b, ok := m[in.Name].([32]byte); ok {
				workloadID = b
			}
		}
	}
	if len(quote) == 0 {
		return addr, workloadID, nil, fmt.Errorf("quote bytes not found in event data")
	}
	// No workloadId in your event => derive it deterministically from the quote
	if workloadID == ([32]byte{}) {
		h := crypto.Keccak256Hash(quote)
		copy(workloadID[:], h.Bytes())
	}
	return
}

func (s *RegistryService) ingestQuote(evmAddr common.Address, workloadID [32]byte, rawQuote []byte) error {
	obj, err := tdxabi.QuoteToProto(rawQuote)
	if err != nil {
		return fmt.Errorf("QuoteToProto: %w", err)
	}
	q, ok := obj.(*pb.QuoteV4)
	if !ok {
		return fmt.Errorf("unexpected proto type: %T", obj)
	}
	fmspc, comps, err := tdxutil.ExtractFromQuote(q)
	if err != nil {
		return fmt.Errorf("extract from quote: %w", err)
	}
	tcbJSON, _ := json.Marshal(struct {
		SGXTCBComponents [16]uint8 `json:"sgxTcbComponents"`
		TDXTCBComponents [16]uint8 `json:"tdxTcbComponents"`
		PCESVN           int       `json:"pcesvn"`
	}{
		SGXTCBComponents: comps.SGX,
		TDXTCBComponents: comps.TDX,
		PCESVN:           int(comps.PCES),
	})

	// Ensure FMSPC row exists (FK).
	if _, err := s.db.Exec(`
		INSERT INTO fmspcs (fmspc, platform, updated_at)
		VALUES ($1, 'TDX', CURRENT_TIMESTAMP)
		ON CONFLICT (fmspc) DO NOTHING
	`, fmspc); err != nil {
		logrus.WithError(err).Warn("failed to upsert fmspc")
	}

	_, err = s.db.Exec(`
		INSERT INTO monitored_tdx_quotes (
			address, quote_data, workload_id, fmspc, tcb_components, current_status, needs_update, last_checked
		) VALUES ($1, $2, $3, $4, $5, 'Unknown', true, CURRENT_TIMESTAMP)
		ON CONFLICT (address) DO UPDATE SET
			quote_data = EXCLUDED.quote_data,
			workload_id = EXCLUDED.workload_id,
			fmspc = EXCLUDED.fmspc,
			tcb_components = EXCLUDED.tcb_components,
			last_checked = CURRENT_TIMESTAMP
	`, evmAddr.Hex(), rawQuote, hex.EncodeToString(workloadID[:]), fmspc, tcbJSON)
	if err != nil {
		return fmt.Errorf("upsert monitored_tdx_quotes: %w", err)
	}
	logrus.WithFields(logrus.Fields{
		"address":    evmAddr.Hex(),
		"workloadId": hex.EncodeToString(workloadID[:]),
		"fmspc":      fmspc,
	}).Info("ingested QuoteRegistered")
	return nil
}

func (s *RegistryService) getLastProcessedBlock() (uint64, error) {
	var v sql.NullInt64
	if err := s.db.QueryRow(`SELECT last_processed_block FROM registry_state WHERE id=1`).Scan(&v); err != nil {
		// table exists from migrations; if row missing, treat as 0
		return 0, nil
	}
	if v.Valid && v.Int64 > 0 {
		return uint64(v.Int64), nil
	}
	return 0, nil
}

func (s *RegistryService) setLastProcessedBlock(bn uint64) error {
	_, err := s.db.Exec(`
		INSERT INTO registry_state (id, last_processed_block, updated_at)
		VALUES (1, $1, CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO UPDATE SET last_processed_block=$1, updated_at=CURRENT_TIMESTAMP
	`, bn)
	return err
}
