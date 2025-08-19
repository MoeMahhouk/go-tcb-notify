package pcs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
	pcstypes "github.com/MoeMahhouk/go-tcb-notify/pkg/pcs"
)

// Client handles communication with Intel PCS API
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Intel PCS API client
func NewClient(baseURL, apiKey string) *Client {
	if baseURL == "" {
		baseURL = "https://api.trustedservices.intel.com"
	}

	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetFMSPCs fetches all FMSPCs from Intel PCS
func (c *Client) GetFMSPCs(ctx context.Context, platform string) ([]models.FMSPCResponse, error) {
	url := fmt.Sprintf("%s/sgx/certification/v4/fmspcs?platform=%s", c.baseURL, platform)

	body, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch FMSPCs: %w", err)
	}

	var fmspcs []models.FMSPCResponse
	if err := json.Unmarshal(body, &fmspcs); err != nil {
		return nil, fmt.Errorf("parse FMSPC list: %w", err)
	}

	return fmspcs, nil
}

// GetTCBInfo fetches TCB info for a specific FMSPC
func (c *Client) GetTCBInfo(ctx context.Context, fmspc string) (*pcstypes.TCBInfoResponse, error) {
	url := fmt.Sprintf("%s/tdx/certification/v4/tcb?fmspc=%s", c.baseURL, fmspc)

	body, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var tcbResp pcstypes.TCBInfoResponse
	if err := json.Unmarshal(body, &tcbResp); err != nil {
		return nil, fmt.Errorf("parse TCB info: %w", err)
	}

	return &tcbResp, nil
}

// doRequest performs an HTTP request to the Intel PCS API
func (c *Client) doRequest(ctx context.Context, method, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Add API key if configured
	if c.apiKey != "" {
		req.Header.Set("Ocp-Apim-Subscription-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	// Handle special status codes
	switch resp.StatusCode {
	case http.StatusOK:
		// Continue processing
	case http.StatusNotFound:
		return nil, ErrNotFound
	default:
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return respBody, nil
}
