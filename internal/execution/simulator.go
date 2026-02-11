package execution

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/yourusername/polymarket-arbitrage/internal/arbitrage"
	"github.com/yourusername/polymarket-arbitrage/internal/exchange"
	"go.uber.org/zap"
)

// Simulator simulates trade execution with realistic slippage and VWAP
type Simulator struct {
	client         *exchange.Client
	logger         *zap.Logger
	slippageModel  string  // "linear", "sqrt", "exponential"
	maxSlippageBps int
}

// NewSimulator creates a new execution simulator
func NewSimulator(client *exchange.Client, slippageModel string, maxSlippageBps int, logger *zap.Logger) *Simulator {
	return &Simulator{
		client:         client,
		logger:         logger,
		slippageModel:  slippageModel,
		maxSlippageBps: maxSlippageBps,
	}
}

// SimulateExecution simulates execution of an arbitrage opportunity
// Returns estimated profit after slippage and whether execution would succeed
func (s *Simulator) SimulateExecution(ctx context.Context, opp *arbitrage.Opportunity) (*SimulationResult, error) {
	result := &SimulationResult{
		OpportunityID: opp.ID,
		StartTime:     time.Now(),
	}
	
	// Simulate each leg
	totalCost := 0.0
	totalRevenue := 0.0
	maxSlippage := 0.0
	allLegsViable := true
	
	for i, leg := range opp.Legs {
		// Fetch current orderbook
		orderBook, err := s.client.GetOrderBook(ctx, leg.TokenID)
		if err != nil {
			result.Success = false
			result.FailureReason = fmt.Sprintf("failed to fetch orderbook for leg %d: %v", i, err)
			return result, nil
		}
		
		// Calculate VWAP for this leg
		vwap, err := s.client.CalculateVWAP(orderBook, leg.Side, leg.TargetSize)
		if err != nil {
			result.Success = false
			result.FailureReason = fmt.Sprintf("insufficient liquidity for leg %d: %v", i, err)
			allLegsViable = false
			break
		}
		
		// Calculate slippage
		slippage, err := s.calculateSlippage(orderBook, leg.Side, leg.TargetSize, leg.TargetPrice)
		if err != nil {
			result.Success = false
			result.FailureReason = fmt.Sprintf("slippage calculation failed for leg %d: %v", i, err)
			allLegsViable = false
			break
		}
		
		slippageBps := int(slippage * 10000)
		if slippageBps > s.maxSlippageBps {
			result.Success = false
			result.FailureReason = fmt.Sprintf("slippage too high for leg %d: %d bps (max %d)", i, slippageBps, s.maxSlippageBps)
			allLegsViable = false
			break
		}
		
		if slippage > maxSlippage {
			maxSlippage = slippage
		}
		
		// Record leg simulation
		legSim := LegSimulation{
			LegIndex:      i,
			TokenID:       leg.TokenID,
			Side:          leg.Side,
			TargetPrice:   leg.TargetPrice,
			TargetSize:    leg.TargetSize,
			EstimatedVWAP: vwap,
			Slippage:      slippage,
			SlippageBps:   slippageBps,
		}
		result.Legs = append(result.Legs, legSim)
		
		// Calculate cost/revenue for this leg
		legCost := vwap * leg.TargetSize
		if leg.Side == exchange.SideBuy {
			totalCost += legCost
		} else {
			totalRevenue += legCost
		}
		
		s.logger.Debug("Leg simulated",
			zap.Int("leg", i),
			zap.String("token", leg.TokenID),
			zap.String("side", string(leg.Side)),
			zap.Float64("target_price", leg.TargetPrice),
			zap.Float64("vwap", vwap),
			zap.Int("slippage_bps", slippageBps),
		)
	}
	
	if !allLegsViable {
		return result, nil
	}
	
	// Calculate final profit
	// For buy-both arbitrage: payout - totalCost
	// For sell-both arbitrage: totalRevenue - cost_of_minting
	var estimatedProfit float64
	
	if opp.Type == "single_buy_both" || opp.Type == "categorical_buy_all" {
		// Payout is always $1 per complete set
		payout := 1.0 * opp.Legs[0].TargetSize
		estimatedProfit = payout - totalCost
	} else if opp.Type == "single_sell_both" {
		// Need to mint pairs first (costs $1), then sell them
		mintingCost := 1.0 * opp.Legs[0].TargetSize
		estimatedProfit = totalRevenue - mintingCost
	} else {
		// Combinatorial: revenue from sells - cost of buys
		estimatedProfit = totalRevenue - totalCost
	}
	
	result.Success = true
	result.TotalCost = totalCost
	result.TotalRevenue = totalRevenue
	result.EstimatedProfit = estimatedProfit
	result.MaxSlippageBps = int(maxSlippage * 10000)
	result.ExecutionTime = time.Since(result.StartTime)
	
	// Check if still profitable after slippage
	if estimatedProfit <= 0 {
		result.Success = false
		result.FailureReason = "not profitable after slippage"
		return result, nil
	}
	
	s.logger.Info("Execution simulation complete",
		zap.String("opportunity", opp.ID),
		zap.Bool("success", result.Success),
		zap.Float64("estimated_profit", estimatedProfit),
		zap.Int("max_slippage_bps", result.MaxSlippageBps),
	)
	
	return result, nil
}

// calculateSlippage calculates slippage based on the configured model
// Models from research: linear, sqrt, exponential
func (s *Simulator) calculateSlippage(orderBook *exchange.OrderBook, side exchange.Side, targetSize, targetPrice float64) (float64, error) {
	// Calculate VWAP
	vwap, err := s.client.CalculateVWAP(orderBook, side, targetSize)
	if err != nil {
		return 0, err
	}
	
	// Base slippage (difference from target price)
	baseSlippage := math.Abs(vwap - targetPrice) / targetPrice
	
	// Apply model-specific adjustment
	switch s.slippageModel {
	case "linear":
		// Linear: slippage proportional to size
		return baseSlippage, nil
		
	case "sqrt":
		// Square root: slippage ~ sqrt(size)
		// This is the default model from research
		totalLiquidity := s.getTotalLiquidity(orderBook, side)
		if totalLiquidity == 0 {
			return baseSlippage, nil
		}
		
		sizeRatio := targetSize / totalLiquidity
		sqrtImpact := math.Sqrt(sizeRatio)
		adjustedSlippage := baseSlippage * (1.0 + sqrtImpact)
		
		return adjustedSlippage, nil
		
	case "exponential":
		// Exponential: slippage increases exponentially with size
		// Used for low-liquidity situations
		totalLiquidity := s.getTotalLiquidity(orderBook, side)
		if totalLiquidity == 0 {
			return baseSlippage, nil
		}
		
		sizeRatio := targetSize / totalLiquidity
		expImpact := math.Exp(sizeRatio) - 1.0
		adjustedSlippage := baseSlippage * (1.0 + expImpact)
		
		return adjustedSlippage, nil
		
	default:
		return baseSlippage, nil
	}
}

// getTotalLiquidity calculates total available liquidity in the orderbook
func (s *Simulator) getTotalLiquidity(orderBook *exchange.OrderBook, side exchange.Side) float64 {
	var total float64
	
	var levels []exchange.BookLevel
	if side == exchange.SideBuy {
		levels = orderBook.Asks
	} else {
		levels = orderBook.Bids
	}
	
	for _, level := range levels {
		total += level.Size
	}
	
	return total
}

// SimulationResult contains the results of an execution simulation
type SimulationResult struct {
	OpportunityID  string
	Success        bool
	FailureReason  string
	Legs           []LegSimulation
	TotalCost      float64
	TotalRevenue   float64
	EstimatedProfit float64
	MaxSlippageBps int
	StartTime      time.Time
	ExecutionTime  time.Duration
}

// LegSimulation contains simulation results for one leg
type LegSimulation struct {
	LegIndex      int
	TokenID       string
	Side          exchange.Side
	TargetPrice   float64
	TargetSize    float64
	EstimatedVWAP float64
	Slippage      float64
	SlippageBps   int
}

// CalculateKellyPosition calculates optimal position size using modified Kelly criterion
// From research: accounts for execution risk
func CalculateKellyPosition(profitBps int, executionProbability, availableLiquidity, maxCapital float64) float64 {
	// Modified Kelly: f = (b×p - q) / b × sqrt(p)
	// Where:
	//   b = profit percentage
	//   p = probability of full execution
	//   q = 1 - p
	
	b := float64(profitBps) / 10000.0
	p := executionProbability
	q := 1.0 - p
	
	if p <= 0 || b <= 0 {
		return 0
	}
	
	// Kelly fraction
	kellyFraction := (b*p - q) / b
	
	// Apply safety factor (sqrt of execution probability)
	safetyFactor := math.Sqrt(p)
	adjustedFraction := kellyFraction * safetyFactor
	
	// Cap at 50% of available liquidity to avoid moving market
	maxFraction := 0.5
	if adjustedFraction > maxFraction {
		adjustedFraction = maxFraction
	}
	
	// Calculate position size
	position := adjustedFraction * availableLiquidity
	
	// Cap at maximum capital
	if position > maxCapital {
		position = maxCapital
	}
	
	return position
}

// EstimateExecutionProbability estimates probability of successful execution
func EstimateExecutionProbability(orderBook *exchange.OrderBook, side exchange.Side, targetSize float64) float64 {
	totalLiquidity := 0.0
	
	var levels []exchange.BookLevel
	if side == exchange.SideBuy {
		levels = orderBook.Asks
	} else {
		levels = orderBook.Bids
	}
	
	for _, level := range levels {
		totalLiquidity += level.Size
	}
	
	if totalLiquidity == 0 {
		return 0
	}
	
	// Probability decreases as target size approaches total liquidity
	ratio := targetSize / totalLiquidity
	
	if ratio <= 0.1 {
		return 0.95 // Very likely
	} else if ratio <= 0.3 {
		return 0.85
	} else if ratio <= 0.5 {
		return 0.70
	} else if ratio <= 0.7 {
		return 0.50
	} else {
		return 0.30 // Risky
	}
}
