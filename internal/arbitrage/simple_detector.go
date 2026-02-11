package arbitrage

import (
	"time"
)

type SimpleDetector struct{}

func NewSimpleDetector() *SimpleDetector {
	return &SimpleDetector{}
}

func (d *SimpleDetector) DetectBinaryArbitrage(prices []float64, minProfitBps int) *Opportunity {
	if len(prices) != 2 {
		return nil
	}

	sum := prices[0] + prices[1]
	
	// Buy arbitrage
	if sum < 0.99 {
		profit := 1.0 - sum
		profitBps := int(profit * 10000)
		
		if profitBps >= minProfitBps {
			return &Opportunity{
				Type:      OpportunityTypeBinary,
				ProfitBps: profitBps,
				Timestamp: time.Now().Unix(),
				Prices:    prices,
				Strategy:  "BUY_BOTH",
			}
		}
	}
	
	// Sell arbitrage
	if sum > 1.01 {
		profit := sum - 1.0
		profitBps := int(profit * 10000)
		
		if profitBps >= minProfitBps {
			return &Opportunity{
				Type:      OpportunityTypeBinary,
				ProfitBps: profitBps,
				Timestamp: time.Now().Unix(),
				Prices:    prices,
				Strategy:  "SELL_BOTH",
			}
		}
	}
	
	return nil
}

func (d *SimpleDetector) DetectCategoricalArbitrage(prices []float64, minProfitBps int) *Opportunity {
	if len(prices) < 2 {
		return nil
	}

	sum := 0.0
	for _, p := range prices {
		sum += p
	}
	
	// Buy arbitrage
	if sum < 0.99 {
		profit := 1.0 - sum
		profitBps := int(profit * 10000)
		
		if profitBps >= minProfitBps {
			return &Opportunity{
				Type:      OpportunityTypeCategorical,
				ProfitBps: profitBps,
				Timestamp: time.Now().Unix(),
				Prices:    prices,
				Strategy:  "BUY_ALL",
			}
		}
	}
	
	// Sell arbitrage
	if sum > 1.01 {
		profit := sum - 1.0
		profitBps := int(profit * 10000)
		
		if profitBps >= minProfitBps {
			return &Opportunity{
				Type:      OpportunityTypeCategorical,
				ProfitBps: profitBps,
				Timestamp: time.Now().Unix(),
				Prices:    prices,
				Strategy:  "SELL_ALL",
			}
		}
	}
	
	return nil
}
