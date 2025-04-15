// Copyright 2025 raph-abdul
// Licensed under the Apache License, Version 2.0.
// Visit http://www.apache.org/licenses/LICENSE-2.0 for details

// Package database /youGo/internal/platform/database/database.go
package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"youGo/internal/config"
)

func NewGORMConnection(cfg config.Database) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)
	log.Println("Database DSN:", dsn) // Added logging

	gormLogLevel := gormlogger.Silent
	// Configure GORM logger
	// Set log level based on environment (e.g., Silent in prod, Info in dev)
	// Example: Set log level based on an environment variable or config field
	// if os.Getenv("APP_ENV") == "development" { // Or use cfg.App.Env
	//  gormLogLevel = gormlogger.Info
	// }

	newLogger := gormlogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer (log to stdout)
		gormlogger.Config{
			SlowThreshold:             time.Second * 2, // Slow SQL threshold (adjust as needed)
			LogLevel:                  gormLogLevel,    // Set log level
			IgnoreRecordNotFoundError: true,            // Don't log ErrRecordNotFound errors
			Colorful:                  true,            // Enable color (disable in prod if logging to files)
		},
	)

	// Connect to the database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger, // Use configured logger
		// Add other GORM configs if needed (e.g., naming strategy)
		// NamingStrategy: schema.NamingStrategy{ ... }
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		// GORM v2 should generally handle this, but good to check
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool parameters (load from config if available)
	// Example values, tune these based on expected load and DB resources
	sqlDB.SetMaxIdleConns(10)           // cfg.Database.MaxIdleConns
	sqlDB.SetMaxOpenConns(100)          // cfg.Database.MaxOpenConns
	sqlDB.SetConnMaxLifetime(time.Hour) // cfg.Database.ConnMaxLifetime (parse duration from config)

	// Optional: Ping the database to verify connection
	if err = sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection pool established successfully.") // Use standard log initially
	return db, nil
}

// Optional: RunMigrations function (if using AutoMigrate)
// func RunMigrations(db *gorm.DB) error {
//  log.Println("Running database migrations...")
//  // Add all your GORM models here
//  err := db.AutoMigrate(
//      &postgres.UserModel{}, // Assuming this is defined in repository/postgres
//      // &postgres.ProductModel{}, // Add other models
//  )
//  if err != nil {
//      return fmt.Errorf("database migration failed: %w", err)
//  }
//  log.Println("Database migrations completed.")
//  return nil
// }
