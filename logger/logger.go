// Package logger provides common logging setup using zap.
package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds logger configuration
type Config struct {
	Level      string // debug, info, warn, error
	OutputPath string // stdout, stderr, or file path
	Format     string // json, console
}

// DefaultConfig returns default logger configuration
func DefaultConfig() Config {
	return Config{
		Level:      "info",
		OutputPath: "stdout",
		Format:     "json",
	}
}

// New creates a new zap logger with the given configuration
func New(cfg Config) (*zap.Logger, error) {
	// Parse log level
	level := parseLevel(cfg.Level)

	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder

	// Select encoder based on format
	var encoder zapcore.Encoder
	if cfg.Format == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Create output writer
	var output zapcore.WriteSyncer
	switch strings.ToLower(cfg.OutputPath) {
	case "stdout", "":
		output = zapcore.AddSync(os.Stdout)
	case "stderr":
		output = zapcore.AddSync(os.Stderr)
	default:
		file, err := os.OpenFile(cfg.OutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		output = zapcore.AddSync(file)
	}

	// Create core
	core := zapcore.NewCore(encoder, output, level)

	// Create logger with options
	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	return logger, nil
}

// NewDevelopment creates a development logger with console output
func NewDevelopment() (*zap.Logger, error) {
	return New(Config{
		Level:      "debug",
		OutputPath: "stdout",
		Format:     "console",
	})
}

// NewProduction creates a production logger with JSON output
func NewProduction() (*zap.Logger, error) {
	return New(Config{
		Level:      "info",
		OutputPath: "stdout",
		Format:     "json",
	})
}

// FromEnv creates a logger from environment variables
func FromEnv() (*zap.Logger, error) {
	cfg := DefaultConfig()

	if level := os.Getenv("LOG_LEVEL"); level != "" {
		cfg.Level = level
	}
	if output := os.Getenv("LOG_OUTPUT"); output != "" {
		cfg.OutputPath = output
	}
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		cfg.Format = format
	}

	// Use console format for dev environment
	if env := os.Getenv("ENV"); env == "dev" || env == "development" {
		cfg.Format = "console"
	}

	return New(cfg)
}

// parseLevel converts string level to zapcore.Level
func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// WithService adds service name to logger
func WithService(logger *zap.Logger, serviceName string) *zap.Logger {
	return logger.With(zap.String("service", serviceName))
}

// WithRequestID adds request ID to logger
func WithRequestID(logger *zap.Logger, requestID string) *zap.Logger {
	return logger.With(zap.String("request_id", requestID))
}
