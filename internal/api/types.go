package api

// APIResponse is the standard Opinion API response wrapper
type APIResponse struct {
	Code   int         `json:"code"`
	Msg    string      `json:"msg"`
	Result interface{} `json:"result"`
}

// MarketDetail represents detailed information about a binary or multi-outcome market
type MarketDetail struct {
	MarketID      int            `json:"marketId"`
	MarketTitle   string         `json:"marketTitle"`
	Status        int            `json:"status"`
	StatusEnum    string         `json:"statusEnum"`
	MarketType    int            `json:"marketType"` // 0=Binary, 1=Categorical/Multi-outcome
	YesLabel      string         `json:"yesLabel"`
	NoLabel       string         `json:"noLabel"`
	YesTokenID    string         `json:"yesTokenId"`
	NoTokenID     string         `json:"noTokenId"`
	QuoteToken    string         `json:"quoteToken"`
	Volume        string         `json:"volume"`
	Volume24h     string         `json:"volume24h"`
	Volume7d      string         `json:"volume7d"`
	ChainID       string         `json:"chainId"`
	ConditionID   string         `json:"conditionId"`
	CreatedAt     int64          `json:"createdAt"`
	CutoffAt      int64          `json:"cutoffAt"`
	ResolvedAt    int64          `json:"resolvedAt"`
	ResultTokenID string         `json:"resultTokenId"`
	ChildMarkets  []MarketDetail `json:"childMarkets"` // For multi-outcome markets
}

// MarketDetailResult wraps the data field in the result
type MarketDetailResult struct {
	Data MarketDetail `json:"data"`
}

// MarketDetailResponse wraps the market detail API response
type MarketDetailResponse struct {
	Code   int                `json:"code"`
	Msg    string             `json:"msg"`
	Errno  int                `json:"errno"`
	Result MarketDetailResult `json:"result"`
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
