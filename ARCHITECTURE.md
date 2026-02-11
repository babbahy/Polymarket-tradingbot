# Polymarket CLOB Arbitrage Bot - Architecture & Implementation Guide

## Overview

This is a production-ready scaffold for an automated arbitrage trading bot targeting Polymarket's Central Limit Order Book (CLOB). The system implements marginal polytope optimization to detect arbitrage opportunities across dependent prediction markets, with comprehensive risk management and execution simulation.

## Architecture Principles

### Safety First
- **Paper Trading Default**: System starts in simulation mode
- **Feature-Flagged Live Trading**: Requires explicit configuration
- **Hard Risk Limits**: Non-bypassable position, order, and loss limits
- **Kill Switch**: Automatic emergency stop on breach conditions
- **Pre-Trade Validation**: Every trade passes execution simulation + risk checks

### Observability
- **Structured Logging**: JSON logs with component tagging
- **Prometheus Metrics**: Comprehensive trade, risk, and system metrics
- **Health Checks**: HTTP endpoints for monitoring
- **Audit Trail**: Full record of opportunities detected, rejected, and executed

### Modularity
- **Clean Separation**: Exchange, arbitrage, execution, risk as independent modules
- **Testable Components**: Each module can be tested in isolation
- **Configuration-Driven**: All parameters externalized to YAML

## Directory Structure

```
polymarket-arbitrage/
│
├── cmd/                          # Application entrypoints
│   ├── trader/                   # Main trading bot
│   │   └── main.go              # CLI, initialization, main loop
│   └── backtest/                # Backtesting tool
│       └── main.go              # Historical replay and analysis
│
├── internal/                     # Private application code
│   │
│   ├── config/                   # Configuration management
│   │   ├── config.go            # Main config loader and validation
│   │   └── risk_limits.go       # Risk parameter definitions
│   │
│   ├── exchange/                 # Polymarket CLOB client (TODO)
│   │   ├── client.go            # HTTP/REST client
│   │   ├── types.go             # API types and responses
│   │   └── orderbook.go         # WebSocket orderbook subscription
│   │
│   ├── arbitrage/                # Arbitrage detection engine (TODO)
│   │   ├── detector.go          # Opportunity detection coordinator
│   │   ├── optimizer.go         # Execution path optimizer
│   │   └── graph.go             # Market dependency graph
│   │
│   ├── execution/                # Trade execution (TODO)
│   │   ├── simulator.go         # Paper trading + VWAP/slippage simulation
│   │   ├── executor.go          # Live execution (feature-flagged)
│   │   └── slippage.go          # Slippage model implementations
│   │
│   ├── risk/                     # Risk management (TODO)
│   │   ├── limits.go            # Position/notional/order limits
│   │   ├── checker.go           # Pre-trade risk validation
│   │   └── killswitch.go        # Emergency stop mechanism
│   │
│   ├── store/                    # State persistence (TODO)
│   │   ├── positions.go         # Position tracking
│   │   └── trades.go            # Trade history
│   │
│   ├── metrics/                  # Observability
│   │   ├── server.go            # ✅ Prometheus HTTP server
│   │   └── collector.go         # TODO: Metric collectors
│   │
│   └── logger/                   # Structured logging
│       └── logger.go            # ✅ Zap logger setup
│
├── pkg/                          # Public reusable libraries
│   └── polytope/                # Marginal polytope math (TODO)
│       ├── solver.go            # Linear programming solver
│       └── constraints.go       # Non-atomic constraint handling
│
├── configs/                      # Configuration files
│   ├── config.yaml              # ✅ Main application config
│   ├── risk_limits.yaml         # ✅ Risk parameters
│   └── markets.yaml             # ✅ Market definitions
│
├── scripts/                      # Development tools
│   └── dev.sh                   # ✅ Development helper script
│
├── test/                         # Integration tests
│   └── integration/             # TODO: End-to-end tests
│
├── .github/workflows/            # CI/CD
│   └── ci.yml                   # ✅ GitHub Actions workflow
│
├── Makefile                      # ✅ Build automation
├── Dockerfile                    # ✅ Container build
├── go.mod                        # ✅ Go dependencies
├── README.md                     # ✅ User documentation
└── .env.example                  # ✅ Environment template
```

## Component Responsibilities

### 1. Configuration (`internal/config`)
**Status: ✅ Complete**

- Loads YAML configuration with environment variable overrides
- Validates all settings before startup
- Enforces safety invariants (e.g., live mode requires feature flag)
- Separate risk limits file for easy adjustment

**Key Features:**
- Mode validation (dry-run, paper, live)
- Feature flag enforcement
- Risk parameter validation
- Configuration immutability after load

### 2. Logger (`internal/logger`)
**Status: ✅ Complete**

- Structured JSON logging via Zap
- Component-tagged loggers (trade, risk, arbitrage, exchange, metrics)
- Configurable output and format
- Performance-optimized for high-frequency logging

**Key Features:**
- Structured fields for filtering/analysis
- Component separation
- Caller information (optional)
- Log level control

### 3. Metrics (`internal/metrics`)
**Status: ✅ Complete**

- Prometheus-compatible HTTP server
- Pre-defined metrics for trades, positions, risk, market data
- Health check endpoint
- Helper functions for metric recording

**Metric Categories:**
- Trade execution (count, profit, latency)
- Arbitrage detection (opportunities, rejections)
- Position tracking (count, notional, P&L)
- Risk utilization (limit usage, kill switch status)
- System health (API requests, errors)

### 4. Exchange Client (`internal/exchange`)
**Status: ⏳ TODO**

**Responsibilities:**
- HTTP REST API client for Polymarket CLOB
- WebSocket subscriptions for orderbook updates
- Rate limiting (10 req/s default)
- Request signing (HMAC or EVM wallet)
- Connection health monitoring

**Implementation Notes:**
- Use `gorilla/websocket` for WS connections
- Implement automatic reconnection with exponential backoff
- Parse orderbook into standardized internal format
- Validate responses and handle API errors gracefully

### 5. Arbitrage Detector (`internal/arbitrage`)
**Status: ⏳ TODO**

**Responsibilities:**
- Build dependency graph from market relationships
- Apply marginal polytope constraints
- Detect profitable arbitrage cycles
- Optimize execution paths

**Implementation Notes:**
- Use `gonum` for linear programming
- Implement constraint solvers for:
  - Mutually exclusive outcomes (probabilities sum to 1)
  - Conditional markets (parent-child dependencies)
  - Non-atomic execution constraints
- Cache dependency graphs for performance
- Filter opportunities by minimum profit threshold (basis points)

### 6. Execution Engine (`internal/execution`)
**Status: ⏳ TODO**

**Responsibilities:**
- **Simulator**: VWAP calculation, slippage modeling, liquidity checks
- **Executor**: Order placement, fill tracking, partial fill handling
- Paper trading simulation
- Order splitting for large trades

**Slippage Models:**
- Linear: `slippage = k * size`
- Square root: `slippage = k * sqrt(size)`
- Exponential: `slippage = k * (e^(size/liquidity) - 1)`

**Implementation Notes:**
- Simulate each leg of arbitrage independently
- Calculate expected profit after slippage
- Support partial fills and order cancellation
- Track simulated vs actual execution for model improvement

### 7. Risk Manager (`internal/risk`)
**Status: ⏳ TODO**

**Responsibilities:**
- Pre-trade validation against all limits
- Position tracking and limit enforcement
- Daily P&L calculation and reset
- Kill switch activation and alert generation

**Risk Checks (all must pass):**
1. Position limit check (per-market and total)
2. Order limit check (size, count, per-market)
3. Daily loss limit check
4. Execution quality check (slippage within bounds)
5. Market liquidity check
6. Price staleness check (age < threshold)

**Kill Switch Triggers:**
- Daily loss limit breach
- Position limit breach
- High error rate (>20% of last 100 trades)
- Connection loss (>30s without heartbeat)

### 8. Storage (`internal/store`)
**Status: ⏳ TODO**

**Responsibilities:**
- Position persistence (open, closed)
- Trade history (all executions)
- P&L calculation
- Data retention management

**Schema (SQLite):**
```sql
CREATE TABLE positions (
    id TEXT PRIMARY KEY,
    market_id TEXT NOT NULL,
    side TEXT NOT NULL,  -- 'long' or 'short'
    size REAL NOT NULL,
    entry_price REAL NOT NULL,
    current_price REAL,
    unrealized_pnl REAL,
    status TEXT NOT NULL,  -- 'open' or 'closed'
    opened_at TIMESTAMP NOT NULL,
    closed_at TIMESTAMP
);

CREATE TABLE trades (
    id TEXT PRIMARY KEY,
    position_id TEXT,
    market_id TEXT NOT NULL,
    side TEXT NOT NULL,
    size REAL NOT NULL,
    price REAL NOT NULL,
    profit REAL,
    execution_time REAL,
    status TEXT NOT NULL,  -- 'filled', 'partial', 'rejected'
    rejection_reason TEXT,
    executed_at TIMESTAMP NOT NULL,
    FOREIGN KEY (position_id) REFERENCES positions(id)
);

CREATE TABLE daily_stats (
    date DATE PRIMARY KEY,
    total_pnl REAL,
    trade_count INTEGER,
    win_count INTEGER,
    loss_count INTEGER
);
```

### 9. Polytope Solver (`pkg/polytope`)
**Status: ⏳ TODO**

**Responsibilities:**
- Linear programming for arbitrage detection
- Constraint satisfaction for market dependencies
- Optimization of execution paths

**Mathematical Foundation:**

Given markets M₁, M₂, ..., Mₙ with prices p₁, p₂, ..., pₙ and dependency constraints C:

1. **Constraint Formulation:**
   - Sum constraints: Σpᵢ = 1 (mutually exclusive)
   - Conditional constraints: pⱼ ≤ pᵢ (parent-child)
   - Liquidity constraints: qᵢ ≤ Lᵢ (available size)

2. **Objective Function:**
   Maximize: Σ(qᵢ × (true_value - pᵢ))
   Subject to: C and budget constraints

3. **Non-Atomic Execution:**
   - Model market impact: p'ᵢ = pᵢ + impact(qᵢ)
   - Sequential optimization for execution order

**Implementation Notes:**
- Use `gonum.org/v1/gonum/optimize` for LP solver
- Implement custom constraint handling for market structures
- Cache constraint matrices for performance
- Handle numerical stability (small probabilities)

## Configuration Reference

### Main Config (`configs/config.yaml`)

```yaml
mode: "paper"  # dry-run, paper, live

features:
  enable_live_trading: false
  enable_short_positions: false
  enable_multi_market_arb: true

polymarket:
  api_url: "https://clob.polymarket.com"
  ws_url: "wss://ws-subscriptions-clob.polymarket.com"
  api_key: "${POLYMARKET_API_KEY}"

arbitrage:
  min_profit_bps: 50  # 0.5%
  max_execution_time: 5  # seconds

execution:
  slippage_model: "sqrt"
  max_slippage_bps: 100  # 1%
  min_liquidity_usd: 1000

logging:
  level: "info"
  format: "json"

metrics:
  enabled: true
  port: 9090
```

### Risk Limits (`configs/risk_limits.yaml`)

```yaml
position:
  max_notional_per_market: 10000
  max_total_notional: 50000
  max_open_positions: 20

orders:
  max_order_size: 5000
  min_order_size: 10

daily:
  max_daily_loss: 5000
  max_trades_per_day: 500

kill_switch:
  enabled: true
  triggers:
    on_daily_loss: true
    on_error_rate: true
```

## Development Workflow

### Initial Setup
```bash
# Install dependencies
make deps

# Run tests
make test

# Run linter
make lint
```

### Running the Bot
```bash
# Paper trading (default)
make run-paper

# Dry-run (detect only)
make run-dry

# With custom config
./bin/polymarket-arbitrage --mode=paper --config=custom.yaml
```

### Monitoring
```bash
# View metrics
curl http://localhost:9090/metrics

# Health check
curl http://localhost:8080/health

# View logs (if logging to file)
tail -f logs/arbitrage.log | jq
```

## Testing Strategy

### Unit Tests
- Configuration validation
- Risk limit checking
- Slippage calculations
- Constraint solving
- Metric recording

### Integration Tests
- Exchange API mocking
- End-to-end arbitrage detection
- Execution simulation accuracy
- Kill switch activation

### Backtesting
- Historical orderbook replay
- Strategy performance analysis
- Risk metric validation
- Parameter optimization

## Deployment Checklist

### Pre-Launch (Paper Trading)
- [ ] Run paper trading for 1+ week
- [ ] Verify all metrics are collected
- [ ] Test kill switch triggers manually
- [ ] Review logs for errors/warnings
- [ ] Validate risk limits are respected

### Live Trading Preparation
- [ ] Set conservative risk limits
- [ ] Enable kill switch
- [ ] Set up monitoring alerts
- [ ] Test with minimal capital first
- [ ] Document rollback procedure
- [ ] Set `enable_live_trading: true`

### Production Monitoring
- [ ] Prometheus + Grafana dashboard
- [ ] Alert on kill switch activation
- [ ] Alert on API errors
- [ ] Alert on approaching limits (80%)
- [ ] Daily P&L review

## Next Steps for Implementation

### Phase 1: Market Data (Week 1)
1. Implement Polymarket CLOB HTTP client
2. WebSocket orderbook subscription
3. Orderbook parsing and normalization
4. Connection health monitoring

### Phase 2: Detection (Week 2)
1. Dependency graph construction
2. Marginal polytope solver
3. Arbitrage opportunity detection
4. Profit calculation (pre-slippage)

### Phase 3: Simulation (Week 3)
1. VWAP calculator
2. Slippage models (linear, sqrt, exponential)
3. Liquidity checks
4. Execution simulation

### Phase 4: Risk (Week 4)
1. Position tracker
2. Pre-trade risk checks
3. Kill switch implementation
4. Alert system

### Phase 5: Paper Trading (Week 5-6)
1. Simulated execution engine
2. Performance metrics
3. Paper trading validation
4. Bug fixes and optimization

### Phase 6: Live Trading (Week 7+)
1. Real execution engine (feature-flagged)
2. Order placement and tracking
3. Partial fill handling
4. Live testing with minimal capital

## Safety Reminders

⚠️ **CRITICAL SAFETY POINTS** ⚠️

1. **Never skip paper trading** - Run for at least 1 week before live
2. **Start small** - Use minimal capital for initial live testing
3. **Monitor continuously** - Set up alerts and review daily
4. **Respect the kill switch** - If it triggers, investigate before resetting
5. **Log everything** - Audit trail is essential for debugging and compliance
6. **Test limits** - Manually verify risk limits work before live trading
7. **Have a rollback plan** - Know how to safely shut down and unwind positions

## Performance Targets

- **Detection latency**: <100ms from orderbook update to opportunity detection
- **Execution latency**: <500ms from detection to order placement
- **API request rate**: <10 req/s (within Polymarket limits)
- **Memory usage**: <500MB for normal operation
- **CPU usage**: <50% of single core

## Dependencies

### Go Packages
- `gorilla/websocket`: WebSocket client
- `prometheus/client_golang`: Metrics
- `spf13/viper`: Configuration
- `uber/zap`: Logging
- `gonum/gonum`: Linear programming

### External Services
- Polymarket CLOB API
- Prometheus (optional, for metrics aggregation)
- Grafana (optional, for dashboards)

---

**Built with Go** - Chosen for concurrency, simplicity, and deployment ease.

**Author**: Claude (Anthropic)  
**License**: MIT  
**Status**: Scaffold Complete ✅ | Core Logic TODO ⏳
