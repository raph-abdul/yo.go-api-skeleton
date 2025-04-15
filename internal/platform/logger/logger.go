// Copyright 2025 raph-abdul
// Licensed under the Apache License, Version 2.0.
// Visit http://www.apache.org/licenses/LICENSE-2.0 for details

// Package logger /youGo/internal/platform/logger/logger.go
package logger

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new Zap logger instance based on configuration.
// level: "debug", "info", "warn", "error", "dpanic", "panic", "fatal"
// format: "console" or "json"
// appEnv: "development" or "production" (influences defaults)
func New(level string, format string, appEnv string) (*zap.Logger, error) {
	var zapLevel zapcore.Level
	// Parse log level string
	err := zapLevel.UnmarshalText([]byte(strings.ToLower(level)))
	if err != nil {
		zapLevel = zap.InfoLevel // Default to InfoLevel if parsing fails
		fmt.Fprintf(os.Stderr, "Warning: Invalid log level '%s'. Defaulting to 'info'.\n", level)
	}

	var cfg zap.Config
	// Choose base config based on environment
	if appEnv == "development" {
		cfg = zap.NewDevelopmentConfig()
		// Development defaults: console encoder, debug level, caller, stacktrace for errors
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Colored levels
	} else {
		cfg = zap.NewProductionConfig()
		// Production defaults: json encoder, info level, no caller, stacktrace for errors
	}

	// Override level based on config
	cfg.Level = zap.NewAtomicLevelAt(zapLevel)

	// Override encoding based on config
	if strings.ToLower(format) == "console" {
		cfg.Encoding = "console"
		// Ensure colored output for console in development
		if appEnv == "development" {
			cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
	} else if strings.ToLower(format) == "json" {
		cfg.Encoding = "json"
	} else {
		fmt.Fprintf(os.Stderr, "Warning: Invalid log format '%s'. Using default '%s'.\n", format, cfg.Encoding)
	}

	// Disable caller and stacktrace in production unless explicitly needed and level allows
	if appEnv != "development" {
		cfg.DisableCaller = true
		// Only include stacktrace for Error level or higher in production
		cfg.DisableStacktrace = zapLevel > zap.ErrorLevel
	}

	// Add custom fields if needed
	// cfg.InitialFields = map[string]interface{}{"service": "youGo"}

	// Build the logger
	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	// Optional: Redirect standard log output to Zap
	// zap.RedirectStdLog(logger)

	logger.Info("Logger initialized",
		zap.String("level", zapLevel.String()),
		zap.String("format", cfg.Encoding),
		zap.String("environment", appEnv),
	)

	return logger, nil
}
