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

type ClickHouse struct {
	// Native protocol addresses: host:port (comma-separated), e.g. "clickhouse:9000"
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
	FMSPCs        []string      // comma-separated list in env
	POLL_INTERVAL time.Duration // e.g. 24h or 1h30m
}

type Config struct {
	LogLevel        string
	RPCURL          string
	RegistryAddress common.Address

	ClickHouse     ClickHouse
	IngestRegistry IngestRegistry
	EvaluateQuotes EvaluateQuotes
	PCS            PCS
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
	// Defaults point to your values so it “just works” until customized.
	rpcURL := getenv("RPC_URL", "http://experi-proxy-cdzrhzcy6czr-74615020.us-east-2.elb.amazonaws.com/")
	regAddr := getenv("REGISTRY_ADDRESS", "0x927Ea8b713123744E6E0a92f4417366B0B000dA5")
	if !common.IsHexAddress(regAddr) {
		return nil, fmt.Errorf("invalid REGISTRY_ADDRESS: %s", regAddr)
	}

	// ClickHouse (native protocol)
	chAddrs := strings.Split(getenv("CH_ADDRS", "clickhouse:9000"), ",")
	for i := range chAddrs {
		chAddrs[i] = strings.TrimSpace(chAddrs[i])
	}
	ch := ClickHouse{
		Addrs:       chAddrs,
		Database:    getenv("CH_DATABASE", "go_tcb_notify"),
		Username:    getenv("CH_USERNAME", "default"),
		Password:    getenv("CH_PASSWORD", ""),
		DialTimeout: getDuration("CH_DIAL_TIMEOUT", 5*time.Second),
		Compression: strings.ToLower(getenv("CH_COMPRESSION", "lz4")),
		Secure:      getBool("CH_SECURE", false),
	}

	cfg := &Config{
		LogLevel:        getenv("LOG_LEVEL", "info"),
		RPCURL:          rpcURL,
		RegistryAddress: common.HexToAddress(regAddr),
		ClickHouse:      ch,
		IngestRegistry: IngestRegistry{
			PollInterval: getDuration("INGEST_POLL_INTERVAL", 15*time.Second),
			BatchBlocks:  uint64(getInt("INGEST_BATCH_BLOCKS", 2500)),
		},
		EvaluateQuotes: EvaluateQuotes{
			BatchSize: getInt("EVAL_BATCH_SIZE", 500),
		},
		PCS: PCS{
			BaseURL:       getenv("PCS_BASE_URL", "https://api.trustedservices.intel.com"),
			APIKey:        getenv("PCS_API_KEY", ""),
			FMSPCs:        parseList(getenv("PCS_FMSPC_LIST", "")),
			POLL_INTERVAL: getDuration("PCS_POLL_INTERVAL", 24*time.Hour),
		},
	}
	return cfg, nil
}

func parseList(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	items := strings.Split(s, ",")
	out := make([]string, 0, len(items))
	for _, it := range items {
		it = strings.TrimSpace(it)
		if it != "" {
			out = append(out, it)
		}
	}
	return out
}
