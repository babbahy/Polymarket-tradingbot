package metrics

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var (
	// Trade metrics
	TradesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "arbitrage_trades_total",
			Help: "Total number of trades executed",
		},
		[]string{"market", "side", "status"},
	)

	TradeProfitUSD = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "arbitrage_trade_profit_usd",
			Help:    "Profit per trade in USD",
			Buckets: []float64{-100, -50, -10, 0, 10, 50, 100, 500, 1000, 5000},
		},
		[]string{"market"},
	)

	TradeExecutionTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "arbitrage_trade_execution_seconds",
			Help:    "Trade execution time in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"market"},
	)

	// Arbitrage detection metrics
	ArbitrageOpportunitiesDetected = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "arbitrage_opportunities_detected_total",
			Help: "Total arbitrage opportunities detected",
		},
		[]string{"type"},
	)

	ArbitrageOpportunitiesExecuted = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "arbitrage_opportunities_executed_total",
			Help: "Total arbitrage opportunities executed",
		},
	)

	ArbitrageOpportunitiesRejected = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "arbitrage_opportunities_rejected_total",
			Help: "Total arbitrage opportunities rejected",
		},
		[]string{"reason"},
	)

	// Position metrics
	OpenPositions = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "arbitrage_open_positions",
			Help: "Number of open positions",
		},
		[]string{"market"},
	)

	PositionNotionalUSD = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "arbitrage_position_notional_usd",
			Help: "Position notional value in USD",
		},
		[]string{"market"},
	)

	TotalPnLUSD = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "arbitrage_total_pnl_usd",
			Help: "Total P&L in USD",
		},
	)

	// Risk metrics
	RiskLimitUtilization = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "arbitrage_risk_limit_utilization",
			Help: "Risk limit utilization (0-1)",
		},
		[]string{"limit_type"},
	)

	KillSwitchActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "arbitrage_kill_switch_active",
			Help: "Kill switch status (1=active, 0=inactive)",
		},
	)

	RiskChecksFailed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "arbitrage_risk_checks_failed_total",
			Help: "Total risk checks that failed",
		},
		[]string{"check_type"},
	)

	// Market data metrics
	OrderbookUpdates = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "arbitrage_orderbook_updates_total",
			Help: "Total orderbook updates received",
		},
		[]string{"market"},
	)

	OrderbookSpreadBps = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "arbitrage_orderbook_spread_bps",
			Help: "Orderbook spread in basis points",
		},
		[]string{"market"},
	)

	OrderbookLiquidityUSD = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "arbitrage_orderbook_liquidity_usd",
			Help: "Total orderbook liquidity in USD",
		},
		[]string{"market", "side"},
	)

	// System metrics
	APIRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "arbitrage_api_requests_total",
			Help: "Total API requests",
		},
		[]string{"endpoint", "status"},
	)

	APIRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "arbitrage_api_request_duration_seconds",
			Help:    "API request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)

	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "arbitrage_errors_total",
			Help: "Total errors by type",
		},
		[]string{"component", "error_type"},
	)
)

// Server provides metrics HTTP endpoint
type Server struct {
	port   int
	path   string
	server *http.Server
	logger *zap.Logger
	mu     sync.Mutex
}

// NewServer creates a new metrics server
func NewServer(port int, path string, logger *zap.Logger) *Server {
	return &Server{
		port:   port,
		path:   path,
		logger: logger,
	}
}

// Start begins serving metrics
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mux := http.NewServeMux()
	mux.Handle(s.path, promhttp.Handler())
	
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	s.logger.Info("Starting metrics server",
		zap.Int("port", s.port),
		zap.String("path", s.path),
	)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Metrics server error", zap.Error(err))
		}
	}()

	return nil
}

// Stop gracefully shuts down the metrics server
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server != nil {
		s.logger.Info("Stopping metrics server")
		return s.server.Close()
	}
	return nil
}

// RecordTrade records a trade execution
func RecordTrade(market, side, status string, profitUSD, executionSeconds float64) {
	TradesTotal.WithLabelValues(market, side, status).Inc()
	TradeProfitUSD.WithLabelValues(market).Observe(profitUSD)
	TradeExecutionTime.WithLabelValues(market).Observe(executionSeconds)
}

// RecordArbitrageOpportunity records detection of an arbitrage opportunity
func RecordArbitrageOpportunity(opportunityType string, executed bool, rejectionReason string) {
	ArbitrageOpportunitiesDetected.WithLabelValues(opportunityType).Inc()
	
	if executed {
		ArbitrageOpportunitiesExecuted.Inc()
	} else if rejectionReason != "" {
		ArbitrageOpportunitiesRejected.WithLabelValues(rejectionReason).Inc()
	}
}

// UpdatePositionMetrics updates position-related metrics
func UpdatePositionMetrics(market string, count int, notionalUSD float64) {
	OpenPositions.WithLabelValues(market).Set(float64(count))
	PositionNotionalUSD.WithLabelValues(market).Set(notionalUSD)
}

// UpdateRiskLimitUtilization updates risk limit utilization
func UpdateRiskLimitUtilization(limitType string, utilization float64) {
	RiskLimitUtilization.WithLabelValues(limitType).Set(utilization)
}

// RecordRiskCheckFailed records a failed risk check
func RecordRiskCheckFailed(checkType string) {
	RiskChecksFailed.WithLabelValues(checkType).Inc()
}

// RecordAPIRequest records an API request
func RecordAPIRequest(endpoint, status string, duration float64) {
	APIRequestsTotal.WithLabelValues(endpoint, status).Inc()
	APIRequestDuration.WithLabelValues(endpoint).Observe(duration)
}

// RecordError records an error
func RecordError(component, errorType string) {
	ErrorsTotal.WithLabelValues(component, errorType).Inc()
}
