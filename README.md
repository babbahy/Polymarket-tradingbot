# Polymarket CLOB Arbitrage Bot

Automated arbitrage trading bot for Polymarket's Central Limit Order Book (CLOB), implementing marginal polytope optimization and dependency-aware execution under non-atomic constraints.

## ⚠️ IMPORTANT: Paper Trading First

**This bot starts in PAPER TRADING mode by default.** Live trading is behind a feature flag and requires explicit configuration.

## Features

- ✅ **Paper Trading**: Full simulation with realistic execution modeling
- ✅ **Dry-Run Mode**: Detect opportunities without placing orders
- ✅ **Risk Management**: Comprehensive pre-trade checks and hard limits
- ✅ **Execution Simulation**: VWAP/slippage/liquidity modeling
- ✅ **Kill Switch**: Automatic emergency stop on limit breaches
- ✅ **Metrics**: Prometheus-compatible metrics endpoint
- ✅ **Structured Logging**: Detailed audit trail of all decisions

## Architecture

```
polymarket-arbitrage/
├── cmd/trader/          # Main trading bot
├── internal/
│   ├── config/          # Configuration management
│   ├── exchange/        # Polymarket CLOB API client
│   ├── arbitrage/       # Marginal polytope arbitrage detection
│   ├── execution/       # Trade execution (paper & live)
│   ├── risk/            # Risk management & limits
│   ├── store/           # Position & trade persistence
│   ├── metrics/         # Prometheus metrics
│   └── logger/          # Structured logging
├── pkg/polytope/        # Mathematical optimization library
└── configs/             # Configuration files
```

## Quick Start

### Prerequisites

- Go 1.22+
- Polymarket API credentials (for live trading)

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/polymarket-arbitrage
cd polymarket-arbitrage

# Install dependencies
make deps

# Copy environment template
cp .env.example .env
# Edit .env with your credentials (optional for paper trading)
```

### Running

```bash
# Paper trading (default)
make run-paper

# Dry-run mode (detect but don't execute)
make run-dry

# Build and run manually
make build
./bin/polymarket-arbitrage --mode=paper --config=configs/config.yaml
```

## Configuration

### Trading Modes

1. **dry-run**: Detects arbitrage but places NO orders (safest)
2. **paper**: Simulates execution with realistic models (default)
3. **live**: Real trading ⚠️ **Requires feature flag**

### Risk Limits

All limits are enforced BEFORE any trade. See `configs/risk_limits.yaml`:

- **Position Limits**: Max notional per market, total exposure, position count
- **Order Limits**: Max order size, open orders, orders per market
- **Daily Limits**: Max daily loss, profit circuit breaker, trade count
- **Kill Switch**: Automatic stop on limit breach, connection loss, error rate

### Feature Flags

In `configs/config.yaml`:

```yaml
features:
  enable_live_trading: false  # Must be true for live mode
  enable_short_positions: false
  enable_multi_market_arb: true
```

## Risk Management

### Pre-Trade Checks

Every trade must pass:
1. Position limit check
2. Order limit check
3. Daily loss limit check
4. Execution quality simulation (VWAP/slippage)
5. Market liquidity verification
6. Price staleness check

### Kill Switch

Automatically triggered by:
- Daily loss limit breach
- Position limit breach
- High error rate (configurable threshold)
- API connection loss

Actions:
- Cancel all open orders
- Send alerts
- Require manual reset (configurable)

## Metrics

Prometheus metrics available at `http://localhost:9090/metrics`:

### Trade Metrics
- `arbitrage_trades_total`: Total trades by market/side/status
- `arbitrage_trade_profit_usd`: Profit distribution
- `arbitrage_trade_execution_seconds`: Execution time

### Arbitrage Metrics
- `arbitrage_opportunities_detected_total`: Opportunities found
- `arbitrage_opportunities_executed_total`: Opportunities executed
- `arbitrage_opportunities_rejected_total`: Rejections by reason

### Risk Metrics
- `arbitrage_risk_limit_utilization`: Limit usage (0-1)
- `arbitrage_kill_switch_active`: Kill switch status
- `arbitrage_risk_checks_failed_total`: Failed checks by type

### Position Metrics
- `arbitrage_open_positions`: Current position count
- `arbitrage_position_notional_usd`: Position sizes
- `arbitrage_total_pnl_usd`: Total P&L

## Development

```bash
# Run tests
make test

# Run with coverage
make test-coverage

# Run linter
make lint

# Format code
make fmt

# Run all checks (CI)
make ci
```

## Deployment

### Docker

```bash
# Build image
make docker-build

# Run in paper mode
make docker-run-paper
```

### Environment Variables

See `.env.example` for all configuration options.

**Required for live trading:**
- `POLYMARKET_API_KEY`
- `POLYMARKET_API_SECRET`
- `POLYMARKET_PASSPHRASE`
- `PRIVATE_KEY`

## Live Trading Checklist

Before enabling live trading:

- [ ] Test thoroughly in paper mode
- [ ] Verify risk limits are appropriate
- [ ] Set up monitoring and alerts
- [ ] Test kill switch behavior
- [ ] Start with minimal capital
- [ ] Set `enable_live_trading: true` in config
- [ ] Run with `--mode=live` flag

## Monitoring

### Health Check

```bash
curl http://localhost:8080/health
```

### View Metrics

```bash
curl http://localhost:9090/metrics
```

### Logs

Structured JSON logs include:
- All arbitrage opportunities (detected & rejected)
- Trade executions (simulated & real)
- Risk check results
- Position updates
- API interactions

## Project Status

🚧 **Work in Progress** 🚧

### Implemented
- ✅ Configuration system
- ✅ Logging infrastructure
- ✅ Metrics server
- ✅ Risk limit definitions
- ✅ Main application scaffold

### TODO
- [ ] Polymarket CLOB API client
- [ ] WebSocket orderbook subscriptions
- [ ] Marginal polytope solver
- [ ] Arbitrage detection engine
- [ ] Execution simulator (VWAP/slippage)
- [ ] Position tracker
- [ ] Risk checker implementation
- [ ] Kill switch implementation
- [ ] Live execution engine (feature-flagged)
- [ ] Backtesting tool

## Contributing

Contributions welcome! Please:
1. Write tests for new features
2. Follow the existing code style
3. Update documentation
4. Run `make verify` before committing

## License

MIT License - see LICENSE file

## Disclaimer

This software is for educational purposes only. Trading involves substantial risk of loss. The authors assume no responsibility for any financial losses incurred through use of this software. Always test thoroughly in paper trading mode before considering live trading.

**Never risk more than you can afford to lose.**
