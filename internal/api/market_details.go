package api

import (
	"context"
	"fmt"
)

// GetMarketDetails fetches detailed information about a specific binary market
func (c *Client) GetMarketDetails(ctx context.Context, marketID string) (*MarketDetail, error) {
	path := fmt.Sprintf("/openapi/market/%s", marketID)

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get market details for market %s: %w", marketID, err)
	}

	var result MarketDetailResponse
	if err := c.decodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode market details response: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("API error: code=%d, msg=%s", result.Code, result.Msg)
	}

	c.log.Debugf("Retrieved market details for market %s: %s", marketID, result.Result.MarketTitle)

	return &result.Result, nil
}
