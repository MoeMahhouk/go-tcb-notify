package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type Config struct {
	LogLevel string
	Debug    bool // Add debug flag

	// Ethereum settings
	Ethereum Ethereum

	// ClickHouse settings
	ClickHouse ClickHouse

	// Service-specific settings
	IngestRegistry IngestRegistry
	EvaluateQuotes EvaluateQuotes
	PCS            PCS
}

type Ethereum struct {
	RPCURL          string
	RegistryAddress common.Address
	StartBlock      uint64 // Add start block
}

type ClickHouse struct {
	Addrs       []string
	Database    string
	Username    string
	Password    string
	DialTimeout time.Duration
	Compression string // "lz4" or "none"
	Secure      bool   // enable TLS (rare for native; default false)
}

type IngestRegistry struct {
	PollInterval time.Duration
	BatchBlocks  uint64
}

type EvaluateQuotes struct {
	BatchSize int
}

type PCS struct {
	BaseURL       string        // e.g. https://api.trustedservices.intel.com
	APIKey        string        // Ocp-Apim-Subscription-Key (if required)
	POLL_INTERVAL time.Duration // Polling interval for PCS updates
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
		log.Printf("WARN: invalid duration %s=%q, using default %s", key, v, def)
	}
	return def
}

func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
		log.Printf("WARN: invalid int %s=%q, using default %d", key, v, def)
	}
	return def
}

func getUint64(key string, def uint64) uint64 {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.ParseUint(v, 10, 64)
		if err == nil {
			return n
		}
		log.Printf("WARN: invalid uint64 %s=%q, using default %d", key, v, def)
	}
	return def
}

func getBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
		log.Printf("WARN: invalid bool %s=%q, using default %v", key, v, def)
	}
	return def
}

func Load() (*Config, error) {
	// Get RPC URL from multiple possible env vars
	rpcURL := getenv("ETHEREUM_RPC_URL", getenv("RPC_URL", "http://localhost:8545"))

	// Registry address
	regAddr := getenv("REGISTRY_ADDRESS", "0x927Ea8b713123744E6E0a92f4417366B0B000dA5")
	if !common.IsHexAddress(regAddr) {
		return nil, fmt.Errorf("invalid REGISTRY_ADDRESS: %s", regAddr)
	}

	// ClickHouse (native protocol)
	chAddrs := strings.Split(getenv("CH_ADDRS", getenv("CLICKHOUSE_ADDRS", "clickhouse:9000")), ",")
	for i := range chAddrs {
		chAddrs[i] = strings.TrimSpace(chAddrs[i])
	}

	ch := ClickHouse{
		Addrs:       chAddrs,
		Database:    getenv("CH_DATABASE", getenv("CLICKHOUSE_DATABASE", "go_tcb_notify")),
		Username:    getenv("CH_USERNAME", getenv("CLICKHOUSE_USERNAME", "default")),
		Password:    getenv("CH_PASSWORD", getenv("CLICKHOUSE_PASSWORD", "")),
		DialTimeout: getDuration("CH_DIAL_TIMEOUT", getDuration("CLICKHOUSE_DIAL_TIMEOUT", 5*time.Second)),
		Compression: strings.ToLower(getenv("CH_COMPRESSION", getenv("CLICKHOUSE_COMPRESSION", "lz4"))),
		Secure:      getBool("CH_SECURE", getBool("CLICKHOUSE_SECURE", false)),
	}

	cfg := &Config{
		LogLevel: getenv("LOG_LEVEL", "info"),
		Debug:    getBool("DEBUG", false),

		Ethereum: Ethereum{
			RPCURL:          rpcURL,
			RegistryAddress: common.HexToAddress(regAddr),
			StartBlock:      getUint64("START_BLOCK", 0),
		},

		ClickHouse: ch,

		IngestRegistry: IngestRegistry{
			PollInterval: getDuration("INGEST_POLL_INTERVAL", 15*time.Minute),
			BatchBlocks:  getUint64("INGEST_BATCH_BLOCKS", 2500),
		},

		EvaluateQuotes: EvaluateQuotes{
			BatchSize: getInt("EVAL_BATCH_SIZE", 500),
		},

		PCS: PCS{
			BaseURL:       getenv("PCS_BASE_URL", "https://api.trustedservices.intel.com"),
			APIKey:        getenv("PCS_API_KEY", ""),
			POLL_INTERVAL: getDuration("PCS_POLL_INTERVAL", 15*time.Minute),
		},
	}
	return cfg, nil
}
