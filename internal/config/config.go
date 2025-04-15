// Copyright 2025 raph-abdul
// Licensed under the Apache License, Version 2.0.
// Visit http://www.apache.org/licenses/LICENSE-2.0 for details

// Package config /youGo/internal/config/config.go
package config

import (
	"errors"
	"fmt"
	"strings" // Needed for environment variable replacer

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
// Values are loaded from config files and/or environment variables.
// Struct tags (`mapstructure`) define mapping from config file keys or env vars.
type Config struct {
	App      AppConfig    `mapstructure:"app"`
	Server   ServerConfig `mapstructure:"server"`
	Log      LogConfig    `mapstructure:"log"`
	Auth     AuthConfig   `mapstructure:"auth"`
	Database Database     `mapstructure:"database"`
}

// AppConfig holds application-specific configuration.
type AppConfig struct {
	Env string `mapstructure:"env"` // e.g., "development", "staging", "production"
}

// ServerConfig holds HTTP server-specific configuration.
type ServerConfig struct {
	Port               string   `mapstructure:"port"`
	CORSAllowedOrigins []string `mapstructure:"cors_allowed_origins"` // Note: Corrected spelling from main.go comment example
}

// Database holds database connection details.
type Database struct {
	Host        string `mapstructure:"host"`
	Port        string `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"` // IMPORTANT: Load sensitive data like passwords from ENV VARS in production.
	DBName      string `mapstructure:"dbname"`
	SSLMode     string `mapstructure:"sslmode"` // e.g., "disable", "require", "verify-full"
	AutoMigrate bool   `mapstructure:"auto_migrate"`
	// You might add connection pool settings here if needed
	// MaxIdleConns int `mapstructure:"max_idle_conns"`
	// MaxOpenConns int `mapstructure:"max_open_conns"`
	// ConnMaxLifetime string `mapstructure:"conn_max_lifetime"` // e.g., "1h"
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level  string `mapstructure:"level"`  // e.g., "debug", "info", "warn", "error"
	Format string `mapstructure:"format"` // e.g., "json", "console"
}

// AuthConfig holds authentication related configuration.
type AuthConfig struct {
	JWTSecret            string `mapstructure:"jwt_secret"`
	AccessTokenDuration  string `mapstructure:"access_token_duration"`  // e.g., "15m", "1h", "24h"
	RefreshTokenDuration string `mapstructure:"refresh_token_duration"` // e.g., "7d", "168h"	// You might add token expiry durations here
}

// Load configuration from file and environment variables.
// path: Directory where the config file is located (e.g., "./configs").
// name: Name of the config file without extension (e.g., "config").
// Viper automatically detects the extension (.yaml, .json, .toml, etc.).
// Environment variables can override file settings. They should be prefixed (e.g., "APP_")
// and use underscores instead of dots (e.g., APP_DATABASE_HOST maps to Database.Host).
func Load(path, name string) (*Config, error) {
	v := viper.New()

	// --- File Loading ---
	v.AddConfigPath(path)   // Set the path to look for the config file in
	v.SetConfigName(name)   // Set the name of the config file (without extension)
	v.SetConfigType("yaml") // Explicitly set the config type (can be inferred)

	// Attempt to read the config file
	err := v.ReadInConfig()
	// Handle error BUT allow "config file not found" if env vars might provide all config
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found; rely on environment variables or defaults. Log this maybe?
		// fmt.Println("Config file not found, relying on environment variables or defaults.")
	}

	fmt.Println("Viper config after ReadInConfig:", v.AllSettings()) // Log config after file read

	// --- Environment Variable Loading ---
	v.AutomaticEnv() // Read in environment variables that match
	// Set a prefix to avoid collisions with other system env vars
	// Example: APP_SERVER_PORT=8080 will override Server.Port
	v.SetEnvPrefix("APP")
	// Replace dots in struct paths with underscores for env var names
	// e.g., Database.Host becomes APP_DATABASE_HOST
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	fmt.Println("Viper config after AutomaticEnv:", v.AllSettings()) // Log config after env override

	// --- Unmarshalling ---
	var cfg Config
	err = v.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	// --- Sensitive Data Check (Important!) ---
	if cfg.App.Env == "production" && cfg.Auth.JWTSecret == "" {
		return nil, errors.New("JWT secret cannot be empty in production")
	}

	// --- Sensitive Data Check (Optional but Recommended) ---
	// You might want to add checks here to ensure critical secrets (DB password, JWT secret)
	// are not empty, especially in production environments (cfg.App.Env == "production").
	// if cfg.App.Env == "production" && cfg.Database.Password == "" {
	//  return nil, errors.New("database password cannot be empty in production")
	// }
	// if cfg.App.Env == "production" && cfg.Auth.JWTSecret == "" {
	//  return nil, errors.New("JWT secret cannot be empty in production")
	// }

	return &cfg, nil
}
