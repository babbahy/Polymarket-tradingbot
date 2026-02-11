package main

import (
	"context"
	"time"

	"github.com/yourusername/polymarket-arbitrage/internal/arbitrage"
	"github.com/yourusername/polymarket-arbitrage/internal/exchange"
	"github.com/yourusername/polymarket-arbitrage/pkg/polytope"
	"go.uber.org/zap"
)

// CombinatorialScanner handles AI-powered combinatorial arbitrage detection
type CombinatorialScanner struct {
	aiDetector *arbitrage.AIDetector
	logger     *zap.Logger
	stats      *ScanStats
}

// ScanStats tracks detection statistics
type ScanStats struct {
	PairsAnalyzed      int
	DependenciesFound  int
	ViolationsFound    int
	ArbitrageFound     int
	TotalProfitBps     int
}

// NewCombinatorialScanner creates a new scanner
func NewCombinatorialScanner(aiDetector *arbitrage.AIDetector, logger *zap.Logger) *CombinatorialScanner {
	return &CombinatorialScanner{
		aiDetector: aiDetector,
		logger:     logger,
		stats:      &ScanStats{},
	}
}

// ScanMarkets performs combinatorial arbitrage detection
func (cs *CombinatorialScanner) ScanMarkets(ctx context.Context, markets []exchange.Market) []ArbitrageResult {
	results := make([]ArbitrageResult, 0)
	
	// Filter to active binary markets
	binaryMarkets := filterBinaryMarkets(markets)
	
	if len(binaryMarkets) < 2 {
		cs.logger.Info("Not enough binary markets for combinatorial analysis")
		return results
	}
	
	cs.logger.Info("🤖 Starting combinatorial arbitrage scan",
		zap.Int("markets", len(binaryMarkets)),
	)
	
	// Analyze pairs (limit to top markets by volume if needed)
	maxPairs := 20 // Limit for API cost control
	pairsChecked := 0
	
	for i := 0; i < len(binaryMarkets) && pairsChecked < maxPairs; i++ {
		for j := i + 1; j < len(binaryMarkets) && pairsChecked < maxPairs; j++ {
			market1 := &binaryMarkets[i]
			market2 := &binaryMarkets[j]
			
			// Check with timeout
			analyzeCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			result := cs.analyzePair(analyzeCtx, market1, market2)
			cancel()
			
			pairsChecked++
			cs.stats.PairsAnalyzed++
			
			if result != nil {
				results = append(results, *result)
			}
		}
	}
	
	cs.logger.Info("🤖 Combinatorial scan complete",
		zap.Int("pairs_analyzed", cs.stats.PairsAnalyzed),
		zap.Int("dependencies_found", cs.stats.DependenciesFound),
		zap.Int("arbitrage_found", cs.stats.ArbitrageFound),
	)
	
	return results
}

// analyzePair analyzes a single market pair
func (cs *CombinatorialScanner) analyzePair(ctx context.Context, market1, market2 *exchange.Market) *ArbitrageResult {
	// Step 1: AI detects dependency
	analysis, err := cs.aiDetector.AnalyzeMarketPair(ctx, market1, market2)
	if err != nil {
		cs.logger.Debug("AI analysis failed",
			zap.Error(err),
		)
		return nil
	}
	
	// Must have high-confidence dependency
	if !analysis.HasDependency || analysis.Confidence < 0.75 {
		return nil
	}
	
	cs.stats.DependenciesFound++
	
	cs.logger.Info("🤖 AI detected dependency",
		zap.String("market1", market1.Question),
		zap.String("market2", market2.Question),
		zap.String("type", analysis.DependencyType),
		zap.Float64("confidence", analysis.Confidence),
	)
	
	// Step 2: Mathematical validation
	opportunity := cs.checkViolation(market1, market2, analysis)
	
	if opportunity != nil {
		cs.stats.ArbitrageFound++
		cs.stats.TotalProfitBps += opportunity.ProfitBps
		
		cs.logger.Info("🎯🤖💰 COMBINATORIAL ARBITRAGE FOUND!",
			zap.String("market1", market1.Question),
			zap.String("market2", market2.Question),
			zap.Int("profit_bps", opportunity.ProfitBps),
			zap.String("explanation", opportunity.Explanation),
		)
		
		return &ArbitrageResult{
			Type:        "combinatorial",
			Market1:     market1.Question,
			Market2:     market2.Question,
			Dependency:  analysis.DependencyType,
			Confidence:  analysis.Confidence,
			ProfitBps:   opportunity.ProfitBps,
			Explanation: opportunity.Explanation,
			Opportunity: opportunity,
		}
	}
	
	return nil
}

// checkViolation uses full IP solver to check for violations
func (cs *CombinatorialScanner) checkViolation(market1, market2 *exchange.Market, analysis *arbitrage.DependencyAnalysis) *polytope.ArbitrageOpportunity {
	// Create full IP solver
	solver := polytope.NewFullIPSolver(analysis.ValidCombinations, 2)
	
	// Combine prices into single vector [A_yes, A_no, B_yes, B_no]
	prices := make([]float64, 4)
	prices[0] = market1.Tokens[0].Price // Market1 YES
	prices[1] = market1.Tokens[1].Price // Market1 NO
	prices[2] = market2.Tokens[0].Price // Market2 YES
	prices[3] = market2.Tokens[1].Price // Market2 NO
	
	// Check for marginal polytope violation
	opportunity, err := solver.CheckViolation(prices)
	
	if err != nil {
		cs.logger.Error("Solver error", zap.Error(err))
		return nil
	}
	
	return opportunity
}

// ArbitrageResult represents a detected opportunity
type ArbitrageResult struct {
	Type        string
	Market1     string
	Market2     string
	Dependency  string
	Confidence  float64
	ProfitBps   int
	Explanation string
	Opportunity *polytope.ArbitrageOpportunity
}

// filterBinaryMarkets filters to active binary markets
func filterBinaryMarkets(markets []exchange.Market) []exchange.Market {
	filtered := make([]exchange.Market, 0)
	
	for _, m := range markets {
		if m.IsMarketActive() && !m.IsMarketClosed() && len(m.Tokens) == 2 {
			filtered = append(filtered, m)
		}
	}
	
	return filtered
}
