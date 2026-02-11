package polytope

import (
	"math"
)

// BregmanProjection computes the Bregman projection onto the marginal polytope
// From paper: μ* = argmin_{μ ∈ M} D(μ||θ)
// Where D is the Bregman divergence (KL divergence for LMSR)
type BregmanProjection struct {
	CurrentPrices  []float64
	ProjectedPrices []float64
	Divergence     float64
	MaxProfit      float64
}

// ComputeProjection calculates the Bregman projection
// This is a simplified version - full implementation would use Frank-Wolfe
func ComputeProjection(currentPrices []float64, constraints *ConstraintSet) *BregmanProjection {
	// For now, use the constraint-based projection
	// Full Frank-Wolfe would iteratively solve this
	
	projected := make([]float64, len(currentPrices))
	copy(projected, currentPrices)
	
	// Normalize to sum to 1.0 (basic projection)
	sum := 0.0
	for _, p := range projected {
		sum += p
	}
	
	if math.Abs(sum-1.0) > 0.01 {
		for i := range projected {
			projected[i] = projected[i] / sum
		}
	}
	
	// Calculate KL divergence
	divergence := KLDivergence(projected, currentPrices)
	
	// Maximum profit is the divergence (from paper)
	maxProfit := divergence
	
	return &BregmanProjection{
		CurrentPrices:   currentPrices,
		ProjectedPrices: projected,
		Divergence:      divergence,
		MaxProfit:       maxProfit,
	}
}

// KLDivergence computes Kullback-Leibler divergence
// D_KL(P||Q) = sum_i P_i * log(P_i / Q_i)
func KLDivergence(p, q []float64) float64 {
	if len(p) != len(q) {
		return 0
	}
	
	divergence := 0.0
	for i := range p {
		if p[i] > 0 && q[i] > 0 {
			divergence += p[i] * math.Log(p[i]/q[i])
		}
	}
	
	return divergence
}

// NegativeEntropy computes -sum(p_i * log(p_i))
// This is R(μ) in the paper's notation
func NegativeEntropy(p []float64) float64 {
	entropy := 0.0
	for _, pi := range p {
		if pi > 0 {
			entropy += pi * math.Log(pi)
		}
	}
	return entropy
}
