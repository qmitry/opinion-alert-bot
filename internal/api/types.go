package api

// APIResponse is the standard Opinion API response wrapper
type APIResponse struct {
	Code   int         `json:"code"`
	Msg    string      `json:"msg"`
	Result interface{} `json:"result"`
}

// MarketDetail represents detailed information about a binary market
type MarketDetail struct {
	MarketID     string  `json:"marketId"`
	MarketTitle  string  `json:"marketTitle"`
	Status       string  `json:"status"`
	StatusEnum   int     `json:"statusEnum"`
	YesLabel     string  `json:"yesLabel"`
	NoLabel      string  `json:"noLabel"`
	YesTokenID   string  `json:"yesTokenId"`
	NoTokenID    string  `json:"noTokenId"`
	QuoteToken   string  `json:"quoteToken"`
	Volume       float64 `json:"volume"`
	Volume24h    float64 `json:"volume24h"`
	Volume7d     float64 `json:"volume7d"`
	ChainID      string  `json:"chainId"`
	ConditionID  string  `json:"conditionId"`
	CreatedAt    int64   `json:"createdAt"`
	CutoffAt     int64   `json:"cutoffAt"`
	ResolvedAt   int64   `json:"resolvedAt"`
	ResultTokenID string `json:"resultTokenId"`
}

// MarketDetailResponse wraps the market detail API response
type MarketDetailResponse struct {
	Code   int          `json:"code"`
	Msg    string       `json:"msg"`
	Result MarketDetail `json:"result"`
}

// TokenPrice represents the latest price information for a token
type TokenPrice struct {
	TokenID   string  `json:"tokenId"`
	Price     string  `json:"price"`
	Side      string  `json:"side"`
	Size      string  `json:"size"`
	Timestamp int64   `json:"timestamp"`
}

// TokenPriceResponse wraps the token price API response
type TokenPriceResponse struct {
	Code   int        `json:"code"`
	Msg    string     `json:"msg"`
	Result TokenPrice `json:"result"`
}
