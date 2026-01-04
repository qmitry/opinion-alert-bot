package api

import (
	"context"
	"fmt"
)

// GetMarketDetails fetches detailed information about a specific market (binary or multi-outcome)
// For multi-outcome markets, it automatically uses the first child market's token
func (c *Client) GetMarketDetails(ctx context.Context, marketID string) (*MarketDetail, error) {
	// Try binary market endpoint first
	path := fmt.Sprintf("/openapi/market/%s", marketID)
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get market details for market %s: %w", marketID, err)
	}

	var result MarketDetailResponse
	if err := c.decodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode market details response: %w", err)
	}

	// If errno=10200, it might be a categorical market - try that endpoint
	if result.Errno == 10200 {
		c.log.Debugf("Market %s not found as binary, trying categorical endpoint...", marketID)
		categoricalPath := fmt.Sprintf("/openapi/market/categorical/%s", marketID)
		resp, err = c.doRequest(ctx, "GET", categoricalPath, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get categorical market details for market %s: %w", marketID, err)
		}

		if err := c.decodeResponse(resp, &result); err != nil {
			return nil, fmt.Errorf("failed to decode categorical market details response: %w", err)
		}
	}

	// Check for API errors
	if result.Code != 0 || result.Errno != 0 {
		return nil, fmt.Errorf("API error: code=%d, errno=%d, msg=%s", result.Code, result.Errno, result.Msg)
	}

	marketData := result.Result.Data

	c.log.Debugf("Decoded market details - Code: %d, Errno: %d, Msg: %s", result.Code, result.Errno, result.Msg)
	c.log.Debugf("Market fields - ID: %d, Title: %s, YesTokenID: %s, NoTokenID: %s, MarketType: %d, Status: %d, ChildMarkets: %d",
		marketData.MarketID, marketData.MarketTitle, marketData.YesTokenID,
		marketData.NoTokenID, marketData.MarketType, marketData.Status, len(marketData.ChildMarkets))

	// Validate market is active
	if marketData.Status == 0 {
		return nil, fmt.Errorf("market %s is not active", marketID)
	}

	// For multi-outcome markets (marketType 1), use the first child market's token
	if marketData.MarketType == 1 {
		if len(marketData.ChildMarkets) == 0 {
			return nil, fmt.Errorf("multi-outcome market %s has no child markets", marketID)
		}
		// Use the first child market's YES token ID
		firstChild := marketData.ChildMarkets[0]
		if firstChild.YesTokenID == "" {
			return nil, fmt.Errorf("multi-outcome market %s first child has no YES token ID", marketID)
		}
		// Override the parent market's YesTokenID with the first child's token
		marketData.YesTokenID = firstChild.YesTokenID
		c.log.Infof("Multi-outcome market %s: automatically selected first outcome token %s (%s)",
			marketID, firstChild.YesTokenID, firstChild.MarketTitle)
	} else if marketData.MarketType == 0 {
		// Binary market - require YES/NO tokens
		if marketData.YesTokenID == "" {
			return nil, fmt.Errorf("binary market %s has no YES token ID - market may not be properly configured", marketID)
		}
	}

	return &marketData, nil
}
