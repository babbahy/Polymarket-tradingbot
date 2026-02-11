package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// Config holds logger configuration
type Config struct {
	Level            string
	Format           string // "json" or "console"
	Output           string // "stdout" or file path
	IncludeCaller    bool
	IncludeStacktrace bool
}

// Initialize sets up the global logger
func Initialize(cfg Config) error {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	var writeSyncer zapcore.WriteSyncer
	if cfg.Output == "stdout" {
		writeSyncer = zapcore.AddSync(os.Stdout)
	} else {
		file, err := os.OpenFile(cfg.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		writeSyncer = zapcore.AddSync(file)
	}

	core := zapcore.NewCore(encoder, writeSyncer, level)

	options := []zap.Option{
		zap.AddCaller(),
	}

	if !cfg.IncludeCaller {
		options = append(options, zap.WithCaller(false))
	}

	if cfg.IncludeStacktrace {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	Log = zap.New(core, options...)
	
	Log.Info("Logger initialized",
		zap.String("level", cfg.Level),
		zap.String("format", cfg.Format),
		zap.String("output", cfg.Output),
	)

	return nil
}

// Sync flushes any buffered log entries
func Sync() error {
	if Log != nil {
		return Log.Sync()
	}
	return nil
}

// WithFields creates a logger with additional fields
func WithFields(fields ...zap.Field) *zap.Logger {
	if Log == nil {
		return zap.NewNop()
	}
	return Log.With(fields...)
}

// Trade creates a logger for trade-related events
func Trade() *zap.Logger {
	return WithFields(zap.String("component", "trade"))
}

// Risk creates a logger for risk management events
func Risk() *zap.Logger {
	return WithFields(zap.String("component", "risk"))
}

// Arbitrage creates a logger for arbitrage detection
func Arbitrage() *zap.Logger {
	return WithFields(zap.String("component", "arbitrage"))
}

// Exchange creates a logger for exchange interactions
func Exchange() *zap.Logger {
	return WithFields(zap.String("component", "exchange"))
}

// Metrics creates a logger for metrics events
func Metrics() *zap.Logger {
	return WithFields(zap.String("component", "metrics"))
}
