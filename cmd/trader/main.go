package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/polymarket-arbitrage/internal/arbitrage"
	"github.com/yourusername/polymarket-arbitrage/internal/config"
	"github.com/yourusername/polymarket-arbitrage/internal/exchange"
	"github.com/yourusername/polymarket-arbitrage/internal/logger"
	"github.com/yourusername/polymarket-arbitrage/internal/metrics"
	"github.com/yourusername/polymarket-arbitrage/pkg/polytope"
	"go.uber.org/zap"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "Path to configuration file")
	mode := flag.String("mode", "", "Trading mode override")
	mockMode := flag.Bool("mock", false, "Use mock data")
	anthropicKey := flag.String("anthropic-key", os.Getenv("ANTHROPIC_API_KEY"), "Anthropic API key")
	enableAI := flag.Bool("enable-ai", false, "Enable AI-powered combinatorial detection")
	showVersion := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Polymarket Arbitrage Bot\nVersion: %s\nBuild Time: %s\n", Version, BuildTime)
		os.Exit(0)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if *mode != "" {
		cfg.Mode = *mode
		if err := cfg.Validate(); err != nil {
			fmt.Fprintf(os.Stderr, "Invalid mode: %v\n", err)
			os.Exit(1)
		}
	}

	logConfig := logger.Config{
		Level:             cfg.Logging.Level,
		Format:            cfg.Logging.Format,
		Output:            cfg.Logging.Output,
		IncludeCaller:     cfg.Logging.IncludeCaller,
		IncludeStacktrace: cfg.Logging.IncludeStacktrace,
	}
	if err := logger.Initialize(logConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	log := logger.Log

	log.Info("🚀 Starting AGENTIC Polymarket Arbitrage Bot",
		zap.String("version", Version),
		zap.String("mode", cfg.Mode),
		zap.Bool("mock", *mockMode),
		zap.Bool("ai_enabled", *enableAI && *anthropicKey != ""),
	)

	if *enableAI && *anthropicKey == "" {
		log.Warn("⚠️  AI disabled - no Anthropic API key")
		*enableAI = false
	}

	var metricsServer *metrics.Server
	if cfg.Metrics.Enabled {
		metricsServer = metrics.NewServer(cfg.Metrics.Port, cfg.Metrics.Path, log)
		if err := metricsServer.Start(); err != nil {
			log.Fatal("Metrics server failed", zap.Error(err))
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		errChan <- runAgenticTrader(ctx, cfg, *mockMode, *enableAI, *anthropicKey)
	}()

	select {
	case sig := <-sigChan:
		log.Info("Shutdown", zap.String("signal", sig.String()))
		cancel()
	case err := <-errChan:
		if err != nil {
			log.Error("Error", zap.Error(err))
		}
	}

	if metricsServer != nil {
		metricsServer.Stop()
	}
	log.Info("👋 Goodbye!")
}

func runAgenticTrader(ctx context.Context, cfg *config.Config, mockMode, enableAI bool, anthropicKey string) error {
	log := logger.Log

	if mockMode {
		log.Info("🎭 MOCK MODE")
		return runMockTrading(ctx, cfg)
	}

	log.Info("🌐 Connecting to Polymarket")
	
	client := exchange.NewClient(
		cfg.Polymarket.APIURL,
		cfg.Polymarket.APIKey,
		cfg.Polymarket.APISecret,
		cfg.Polymarket.RequestsPerSecond,
		logger.Exchange(),
	)
	defer client.Close()

	var aiDetector *arbitrage.AIDetector
	if enableAI {
		aiDetector = arbitrage.NewAIDetector(anthropicKey, logger.Arbitrage())
		log.Info("🤖 AI ENABLED: Claude Sonnet 4 + Full Mathematical Framework")
	}

	log.Info("✅ System initialized")
	log.Info("🎯 Types: Binary, Categorical, Combinatorial (AI)")
	log.Info("🔍 Starting agentic market scanning...")

	return runAgenticScanning(ctx, cfg, client, aiDetector)
}

func runAgenticScanning(ctx context.Context, cfg *config.Config, client *exchange.Client, aiDetector *arbitrage.AIDetector) error {
	log := logger.Log
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	scanCount := 0

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			scanCount++
			
			fetchCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			markets, err := client.GetMarkets(fetchCtx)
			cancel()
			
			if err != nil {
				log.Error("Fetch failed", zap.Error(err))
				continue
			}

			// Count by type
			binaryCount := 0
			categoricalCount := 0
			
			for _, m := range markets {
				if len(m.Tokens) == 2 {
					binaryCount++
				} else if len(m.Tokens) > 2 {
					categoricalCount++
				}
			}

			log.Info("✅ Markets fetched", 
				zap.Int("scan", scanCount),
				zap.Int("total", len(markets)),
				zap.Int("binary", binaryCount),
				zap.Int("categorical", categoricalCount),
			)

			// Phase 1: Binary arbitrage
			binaryOpportunities := 0
			for _, market := range markets {
				if len(market.Tokens) != 2 {
					continue
				}

				prices := []float64{market.Tokens[0].Price, market.Tokens[1].Price}
				solution := polytope.CheckCategoricalArbitrage(prices, cfg.Arbitrage.MinProfitBps)
				
				if solution.Exists {
					binaryOpportunities++
					log.Info("🎯 BINARY ARBITRAGE!",
						zap.String("market", market.Question),
						zap.Float64("yes", prices[0]),
						zap.Float64("no", prices[1]),
						zap.Float64("sum", prices[0]+prices[1]),
						zap.Int("profit_bps", solution.ProfitBps),
					)
				}
			}

			// Phase 2: Categorical arbitrage
			categoricalOpportunities := 0
			for _, market := range markets {
				if len(market.Tokens) < 3 {
					continue
				}

				prices := make([]float64, len(market.Tokens))
				for i, t := range market.Tokens {
					prices[i] = t.Price
				}

				solution := polytope.CheckCategoricalArbitrage(prices, cfg.Arbitrage.MinProfitBps)
				
				if solution.Exists {
					categoricalOpportunities++
					log.Info("🎯 CATEGORICAL ARBITRAGE!",
						zap.String("market", market.Question),
						zap.Int("outcomes", len(prices)),
						zap.Float64s("prices", prices),
						zap.Int("profit_bps", solution.ProfitBps),
					)
				}
			}

			// Summary
			log.Info("📊 Scan summary",
				zap.Int("binary_opportunities", binaryOpportunities),
				zap.Int("categorical_opportunities", categoricalOpportunities),
			)

			// Phase 3: AI combinatorial (if enabled)
			if aiDetector != nil && len(markets) >= 2 {
				log.Info("🤖 Starting AI+Math combinatorial scan...")
				
				maxPairs := 5
				pairsAnalyzed := 0
				
				for i := 0; i < len(markets) && pairsAnalyzed < maxPairs; i++ {
					for j := i + 1; j < len(markets) && pairsAnalyzed < maxPairs; j++ {
						market1 := markets[i]
						market2 := markets[j]
						
						if len(market1.Tokens) != 2 || len(market2.Tokens) != 2 {
							continue
						}
						
						pairsAnalyzed++
						
						log.Info("🤖 AI analyzing pair",
							zap.String("market1", market1.Question[:min(len(market1.Question), 40)]),
							zap.String("market2", market2.Question[:min(len(market2.Question), 40)]),
						)
						
						analysis, err := aiDetector.AnalyzeMarketPair(ctx, &market1, &market2)
						if err != nil {
							log.Error("AI analysis failed", zap.Error(err))
							continue
						}
						
						if analysis.HasDependency {
							log.Info("🤖💡 AI DETECTED DEPENDENCY!",
								zap.String("type", analysis.DependencyType),
								zap.Float64("confidence", analysis.Confidence),
								zap.String("explanation", analysis.Explanation),
							)
							
							solver := polytope.NewFullIPSolver(analysis.ValidCombinations, 2)
							
							prices := []float64{
								market1.Tokens[0].Price,
								market1.Tokens[1].Price,
								market2.Tokens[0].Price,
								market2.Tokens[1].Price,
							}
							
							opportunity, err := solver.CheckViolation(prices)
							if err != nil {
								log.Error("Solver failed", zap.Error(err))
								continue
							}
							
							if opportunity != nil && opportunity.ProfitBps > cfg.Arbitrage.MinProfitBps {
								log.Info("🎯🤖💰 COMBINATORIAL ARBITRAGE!!!",
									zap.String("market1", market1.Question),
									zap.String("market2", market2.Question),
									zap.Int("profit_bps", opportunity.ProfitBps),
								)
							}
						}
					}
				}
			}
		}
	}
}

func runMockTrading(ctx context.Context, cfg *config.Config) error {
	log := logger.Log
	log.Info("Running in mock mode - using test data")
	
	<-ctx.Done()
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
