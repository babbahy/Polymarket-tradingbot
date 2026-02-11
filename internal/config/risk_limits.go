package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// RiskLimits holds all risk management parameters
type RiskLimits struct {
	Position      PositionLimits      `mapstructure:"position"`
	Orders        OrderLimits         `mapstructure:"orders"`
	Daily         DailyLimits         `mapstructure:"daily"`
	Market        MarketLimits        `mapstructure:"market"`
	Concentration ConcentrationLimits `mapstructure:"concentration"`
	KillSwitch    KillSwitchConfig    `mapstructure:"kill_switch"`
	PreTradeChecks PreTradeChecks     `mapstructure:"pre_trade_checks"`
	Alerts        AlertConfig         `mapstructure:"alerts"`
}

// PositionLimits defines position size constraints
type PositionLimits struct {
	MaxNotionalPerMarket float64 `mapstructure:"max_notional_per_market"`
	MaxTotalNotional     float64 `mapstructure:"max_total_notional"`
	MaxOpenPositions     int     `mapstructure:"max_open_positions"`
	MaxInventorySkew     float64 `mapstructure:"max_inventory_skew"`
}

// OrderLimits defines order constraints
type OrderLimits struct {
	MaxOpenOrders      int     `mapstructure:"max_open_orders"`
	MaxOrderSize       float64 `mapstructure:"max_order_size"`
	MinOrderSize       float64 `mapstructure:"min_order_size"`
	MaxOrdersPerMarket int     `mapstructure:"max_orders_per_market"`
}

// DailyLimits defines daily risk constraints
type DailyLimits struct {
	MaxDailyLoss      float64 `mapstructure:"max_daily_loss"`
	MaxDailyProfit    float64 `mapstructure:"max_daily_profit"`
	MaxTradesPerDay   int     `mapstructure:"max_trades_per_day"`
	ResetTime         string  `mapstructure:"reset_time"`
}

// MarketLimits defines market-level constraints
type MarketLimits struct {
	MinMarketLiquidity float64  `mapstructure:"min_market_liquidity"`
	MaxSpreadBps       int      `mapstructure:"max_spread_bps"`
	Blacklist          []string `mapstructure:"blacklist"`
	Whitelist          []string `mapstructure:"whitelist"`
}

// ConcentrationLimits defines exposure concentration limits
type ConcentrationLimits struct {
	MaxMarketConcentration float64 `mapstructure:"max_market_concentration"`
	MaxEventConcentration  float64 `mapstructure:"max_event_concentration"`
}

// KillSwitchConfig defines kill switch behavior
type KillSwitchConfig struct {
	Enabled             bool                `mapstructure:"enabled"`
	Triggers            KillSwitchTriggers  `mapstructure:"triggers"`
	Actions             KillSwitchActions   `mapstructure:"actions"`
	RequireManualReset  bool                `mapstructure:"require_manual_reset"`
}

// KillSwitchTriggers defines conditions that activate kill switch
type KillSwitchTriggers struct {
	OnDailyLoss        bool    `mapstructure:"on_daily_loss"`
	OnPositionLimit    bool    `mapstructure:"on_position_limit"`
	OnErrorRate        bool    `mapstructure:"on_error_rate"`
	ErrorRateThreshold float64 `mapstructure:"error_rate_threshold"`
	ErrorRateWindow    int     `mapstructure:"error_rate_window"`
	OnConnectionLoss   bool    `mapstructure:"on_connection_loss"`
	ConnectionTimeout  int     `mapstructure:"connection_timeout"`
}

// KillSwitchActions defines actions taken when kill switch activates
type KillSwitchActions struct {
	CancelAllOrders bool `mapstructure:"cancel_all_orders"`
	ClosePositions  bool `mapstructure:"close_positions"`
	SendAlert       bool `mapstructure:"send_alert"`
}

// PreTradeChecks defines which checks to perform before trades
type PreTradeChecks struct {
	CheckPositionLimits    bool `mapstructure:"check_position_limits"`
	CheckOrderLimits       bool `mapstructure:"check_order_limits"`
	CheckDailyLimits       bool `mapstructure:"check_daily_limits"`
	CheckExecutionQuality  bool `mapstructure:"check_execution_quality"`
	CheckMarketLiquidity   bool `mapstructure:"check_market_liquidity"`
	CheckPriceStaleness    bool `mapstructure:"check_price_staleness"`
	MaxPriceAgeSeconds     int  `mapstructure:"max_price_age_seconds"`
}

// AlertConfig defines alerting behavior
type AlertConfig struct {
	WarnThreshold float64  `mapstructure:"warn_threshold"`
	Channels      []string `mapstructure:"channels"`
}

// LoadRiskLimits loads risk limits from a separate config file
func LoadRiskLimits(path string) (*RiskLimits, error) {
	v := viper.New()
	v.SetConfigFile(path)
	
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read risk limits config: %w", err)
	}
	
	var limits RiskLimits
	if err := v.Unmarshal(&limits); err != nil {
		return nil, fmt.Errorf("failed to unmarshal risk limits: %w", err)
	}
	
	if err := limits.Validate(); err != nil {
		return nil, fmt.Errorf("invalid risk limits: %w", err)
	}
	
	return &limits, nil
}

// Validate checks if risk limits are valid
func (r *RiskLimits) Validate() error {
	// Validate position limits
	if r.Position.MaxNotionalPerMarket <= 0 {
		return fmt.Errorf("max_notional_per_market must be positive")
	}
	if r.Position.MaxTotalNotional <= 0 {
		return fmt.Errorf("max_total_notional must be positive")
	}
	if r.Position.MaxNotionalPerMarket > r.Position.MaxTotalNotional {
		return fmt.Errorf("max_notional_per_market cannot exceed max_total_notional")
	}
	if r.Position.MaxInventorySkew < 0 || r.Position.MaxInventorySkew > 1 {
		return fmt.Errorf("max_inventory_skew must be between 0 and 1")
	}
	
	// Validate order limits
	if r.Orders.MaxOrderSize <= 0 {
		return fmt.Errorf("max_order_size must be positive")
	}
	if r.Orders.MinOrderSize <= 0 {
		return fmt.Errorf("min_order_size must be positive")
	}
	if r.Orders.MinOrderSize > r.Orders.MaxOrderSize {
		return fmt.Errorf("min_order_size cannot exceed max_order_size")
	}
	
	// Validate daily limits
	if r.Daily.MaxDailyLoss <= 0 {
		return fmt.Errorf("max_daily_loss must be positive")
	}
	if _, err := time.Parse("15:04:05", r.Daily.ResetTime); err != nil {
		return fmt.Errorf("invalid reset_time format (use HH:MM:SS): %w", err)
	}
	
	// Validate concentration limits
	if r.Concentration.MaxMarketConcentration <= 0 || r.Concentration.MaxMarketConcentration > 1 {
		return fmt.Errorf("max_market_concentration must be between 0 and 1")
	}
	if r.Concentration.MaxEventConcentration <= 0 || r.Concentration.MaxEventConcentration > 1 {
		return fmt.Errorf("max_event_concentration must be between 0 and 1")
	}
	
	return nil
}

// GetResetTime parses the daily reset time
func (d *DailyLimits) GetResetTime() (time.Time, error) {
	parsed, err := time.Parse("15:04:05", d.ResetTime)
	if err != nil {
		return time.Time{}, err
	}
	
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(),
		parsed.Hour(), parsed.Minute(), parsed.Second(), 0, now.Location()), nil
}
