package exchange

type Market struct {
	ID              string
	Question        string
	Active          bool
	Closed          bool
	EnableOrderBook bool
	Tokens          []Token
	
	// Liquidity fields
	Liquidity       float64  // Total liquidity in USD
	Volume24h       float64  // 24h volume
	
	// Metadata
	Description     string
	EndDate         string
	OutcomePrices   string
	MarketSlug      string
	ConditionID     string
}

type Token struct {
	TokenID string
	Outcome string
	Price   float64
	Winner  bool
}

type SimplifiedMarket struct {
	Question     string
	MarketSlug   string
	ConditionID  string
	Tokens       []SimplifiedToken
	Active       bool
	Closed       bool
	EnableOrderBook bool
	
	// Add liquidity
	Liquidity    float64
	Volume24h    float64
}

type SimplifiedToken struct {
	Outcome string
	Price   float64
	Winner  bool
}

func (m *Market) IsMarketActive() bool {
	if m.Closed {
		return false
	}
	if !m.Active {
		return false
	}
	return true
}

func (m *Market) IsMarketClosed() bool {
	return m.Closed
}

// Estimate liquidity if not provided
func (m *Market) EstimateLiquidity() float64 {
	if m.Liquidity > 0 {
		return m.Liquidity
	}
	
	// Estimate from volume (liquidity ~ 15% of daily volume)
	if m.Volume24h > 0 {
		return m.Volume24h * 0.15
	}
	
	// Default conservative estimate
	return 5000.0
}

// Order book types
type OrderBook struct {
	Bids []BookLevel
	Asks []BookLevel
}

type BookLevel struct {
	Price  float64
	Size   float64
}

type Side string

const (
	SideBuy  Side = "BUY"
	SideSell Side = "SELL"
)

// Gamma API types
type GammaMarket struct {
	ConditionID   string `json:"conditionId"`
	Question      string `json:"question"`
	Slug          string `json:"slug"`
	Active        bool   `json:"active"`
	Closed        bool   `json:"closed"`
	Outcomes      string `json:"outcomes"`      // JSON array
	OutcomePrices string `json:"outcomePrices"` // JSON array
}
