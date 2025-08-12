package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	clickdb "github.com/MoeMahhouk/go-tcb-notify/internal/storage/clickhouse"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ch, err := clickdb.Open(ctx, &cfg.ClickHouse)
	if err != nil {
		log.Fatalf("clickhouse: %v", err)
	}

	client := &http.Client{Timeout: 20 * time.Second}

	interval := cfg.PCS.POLL_INTERVAL

	log.Printf("[pcs] starting poller interval=%s base=%s", interval, cfg.PCS.BaseURL)

	for {
		fmspcs := cfg.PCS.FMSPCs
		if len(fmspcs) == 0 {
			// Best-effort discovery; some deployments require 'all'
			list, err := fetchFMSPCList(ctx, client, cfg, "all")
			if err != nil {
				log.Printf("WARN discover fmspcs: %v", err)
			} else {
				fmspcs = list
			}
		}

		for _, fmspc := range fmspcs {
			if err := fetchAndStore(ctx, client, ch, cfg, fmspc); err != nil {
				log.Printf("WARN fetch FMSPC %s: %v", fmspc, err)
			} else {
				log.Printf("[pcs] updated TCB info for %s", fmspc)
			}
		}

		time.Sleep(interval)
	}
}

func fetchFMSPCList(ctx context.Context, httpc *http.Client, cfg *config.Config, platform string) ([]string, error) {
	u, err := url.Parse(cfg.PCS.BaseURL)
	if err != nil {
		return nil, err
	}
	u.Path = "/sgx/certification/v4/fmspcs"
	q := u.Query()
	q.Set("platform", platform) // "TDX"
	u.RawQuery = q.Encode()

	req, _ := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if cfg.PCS.APIKey != "" {
		req.Header.Set("Ocp-Apim-Subscription-Key", cfg.PCS.APIKey)
	}
	resp, err := httpc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fmspcs request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fmspcs %s -> %d: %s", u.String(), resp.StatusCode, string(b))
	}

	b, _ := io.ReadAll(resp.Body)

	// Preferred: bare array [{ "fmspc": "...", "platform": "TDX" }, ...]
	type arrItem struct {
		FMSPC    string `json:"fmspc"`
		Platform string `json:"platform"`
	}
	var arr []arrItem
	if err := json.Unmarshal(b, &arr); err == nil && len(arr) > 0 {
		out := make([]string, 0, len(arr))
		for _, it := range arr {
			if it.FMSPC == "" {
				continue
			}
			if platform == "all" || strings.EqualFold(it.Platform, platform) {
				out = append(out, it.FMSPC)
			}
		}
		if len(out) > 0 {
			return out, nil
		}
	}

	// Fallback: wrapper object { "fmspcs": [{ "fmspc": "..." }, ...] }
	var wrap struct {
		FMSPCs []struct {
			FMSPC string `json:"fmspc"`
		} `json:"fmspcs"`
	}
	if err := json.Unmarshal(b, &wrap); err == nil && len(wrap.FMSPCs) > 0 {
		out := make([]string, 0, len(wrap.FMSPCs))
		for _, e := range wrap.FMSPCs {
			if e.FMSPC != "" {
				out = append(out, e.FMSPC)
			}
		}
		return out, nil
	}

	return nil, fmt.Errorf("unexpected fmspcs JSON: %s", string(b))
}

func fetchAndStore(ctx context.Context, httpc *http.Client, ch clickhouse.Conn, cfg *config.Config, fmspc string) error {
	// TDX Get TCB Info: /tdx/certification/v4/tcb?fmspc=<hex>
	u, err := url.Parse(cfg.PCS.BaseURL)
	if err != nil {
		return err
	}
	u.Path = "/tdx/certification/v4/tcb"
	q := u.Query()
	q.Set("fmspc", fmspc)
	u.RawQuery = q.Encode()

	req, _ := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if cfg.PCS.APIKey != "" {
		req.Header.Set("Ocp-Apim-Subscription-Key", cfg.PCS.APIKey)
	}
	resp, err := httpc.Do(req)
	if err != nil {
		return fmt.Errorf("pcs request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("pcs %s -> %d: %s", u.String(), resp.StatusCode, string(b))
	}
	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return fmt.Errorf("decode pcs json: %w", err)
	}

	// Extract a couple of fields for convenience; keep raw_json
	var issueDate, nextUpdate string
	var tcbEvalNum float64
	if tcbInfo, ok := raw["tcbInfo"].(map[string]any); ok {
		if v, ok := tcbInfo["issueDate"].(string); ok {
			issueDate = v
		}
		if v, ok := tcbInfo["nextUpdate"].(string); ok {
			nextUpdate = v
		}
		if v, ok := tcbInfo["tcbEvaluationDataNumber"].(float64); ok {
			tcbEvalNum = v
		}
	}
	rawJSON, _ := json.Marshal(raw)

	if issueDate == "" {
		issueDate = "1970-01-01"
	}
	if nextUpdate == "" {
		nextUpdate = "1970-01-01"
	}
	return ch.Exec(ctx, clickdb.UpsertPCSTCBInfo,
		fmspc,
		uint32(tcbEvalNum),
		issueDate,
		nextUpdate,
		string(rawJSON),
	)
}
