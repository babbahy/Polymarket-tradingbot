package logger

import (
	"bytes"
	"encoding/json"
	"testing"

	"go.uber.org/zap"
)

func TestLoggerInitialization(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "json format to stdout",
			config: Config{
				Level:  "info",
				Format: "json",
				Output: "stdout",
			},
		},
		{
			name: "console format",
			config: Config{
				Level:  "debug",
				Format: "console",
				Output: "stdout",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Initialize(tt.config)
			if err != nil {
				t.Errorf("Initialize() error = %v", err)
			}
			if Log == nil {
				t.Error("Logger not initialized")
			}
			Sync()
		})
	}
}

func TestWithFields(t *testing.T) {
	// Initialize a test logger
	var buf bytes.Buffer
	encoder := zap.NewDevelopmentEncoderConfig()
	encoder.TimeKey = ""
	core := zap.NewCore(
		zap.NewJSONEncoder(encoder),
		zap.AddSync(&buf),
		zap.DebugLevel,
	)
	Log = zap.New(core)

	logger := WithFields(
		zap.String("component", "test"),
		zap.Int("value", 42),
	)

	logger.Info("test message")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	if logEntry["component"] != "test" {
		t.Errorf("Expected component=test, got %v", logEntry["component"])
	}
	if logEntry["value"] != float64(42) {
		t.Errorf("Expected value=42, got %v", logEntry["value"])
	}
}

func TestComponentLoggers(t *testing.T) {
	// Initialize a test logger
	var buf bytes.Buffer
	encoder := zap.NewDevelopmentEncoderConfig()
	encoder.TimeKey = ""
	core := zap.NewCore(
		zap.NewJSONEncoder(encoder),
		zap.AddSync(&buf),
		zap.DebugLevel,
	)
	Log = zap.New(core)

	components := map[string]*zap.Logger{
		"trade":     Trade(),
		"risk":      Risk(),
		"arbitrage": Arbitrage(),
		"exchange":  Exchange(),
		"metrics":   Metrics(),
	}

	for name, logger := range components {
		buf.Reset()
		logger.Info("test message")

		var logEntry map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
			t.Fatalf("Failed to parse log output for %s: %v", name, err)
		}

		if logEntry["component"] != name {
			t.Errorf("Expected component=%s, got %v", name, logEntry["component"])
		}
	}
}
