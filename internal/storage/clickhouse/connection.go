package clickhouse

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
)

func Open(ctx context.Context, c *config.ClickHouse) (clickhouse.Conn, error) {
	opts := &clickhouse.Options{
		Protocol:    clickhouse.Native,
		Addr:        c.Addrs,
		DialTimeout: c.DialTimeout,
		Auth: clickhouse.Auth{
			Database: c.Database,
			Username: c.Username,
			Password: c.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: compressionMethod(c.Compression),
		},
	}
	if c.Secure {
		opts.TLS = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("open clickhouse: %w", err)
	}
	// Simple ping to verify connectivity
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping clickhouse: %w", err)
	}
	return conn, nil
}

func compressionMethod(s string) clickhouse.CompressionMethod {
	switch s {
	case "lz4", "LZ4":
		return clickhouse.CompressionLZ4
	case "zstd", "ZSTD":
		return clickhouse.CompressionZSTD
	default:
		return clickhouse.CompressionNone
	}
}
