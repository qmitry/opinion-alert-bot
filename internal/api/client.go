package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// Client represents the Opinion API HTTP client
type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
	log        *logrus.Logger
}

// NewClient creates a new Opinion API client
func NewClient(apiKey, baseURL string, log *logrus.Logger) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		apiKey:  apiKey,
		baseURL: baseURL,
		log:     log,
	}
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication header
	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	c.log.Debugf("Making %s request to %s", method, url)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

// decodeResponse reads and decodes a JSON response
func (c *Client) decodeResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	c.log.Debugf("Response body: %s", string(bodyBytes))

	if err := json.Unmarshal(bodyBytes, v); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	return nil
}
