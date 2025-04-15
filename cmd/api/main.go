// Copyright 2025 raph-abdul

// @license.name  Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0

// @title           youGo
// @version         1.0
// @termsOfService  http://swagger.io/terms/

// @host      localhost:8080
// @BasePath  /api/v1
// @schemes http https
package main

import (
	"context"
	"errors"
	"fmt"
	stlog "log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"youGo/internal/api/handler"
	"youGo/internal/api/middleware"
	"youGo/internal/api/router"
	"youGo/internal/auth"
	"youGo/internal/config"
	"youGo/internal/platform/database"
	"youGo/internal/platform/logger"
	"youGo/internal/platform/validator"
	repoImpl "youGo/internal/repository/postgres"
	"youGo/internal/service"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load("./configs", "config")
	if err != nil {
		stlog.Fatalf("❌ Failed to load configuration: %v", err)
	}

	// 2. Initialize Logger
	appLogger, err := logger.New(cfg.Log.Level, cfg.Log.Format, cfg.App.Env)
	if err != nil {
		stlog.Fatalf("❌ Failed to initialize logger: %v", err)
	}
	defer func() { _ = appLogger.Sync() }()
	appLogger.Info("✅ Logger initialized", zap.String("level", cfg.Log.Level), zap.String("env", cfg.App.Env))
	appLogger.Info("Loaded config:", zap.Any("config", cfg)) // Log the entire config

	appLogger.Info("Config file path used:", zap.String("path", "./configs/config.prod.yaml")) // Add this
	appLogger.Info("Loaded config:", zap.Any("config", cfg))

	// 3. Setup Database Connection
	// Override config values with environment variables if present
	dbHost := os.Getenv("DB_HOST")
	if dbHost != "" {
		cfg.Database.Host = dbHost
	}

	dbPortStr := os.Getenv("DB_PORT")
	if dbPortStr != "" {
		dbPort, err := strconv.Atoi(dbPortStr)
		if err != nil {
			appLogger.Warn("Invalid DB_PORT format, using config value", zap.Error(err))
		} else {
			cfg.Database.Port = fmt.Sprintf("%d", dbPort)
		}
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser != "" {
		cfg.Database.User = dbUser
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword != "" {
		cfg.Database.Password = dbPassword
	}

	dbname := os.Getenv("DB_NAME")
	if dbname != "" {
		cfg.Database.DBName = dbname // Corrected field name
	}

	dbSslMode := os.Getenv("DB_SSLMODE")
	if dbSslMode != "" {
		cfg.Database.SSLMode = dbSslMode // Corrected field name
	}

	accessTokenDuration := os.Getenv("APP_AUTH_ACCESS_TOKEN_DURATION")
	if accessTokenDuration != "" {
		cfg.Auth.AccessTokenDuration = accessTokenDuration
	}

	refreshTokenDuration := os.Getenv("APP_AUTH_REFRESH_TOKEN_DURATION")
	if refreshTokenDuration != "" {
		cfg.Auth.RefreshTokenDuration = refreshTokenDuration
	}

	corsAllowedOrigins := os.Getenv("APP_SERVER_CORS_ALLOWED_ORIGINS")
	if corsAllowedOrigins != "" {
		cfg.Server.CORSAllowedOrigins = strings.Split(corsAllowedOrigins, ",") // Split the string
	}

	jwtSecret := os.Getenv("APP_AUTH_JWT_SECRET")
	if jwtSecret != "" {
		cfg.Auth.JWTSecret = jwtSecret // Corrected field name
	}

	port := os.Getenv("APP_SERVER_PORT")
	if port != "" {
		cfg.Server.Port = port // Corrected field name
	}

	//host := os.Getenv("DB_HOST")
	//port := os.Getenv("DB_PORT")
	//user := os.Getenv("DB_USER")
	//password := os.Getenv("DB_PASSWORD")
	//thename := os.Getenv("DB_NAME")
	//sslmode := os.Getenv("DB_SSLMODE")

	// Log the final connection details
	appLogger.Info("Connecting to database with:",
		zap.String("host", cfg.Database.Host),
		zap.String("port", cfg.Database.Port),
		zap.String("user", cfg.Database.User),
		zap.String("dbname", cfg.Database.DBName),
		zap.String("sslmode", cfg.Database.SSLMode),
		zap.String("APP_SERVER_CORS_ALLOWED_ORIGINS", os.Getenv("APP_SERVER_CORS_ALLOWED_ORIGINS")),
	)

	//fmt.Println("Connecting with:")
	//fmt.Println("  DB_HOST:", os.Getenv("DB_HOST"))
	//fmt.Println("  DB_PORT:", os.Getenv("DB_PORT"))
	//fmt.Println("  DB_USER:", os.Getenv("DB_USER"))
	//fmt.Println("  DB_PASSWORD:", os.Getenv("DB_PASSWORD")) // Be careful printing passwords in logs!
	//fmt.Println("  DB_NAME:", os.Getenv("DB_NAME"))
	//fmt.Println("  DB_SSLMODE:", os.Getenv("DB_SSLMODE"))
	//
	//connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
	//	host, port, user, password, thename, sslmode)
	//conn, err := pgx.Connect(context.Background(), connStr)

	dbInstance, err := database.NewGORMConnection(cfg.Database)
	if err != nil {
		appLogger.Fatal("❌ Failed to connect to database", zap.Error(err))
	}
	appLogger.Info("✅ Database connection pool established", zap.String("db_host", cfg.Database.Host))

	sqlDB, err := dbInstance.DB()
	if err != nil {
		appLogger.Warn("Could not get underlying *sql.DB from GORM", zap.Error(err))
	} else {
		defer func() {
			appLogger.Info("Closing database connection pool...")
			if err := sqlDB.Close(); err != nil {
				appLogger.Error("Error closing database connection", zap.Error(err))
			} else {
				appLogger.Info("✅ Database connection pool closed")
			}
		}()
	}

	// --- Optional: Run Migrations ---
	// NOTE: AutoMigrate is convenient for development but has limitations.
	// For production, using dedicated migration tools/scripts (like migrate.sh with SQL files)
	// is strongly recommended for better control and safety.
	// if cfg.Database.AutoMigrate { ... } // Keep commented out or remove

	appLogger.Info("Server config: Port", zap.String("port", cfg.Server.Port)) // Log after load

	// ... (database connection setup) ...

	// 4. Dependency Injection: Initialize Layers (Repositories -> Services -> Handlers)
	appLogger.Info("Initializing dependencies...")

	// --- Initialize Repositories (using specific implementation) ---
	// Replace with your actual repositories. Pass the GORM DB instance.
	userRepo := repoImpl.NewUserRepository(dbInstance)
	// productRepo := repoimpl.NewProductRepository(dbInstance) // Example
	// ... add other repositories ...

	appLogger.Debug("Repositories initialized") // Use Debug for verbose init steps

	appLogger.Info("Checking env vars:",
		zap.String("APP_AUTH_ACCESS_TOKEN_DURATION", os.Getenv("APP_AUTH_ACCESS_TOKEN_DURATION")),
		zap.String("APP_AUTH_REFRESH_TOKEN_DURATION", os.Getenv("APP_AUTH_REFRESH_TOKEN_DURATION")),
		zap.String("APP_SERVER_CORS_ALLOWED_ORIGINS", os.Getenv("APP_SERVER_CORS_ALLOWED_ORIGINS")),
	)

	// Parse token durations
	accessDuration, err := time.ParseDuration(cfg.Auth.AccessTokenDuration)
	if err != nil {
		stlog.Fatalf("❌ Invalid access token duration '%s': %v", cfg.Auth.AccessTokenDuration, err) // Use stlog before logger maybe
	}
	refreshDuration, err := time.ParseDuration(cfg.Auth.RefreshTokenDuration)
	if err != nil {
		stlog.Fatalf("❌ Invalid refresh token duration '%s': %v", cfg.Auth.RefreshTokenDuration, err)
	}

	appLogger.Info("Auth config after parsing:", zap.Duration("access_token_duration", accessDuration), zap.Duration("refresh_token_duration", refreshDuration))

	// --- Initialize Services ---
	// Pass repository interfaces and potentially logger or config values

	// Example: Auth service needs JWT secret from config
	authSvc := auth.NewAuthService(userRepo, []byte(cfg.Auth.JWTSecret), accessDuration, refreshDuration) // Passes repo interface
	userSvc := service.NewUserService(userRepo, appLogger)
	// ... add other services ...

	appLogger.Debug("Services initialized")

	// --- Initialize Handlers ---
	// Pass service interfaces and potentially logger

	authHandler := handler.NewAuthHandler(authSvc, userSvc, appLogger)

	// If user handler needs logger:
	userHandler := handler.NewUserHandler(userSvc) // Pass appLogger
	// If not, your original line is correct:
	// userHandler := userhandler.NewUserHandler(userSvc)

	// productHandler := producthandler.NewProductHandler(productSvc, appLogger) // Example
	// ... add other handlers ...
	appLogger.Info("✅ Dependencies initialized")

	// 5. Set up Echo instance
	e := echo.New()
	// Hide the Echo startup banner
	e.HideBanner = true
	// Using go-playground/validator:
	e.Validator = validator.NewValidator() // Implement this helper

	e.Use(echomiddleware.Logger()) // Add logger middleware
	e.Use(echomiddleware.CORS())

	// Consider setting custom JSON Serializer, Error Handler here if needed
	// e.HTTPErrorHandler = customErrorHandler.HandleError

	// --- Standard Middleware ---
	e.Use(echomiddleware.RequestID()) // <-- Add this line
	e.Use(echomiddleware.Recover())
	e.Use(middleware.RequestLogger(appLogger)) // Logger will now pick up request ID

	// --- Custom Global Middleware ---
	//e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{ // Example basic CORS
	//	//AllowOrigins: cfg.Auth.CORSAlowedOrigins, // Load allowed origins from config!
	//	AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete, http.MethodOptions},
	//	AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization}, // Adjust as needed
	//}))
	// Add other global middleware like rate limiting if required
	// e.Use(apimiddleware.RateLimiter(cfg.RateLimit)) // Example

	// Auth Middleware Instance (depends on AuthService)
	authMiddleware := middleware.JWTAuth(authSvc, appLogger)
	appLogger.Info("✅ Standard and custom middleware configured")

	// --- Configure Routing ---
	// Pass dependencies (handlers, middleware instance) to router setup

	// 6. Configure Routing (using internal/api/router)
	// Define a Dependencies struct in router package for cleaner passing
	routerDeps := router.Dependencies{
		Logger:         appLogger,
		AuthMiddleware: authMiddleware,
		AuthHandler:    authHandler,
		UserHandler:    userHandler,
	}

	router.SetupRoutes(e, routerDeps) // Pass Echo instance and dependencies struct
	appLogger.Info("✅ API routes configured")

	// 7. Start Server
	serverAddr := fmt.Sprintf(":%s", cfg.Server.Port) // Get port from config
	appLogger.Info("🚀 Starting server...", zap.String("address", serverAddr), zap.String("env", cfg.App.Env))
	appLogger.Info("Server Port from Config:", zap.String("port", cfg.Server.Port)) // Add this line
	zap.String("Server address:", serverAddr)                                       // Also log with standard log

	// Start server in a goroutine so it doesn't block graceful shutdown
	go func() {
		if err := e.Start(serverAddr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Fatal("❌ Server failed to start", zap.Error(err))
		}
	}()

	// 8. Graceful Shutdown
	// Wait for interrupt signal (Ctrl+C) or SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit // Block execution until signal is received

	appLogger.Info("🚦 Received shutdown signal. Starting graceful shutdown...")

	// Create a context with timeout for shutdown
	// Give active requests time to finish (e.g., 10-15 seconds)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Attempt graceful server shutdown
	if err := e.Shutdown(ctx); err != nil {
		appLogger.Fatal("❌ Server forced to shutdown", zap.Error(err))
	}

	appLogger.Info("✅ Server gracefully stopped")
}
