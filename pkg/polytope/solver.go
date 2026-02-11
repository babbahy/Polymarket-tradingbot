package polytope

import (
	"fmt"
)

// ConstraintSet represents the marginal polytope constraints
// From paper: Z = {z ∈ {0,1}^I : A^T × z ≥ b}
type ConstraintSet struct {
	NumVariables     int         // Number of binary outcome variables
	ConstraintMatrix [][]float64 // A^T matrix
	ConstraintVector []float64   // b vector
	ValidOutcomes    [][]int     // Valid outcome combinations
}

// NewConstraintSet creates constraints from valid outcome combinations
func NewConstraintSet(validOutcomes [][]int, numVariables int) *ConstraintSet {
	return &ConstraintSet{
		NumVariables:     numVariables,
		ValidOutcomes:    validOutcomes,
		ConstraintMatrix: make([][]float64, 0),
		ConstraintVector: make([]float64, 0),
	}
}

// CheckArbitrage checks if current prices violate the marginal polytope
// This is a simplified version - full IP solver is in full_solver.go
func (cs *ConstraintSet) CheckArbitrage(marketAPrices, marketBPrices []float64) *Solution {
	// For binary markets with dependency, check the key constraints
	
	if len(marketAPrices) != 2 || len(marketBPrices) != 2 {
		return &Solution{Exists: false}
	}
	
	pAYes := marketAPrices[0]
	pANo := marketAPrices[1]
	pBYes := marketBPrices[0]
	pBNo := marketBPrices[1]
	
	// Check each constraint type based on valid outcomes
	numValid := len(cs.ValidOutcomes)
	
	// IMPLICATION: If only 3 combinations are valid (missing [1,0]),
	// then B implies A, so P(B=YES) ≤ P(A=YES)
	if numValid == 3 {
		hasCombo := func(combo []int) bool {
			for _, v := range cs.ValidOutcomes {
				if len(v) == 2 && v[0] == combo[0] && v[1] == combo[1] {
					return true
				}
			}
			return false
		}
		
		// Check which combination is missing
		if !hasCombo([]int{1, 0}) {
			// B implies A: P(B) ≤ P(A)
			if pBYes > pAYes {
				profit := pBYes - pAYes
				profitBps := int(profit * 10000)
				
				return &Solution{
					Exists:        true,
					ProfitBps:     profitBps,
					CurrentPrices: []float64{pAYes, pANo, pBYes, pBNo},
					OptimalPrices: []float64{pBYes, 1 - pBYes, pBYes, 1 - pBYes},
					TradingActions: []TradingAction{
						{Action: "SELL", Price: pBYes},
						{Action: "BUY", Price: pAYes},
					},
					Explanation: fmt.Sprintf("Implication violation: P(B)=%.4f > P(A)=%.4f", pBYes, pAYes),
				}
			}
		}
		
		if !hasCombo([]int{0, 1}) {
			// A implies B: P(A) ≤ P(B)
			if pAYes > pBYes {
				profit := pAYes - pBYes
				profitBps := int(profit * 10000)
				
				return &Solution{
					Exists:        true,
					ProfitBps:     profitBps,
					CurrentPrices: []float64{pAYes, pANo, pBYes, pBNo},
					OptimalPrices: []float64{pAYes, 1 - pAYes, pAYes, 1 - pAYes},
					TradingActions: []TradingAction{
						{Action: "SELL", Price: pAYes},
						{Action: "BUY", Price: pBYes},
					},
					Explanation: fmt.Sprintf("Implication violation: P(A)=%.4f > P(B)=%.4f", pAYes, pBYes),
				}
			}
		}
	}
	
	// MUTUAL EXCLUSION: If [1,1] is invalid
	// then P(A=YES) + P(B=YES) ≤ 1.0
	if numValid == 3 {
		hasCombo := func(combo []int) bool {
			for _, v := range cs.ValidOutcomes {
				if len(v) == 2 && v[0] == combo[0] && v[1] == combo[1] {
					return true
				}
			}
			return false
		}
		
		if !hasCombo([]int{1, 1}) {
			sum := pAYes + pBYes
			if sum > 1.01 {
				profit := sum - 1.0
				profitBps := int(profit * 10000)
				
				return &Solution{
					Exists:        true,
					ProfitBps:     profitBps,
					CurrentPrices: []float64{pAYes, pANo, pBYes, pBNo},
					TradingActions: []TradingAction{
						{Action: "SELL", Price: pAYes},
						{Action: "SELL", Price: pBYes},
					},
					Explanation: fmt.Sprintf("Mutual exclusion violation: P(A)+P(B)=%.4f > 1.0", sum),
				}
			}
		}
	}
	
	return &Solution{Exists: false}
}

// CheckCategoricalArbitrage checks multi-outcome markets
func CheckCategoricalArbitrage(prices []float64, minProfitBps int) *Solution {
	sum := 0.0
	for _, p := range prices {
		sum += p
	}
	
	// Buy all arbitrage
	if sum < 0.99 {
		profit := 1.0 - sum
		profitBps := int(profit * 10000)
		
		if profitBps >= minProfitBps {
			actions := make([]TradingAction, len(prices))
			for i, p := range prices {
				actions[i] = TradingAction{
					Action: "BUY",
					Price:  p,
					Size:   1.0,
				}
			}
			
			return &Solution{
				Exists:         true,
				ProfitBps:      profitBps,
				CurrentPrices:  prices,
				TradingActions: actions,
				Explanation:    fmt.Sprintf("Categorical buy-all: sum=%.4f < 1.0", sum),
			}
		}
	}
	
	// Sell all arbitrage
	if sum > 1.01 {
		profit := sum - 1.0
		profitBps := int(profit * 10000)
		
		if profitBps >= minProfitBps {
			actions := make([]TradingAction, len(prices))
			for i, p := range prices {
				actions[i] = TradingAction{
					Action: "SELL",
					Price:  p,
					Size:   1.0,
				}
			}
			
			return &Solution{
				Exists:         true,
				ProfitBps:      profitBps,
				CurrentPrices:  prices,
				TradingActions: actions,
				Explanation:    fmt.Sprintf("Categorical sell-all: sum=%.4f > 1.0", sum),
			}
		}
	}
	
	return &Solution{Exists: false}
}
