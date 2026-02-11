package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Mode       string           `yaml:"mode"`
	Polymarket PolymarketConfig `yaml:"polymarket"`
	Arbitrage  ArbitrageConfig  `yaml:"arbitrage"`
	Execution  ExecutionConfig  `yaml:"execution"`
	Storage    StorageConfig    `yaml:"storage"`
	Logging    LoggingConfig    `yaml:"logging"`
	Metrics    MetricsConfig    `yaml:"metrics"`
}

type PolymarketConfig struct {
	APIURL             string `yaml:"api_url"`
	APIKey             string `yaml:"api_key"`
	APISecret          string `yaml:"api_secret"`
	RequestsPerSecond  int    `yaml:"requests_per_second"`
}

type ArbitrageConfig struct {
	MinProfitBps    int              `yaml:"min_profit_bps"`
	LiquidityFilter LiquidityFilter  `yaml:"liquidity_filter"`
	ExecutionModel  ExecutionModel   `yaml:"execution_model"`
}

type LiquidityFilter struct {
	Enabled      bool    `yaml:"enabled"`
	MinLiquidity float64 `yaml:"min_liquidity"`
	MaxLiquidity float64 `yaml:"max_liquidity"`
}

type ExecutionModel struct {
	Enabled         bool    `yaml:"enabled"`
	TradingFeeBps   int     `yaml:"trading_fee_bps"`
	GasFeeUSD       float64 `yaml:"gas_fee_usd"`
	MinNetProfitBps int     `yaml:"min_net_profit_bps"`
}

type ExecutionConfig struct {
	SlippageModel   string `yaml:"slippage_model"`
	MaxSlippageBps  int    `yaml:"max_slippage_bps"`
}

type StorageConfig struct {
	Backend  string       `yaml:"backend"`
	SQLite   SQLiteConfig `yaml:"sqlite"`
}

type SQLiteConfig struct {
	Path string `yaml:"path"`
}

type LoggingConfig struct {
	Level             string `yaml:"level"`
	Format            string `yaml:"format"`
	Output            string `yaml:"output"`
	IncludeCaller     bool   `yaml:"include_caller"`
	IncludeStacktrace bool   `yaml:"include_stacktrace"`
}

type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Port    int    `yaml:"port"`
	Path    string `yaml:"path"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.Mode == "" {
		cfg.Mode = "paper"
	}

	if cfg.Polymarket.RequestsPerSecond == 0 {
		cfg.Polymarket.RequestsPerSecond = 10
	}

	if cfg.Arbitrage.MinProfitBps == 0 {
		cfg.Arbitrage.MinProfitBps = 50
	}

	if cfg.Metrics.Port == 0 {
		cfg.Metrics.Port = 9090
	}

	if cfg.Metrics.Path == "" {
		cfg.Metrics.Path = "/metrics"
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	validModes := map[string]bool{
		"paper": true,
		"live":  true,
	}

	if !validModes[c.Mode] {
		return fmt.Errorf("invalid mode: %s (must be 'paper' or 'live')", c.Mode)
	}

	validStorageBackends := map[string]bool{
		"sqlite":   true,
		"postgres": true,
	}

	if !validStorageBackends[c.Storage.Backend] {
		return fmt.Errorf("invalid storage backend: %s", c.Storage.Backend)
	}

	return nil
}
