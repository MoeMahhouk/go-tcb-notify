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
	Debug    bool

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
	StartBlock      uint64
}

type ClickHouse struct {
	// Native protocol addresses: host:port (comma-separated)
	Addrs       []string
	Database    string
	Username    string
	Password    string
	DialTimeout time.Duration
	Compression string // "lz4" or "none"
	Secure      bool   // enable TLS
	SkipVerify  bool   // skip TLS verification
}

type IngestRegistry struct {
	PollInterval time.Duration
	BatchBlocks  uint64
}

type EvaluateQuotes struct {
	BatchSize        int
	PollInterval     time.Duration // How often to check for new quotes
	GetCollateral    bool          // Whether to fetch collateral from Intel PCS
	CheckRevocations bool          // Whether to check certificate revocations
}

type PCS struct {
	BaseURL       string        // Intel PCS API base URL
	APIKey        string        // Optional API key
	POLL_INTERVAL time.Duration // How often to refresh TCB info
}

// Helper functions
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
	// RPC URL with multiple fallbacks
	rpcURL := getenv("ETHEREUM_RPC_URL", getenv("RPC_URL", "http://localhost:8545"))

	// Registry address
	regAddr := getenv("REGISTRY_ADDRESS", "0x0000000000000000000000000000000000000000")
	if !common.IsHexAddress(regAddr) {
		return nil, fmt.Errorf("invalid registry address: %s", regAddr)
	}

	// ClickHouse configuration
	ch := ClickHouse{
		Addrs:       strings.Split(getenv("CH_ADDRS", "localhost:9000"), ","),
		Database:    getenv("CH_DATABASE", "default"),
		Username:    getenv("CH_USERNAME", "default"),
		Password:    getenv("CH_PASSWORD", ""),
		DialTimeout: getDuration("CH_DIAL_TIMEOUT", 5*time.Second),
		Compression: getenv("CH_COMPRESSION", "lz4"),
		Secure:      getBool("CH_SECURE", false),
		SkipVerify:  getBool("CH_SKIP_VERIFY", false),
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
			PollInterval: getDuration("INGEST_POLL_INTERVAL", 15*time.Second),
			BatchBlocks:  getUint64("INGEST_BATCH_BLOCKS", 128),
		},

		EvaluateQuotes: EvaluateQuotes{
			BatchSize:        getInt("EVAL_BATCH_SIZE", 500),
			PollInterval:     getDuration("EVAL_POLL_INTERVAL", 5*time.Second),
			GetCollateral:    getBool("EVAL_GET_COLLATERAL", true),    // Default to true for security
			CheckRevocations: getBool("EVAL_CHECK_REVOCATIONS", true), // Default to true for security
		},

		PCS: PCS{
			BaseURL:       getenv("PCS_BASE_URL", "https://api.trustedservices.intel.com"),
			APIKey:        getenv("PCS_API_KEY", ""),
			POLL_INTERVAL: getDuration("PCS_POLL_INTERVAL", 1*time.Hour),
		},
	}

	return cfg, nil
}
