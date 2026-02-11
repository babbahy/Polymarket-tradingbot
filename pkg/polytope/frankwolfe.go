package polytope

import (
	"math"
)

// FrankWolfe implements the Frank-Wolfe algorithm for Bregman projection
// From paper: Iteratively solve convex optimization over growing active set
type FrankWolfe struct {
	MaxIterations   int
	ConvergenceThreshold float64
	Epsilon         float64 // Contraction parameter
}

// Iterate performs Frank-Wolfe iterations
func (fw *FrankWolfe) Iterate(currentPrices []float64, constraints *ConstraintSet, maxIter int) *BregmanProjection {
	// Simplified Frank-Wolfe
	// Full implementation would:
	// 1. Start with active set Z_0
	// 2. Each iteration: solve convex opt, find descent vertex via IP
	// 3. Add to active set, check convergence
	
	// For now, return the constraint-based projection
	return ComputeProjection(currentPrices, constraints)
}

// convergenceGap computes g(μ_t) = ∇F(μ_t)·(μ_t - z_t)
func convergenceGap(mu, z, gradient []float64) float64 {
	gap := 0.0
	for i := range mu {
		gap += gradient[i] * (mu[i] - z[i])
	}
	return gap
}

// computeGradient computes gradient of negative entropy
// ∇R(μ) = ln(μ) + 1
func computeGradient(mu []float64) []float64 {
	gradient := make([]float64, len(mu))
	for i, mi := range mu {
		if mi > 0 {
			gradient[i] = math.Log(mi) + 1
		} else {
			gradient[i] = -100 // Large negative for zero probabilities
		}
	}
	return gradient
}
