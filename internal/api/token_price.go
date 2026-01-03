package api

import (
	"context"
	"fmt"
	"strconv"
)

// GetTokenPrice fetches the latest price for a specific token
func (c *Client) GetTokenPrice(ctx context.Context, tokenID string) (*TokenPrice, error) {
	path := fmt.Sprintf("/openapi/token/latest-price?token_id=%s", tokenID)

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get token price for token %s: %w", tokenID, err)
	}

	var result TokenPriceResponse
	if err := c.decodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode token price response: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("API error: code=%d, msg=%s", result.Code, result.Msg)
	}

	c.log.Debugf("Retrieved price for token %s: %s", tokenID, result.Result.Price)

	return &result.Result, nil
}

// ParseTokenPrice converts the price string to a float64
func ParseTokenPrice(price string) (float64, error) {
	p, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid price format: %w", err)
	}
	return p, nil
}

// ParseTokenSize converts the size string to a float64
func ParseTokenSize(size string) (float64, error) {
	if size == "" {
		return 0, nil
	}
	s, err := strconv.ParseFloat(size, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size format: %w", err)
	}
	return s, nil
}
