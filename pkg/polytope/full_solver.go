package polytope

import (
	"fmt"
	"math"
)

type FullIPSolver struct {
	NumMarkets     int
	NumOutcomes    int
	ValidOutcomes  [][]int
	Tolerance      float64
	MaxIterations  int
}

func NewFullIPSolver(validOutcomes [][]int, numMarkets int) *FullIPSolver {
	return &FullIPSolver{
		NumMarkets:    numMarkets,
		NumOutcomes:   len(validOutcomes[0]) * 2,
		ValidOutcomes: validOutcomes,
		Tolerance:     1e-6,
		MaxIterations: 100,
	}
}

func (s *FullIPSolver) CheckViolation(prices []float64) (*ArbitrageOpportunity, error) {
	polytope := s.buildMarginalPolytope()
	
	if s.isInPolytope(prices, polytope) {
		return nil, nil
	}
	
	projected := s.bregmanProjection(prices, polytope)
	divergence := s.klDivergence(projected, prices)
	profitBps := int(divergence * 10000)
	
	if profitBps >= 10 {
		return &ArbitrageOpportunity{
			Type:          "combinatorial",
			CurrentPrices: prices,
			OptimalPrices: projected,
			Profit:        divergence,
			ProfitBps:     profitBps,
			Strategy:      s.computeStrategy(prices, projected),
			Explanation:   fmt.Sprintf("Marginal polytope violation: KL=%.4f", divergence),
		}, nil
	}
	
	return nil, nil
}

func (s *FullIPSolver) buildMarginalPolytope() *Polytope {
	vertices := make([][]float64, len(s.ValidOutcomes))
	
	for i, outcome := range s.ValidOutcomes {
		vertex := make([]float64, len(outcome)*2)
		
		for j, val := range outcome {
			if val == 1 {
				vertex[j*2] = 1.0
				vertex[j*2+1] = 0.0
			} else {
				vertex[j*2] = 0.0
				vertex[j*2+1] = 1.0
			}
		}
		
		vertices[i] = vertex
	}
	
	return &Polytope{
		Vertices:  vertices,
		Dimension: len(vertices[0]),
	}
}

func (s *FullIPSolver) isInPolytope(point []float64, polytope *Polytope) bool {
	for _, vertex := range polytope.Vertices {
		dist := s.euclideanDistance(point, vertex)
		if dist < s.Tolerance {
			return true
		}
	}
	
	projected := s.bregmanProjection(point, polytope)
	dist := s.euclideanDistance(point, projected)
	
	return dist < s.Tolerance
}

func (s *FullIPSolver) bregmanProjection(prices []float64, polytope *Polytope) []float64 {
	mu := s.initializeProjection(prices, polytope)
	
	for iter := 0; iter < s.MaxIterations; iter++ {
		gradient := s.computeGradient(mu)
		vertex := s.findDescentVertex(gradient, polytope)
		stepSize := s.lineSearch(mu, vertex, gradient)
		
		muNext := make([]float64, len(mu))
		for i := range mu {
			muNext[i] = (1-stepSize)*mu[i] + stepSize*vertex[i]
		}
		
		gap := s.dualityGap(mu, vertex, gradient)
		mu = muNext
		
		if gap < s.Tolerance {
			break
		}
	}
	
	return mu
}

func (s *FullIPSolver) initializeProjection(prices []float64, polytope *Polytope) []float64 {
	minDist := math.Inf(1)
	var closest []float64
	
	for _, vertex := range polytope.Vertices {
		dist := s.klDivergence(vertex, prices)
		if dist < minDist {
			minDist = dist
			closest = vertex
		}
	}
	
	return closest
}

func (s *FullIPSolver) computeGradient(mu []float64) []float64 {
	gradient := make([]float64, len(mu))
	
	for i := range mu {
		if mu[i] > 1e-10 {
			gradient[i] = math.Log(mu[i]) + 1
		} else {
			gradient[i] = -100.0
		}
	}
	
	return gradient
}

func (s *FullIPSolver) findDescentVertex(gradient []float64, polytope *Polytope) []float64 {
	minValue := math.Inf(1)
	var bestVertex []float64
	
	for _, vertex := range polytope.Vertices {
		value := 0.0
		for i := range gradient {
			value += gradient[i] * vertex[i]
		}
		
		if value < minValue {
			minValue = value
			bestVertex = vertex
		}
	}
	
	return bestVertex
}

func (s *FullIPSolver) lineSearch(current, vertex, gradient []float64) float64 {
	alpha := 1.0
	beta := 0.5
	sigma := 0.1
	maxIter := 20
	
	direction := make([]float64, len(current))
	for i := range direction {
		direction[i] = vertex[i] - current[i]
	}
	
	dirDeriv := 0.0
	for i := range gradient {
		dirDeriv += gradient[i] * direction[i]
	}
	
	currentObj := s.negativeEntropy(current)
	
	for iter := 0; iter < maxIter; iter++ {
		test := make([]float64, len(current))
		for i := range test {
			test[i] = current[i] + alpha*direction[i]
			if test[i] < 0 {
				test[i] = 1e-10
			}
		}
		
		testObj := s.negativeEntropy(test)
		if testObj <= currentObj + sigma*alpha*dirDeriv {
			return alpha
		}
		
		alpha *= beta
	}
	
	return alpha
}

func (s *FullIPSolver) negativeEntropy(p []float64) float64 {
	entropy := 0.0
	for _, pi := range p {
		if pi > 1e-10 {
			entropy -= pi * math.Log(pi)
		}
	}
	return entropy
}

func (s *FullIPSolver) dualityGap(mu, vertex, gradient []float64) float64 {
	gap := 0.0
	for i := range mu {
		gap += gradient[i] * (mu[i] - vertex[i])
	}
	return math.Abs(gap)
}

func (s *FullIPSolver) klDivergence(p, q []float64) float64 {
	divergence := 0.0
	
	for i := range p {
		if p[i] > 1e-10 && q[i] > 1e-10 {
			divergence += p[i] * math.Log(p[i]/q[i])
		}
	}
	
	return divergence
}

func (s *FullIPSolver) euclideanDistance(a, b []float64) float64 {
	sum := 0.0
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

func (s *FullIPSolver) computeStrategy(current, optimal []float64) []TradeAction {
	actions := make([]TradeAction, 0)
	
	for i := range current {
		diff := optimal[i] - current[i]
		
		if math.Abs(diff) > 0.005 {
			action := TradeAction{
				OutcomeIndex: i,
				CurrentPrice: current[i],
				TargetPrice:  optimal[i],
				Size:         math.Abs(diff),
			}
			
			if diff > 0 {
				action.Side = "BUY"
			} else {
				action.Side = "SELL"
			}
			
			actions = append(actions, action)
		}
	}
	
	return actions
}
