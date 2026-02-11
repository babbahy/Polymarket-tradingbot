package polytope

// Solution represents an arbitrage opportunity found by the solver
type Solution struct {
	Exists         bool
	ProfitBps      int
	OptimalPrices  []float64
	CurrentPrices  []float64
	TradingActions []TradingAction
	Explanation    string
}

// TradingAction represents a buy/sell action
type TradingAction struct {
	Market       string
	TokenID      string
	Action       string  // "BUY" or "SELL"
	Size         float64
	Price        float64
	OutcomeIndex int
	CurrentPrice float64
	TargetPrice  float64
}

// ArbitrageOpportunity represents a detected combinatorial arbitrage
type ArbitrageOpportunity struct {
	Type          string        // "combinatorial"
	CurrentPrices []float64     // Current market prices θ
	OptimalPrices []float64     // Projected prices μ*
	Profit        float64       // KL divergence D(μ*||θ)
	ProfitBps     int           // Profit in basis points
	Strategy      []TradeAction // Optimal trading actions
	Explanation   string        // Human-readable explanation
}

// TradeAction represents a single trade to execute
type TradeAction struct {
	OutcomeIndex int     // Which outcome (0=Market1_YES, 1=Market1_NO, ...)
	Side         string  // "BUY" or "SELL"
	CurrentPrice float64 // Current market price
	TargetPrice  float64 // Target price after arbitrage
	Size         float64 // Trade size
}

// Polytope represents the marginal polytope M(Z)
type Polytope struct {
	Vertices  [][]float64 // Valid outcome vertices
	Dimension int         // Number of outcome probabilities
}
