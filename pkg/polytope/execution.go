package polytope

import (
	"math"
)

// ExecutionCosts represents all costs of executing a trade
type ExecutionCosts struct {
	TradingFeeBps int     // Trading fee (e.g., 20 = 2%)
	SlippageBps   int     // Market impact
	GasFeeBps     int     // Gas fees as basis points
	TotalBps      int     // Sum of all costs
}

// EstimateSlippage calculates expected slippage using square root model
// Formula: slippage = 100 * sqrt(trade_size / liquidity)
func EstimateSlippage(tradeSize, liquidity float64) float64 {
	if liquidity <= 0 {
		return 100.0 // 100% slippage if no liquidity
	}
	
	if tradeSize <= 0 {
		return 0.0
	}
	
	// Square root market impact model
	impact := math.Sqrt(tradeSize / liquidity)
	slippagePercent := impact * 100.0
	
	return slippagePercent
}

// CalculateExecutionCosts computes all costs for executing a trade
func CalculateExecutionCosts(
	tradeSize float64,
	liquidity float64,
	tradingFeeBps int,
	gasFeeUSD float64,
) ExecutionCosts {
	
	// Trading fee
	tradingCost := tradingFeeBps
	
	// Slippage
	slippagePercent := EstimateSlippage(tradeSize, liquidity)
	slippageCost := int(slippagePercent * 100) // Convert to bps
	
	// Gas fee as percentage of trade
	gasCostBps := 0
	if tradeSize > 0 {
		gasCostBps = int(gasFeeUSD / tradeSize * 10000)
	}
	
	total := tradingCost + slippageCost + gasCostBps
	
	return ExecutionCosts{
		TradingFeeBps: tradingCost,
		SlippageBps:   slippageCost,
		GasFeeBps:     gasCostBps,
		TotalBps:      total,
	}
}

// CalculateOptimalPosition finds max position size to keep slippage reasonable
// Rule: Keep slippage under 1/3 of gross profit
func CalculateOptimalPosition(arbitrageBps int, liquidity float64) float64 {
	if liquidity <= 0 {
		return 0
	}
	
	// Target: slippage = arbitrage / 3
	maxSlippagePercent := float64(arbitrageBps) / 300.0 // Divide by 300 = (3 * 100)
	
	// Solve: slippagePercent = 100 * sqrt(size / liquidity)
	// maxSlippagePercent = 100 * sqrt(size / liquidity)
	// maxSlippagePercent / 100 = sqrt(size / liquidity)
	// (maxSlippagePercent / 100)^2 = size / liquidity
	// size = liquidity * (maxSlippagePercent / 100)^2
	
	optimalSize := liquidity * math.Pow(maxSlippagePercent/100.0, 2)
	
	return optimalSize
}

// IsExecutable checks if opportunity is profitable after execution costs
func IsExecutable(
	grossProfitBps int,
	tradeSize float64,
	liquidity float64,
	tradingFeeBps int,
	gasFeeUSD float64,
	minNetProfitBps int,
) (bool, ExecutionCosts) {
	
	costs := CalculateExecutionCosts(tradeSize, liquidity, tradingFeeBps, gasFeeUSD)
	netProfitBps := grossProfitBps - costs.TotalBps
	
	return netProfitBps >= minNetProfitBps, costs
}
