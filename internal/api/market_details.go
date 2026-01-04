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

	if result.Code != 0 || result.Errno != 0 {
		return nil, fmt.Errorf("API error: code=%d, errno=%d, msg=%s", result.Code, result.Errno, result.Msg)
	}

	marketData := result.Result.Data

	c.log.Debugf("Decoded market details - Code: %d, Errno: %d, Msg: %s", result.Code, result.Errno, result.Msg)
	c.log.Debugf("Market fields - ID: %d, Title: %s, YesTokenID: %s, NoTokenID: %s, Status: %d",
		marketData.MarketID, marketData.MarketTitle, marketData.YesTokenID,
		marketData.NoTokenID, marketData.Status)

	// Validate required fields
	if marketData.YesTokenID == "" {
		return nil, fmt.Errorf("market %s has no YES token ID - market may not be active or response structure is incorrect", marketID)
	}

	return &marketData, nil
}
