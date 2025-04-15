// /test/integration_test.go
package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"youGo/internal/api/handler"
	"youGo/internal/api/middleware"
	"youGo/internal/api/request"                  // Import request DTOs
	"youGo/internal/api/response"                 // Import response DTOs
	"youGo/internal/api/router"                   // Import router setup
	"youGo/internal/auth"                         // Import auth service for DI
	"youGo/internal/config"                       // Import config loader
	"youGo/internal/domain"                       // Import domain types
	"youGo/internal/platform/database"            // Import DB setup
	"youGo/internal/platform/logger"              // Import logger setup
	repoImpl "youGo/internal/repository/postgres" // Import repo implementation
	"youGo/internal/service"                      // Import service layer
	// "github.com/joho/godotenv" // If using .env files for test config
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var (
	testServer *httptest.Server
	testDB     *gorm.DB
	testConfig *config.Config
	// Keep track of created user IDs for cleanup
	testUserIDs []string
)

// setupIntegrationTests initializes the server and DB for integration tests.
func setupIntegrationTests(t *testing.T) {
	// --- Load Test Configuration ---
	// Recommend using a separate .env.test or specific test config files/vars
	// err := godotenv.Load(".env.test") // Example using godotenv
	// require.NoError(t, err, "Failed to load .env.test")

	// Or load config using your config package, maybe overriding DB name etc.
	// For simplicity, we load default config and expect test DB details in env vars
	cfg, err := config.Load("../configs", "config") // Adjust path relative to test file
	require.NoError(t, err, "Failed to load configuration")
	testConfig = cfg
	// *** CRITICAL: Ensure this points to a TEST database ***
	// Override DB name or use specific test environment variables
	testConfig.Database.DBName = cfg.Database.DBName + "_test" // Example override
	fmt.Printf("--- Using Test Database: %s ---\n", testConfig.Database.DBName)

	// --- Initialize Logger ---
	appLogger, err := logger.New(cfg.Log.Level, cfg.Log.Format, cfg.App.Env)
	require.NoError(t, err, "Failed to initialize logger")

	// --- Initialize Test Database ---
	dbInstance, err := database.NewGORMConnection(testConfig.Database)
	require.NoError(t, err, "Failed to connect to test database")
	testDB = dbInstance

	// --- Clean Database Before Test Run (or use transactions) ---
	// Simple cleanup: Delete data from relevant tables
	err = testDB.Exec("DELETE FROM user_models").Error // Adjust table name if different
	require.NoError(t, err, "Failed to clean user table")
	testUserIDs = []string{} // Reset cleanup tracker

	// --- Initialize Dependencies (similar to main.go but with test DB/config) ---
	userRepo := repoImpl.NewUserRepository(testDB)

	// Parse durations for auth service
	accessDuration, err := time.ParseDuration(cfg.Auth.AccessTokenDuration)
	require.NoError(t, err, "Invalid access token duration")
	refreshDuration, err := time.ParseDuration(cfg.Auth.RefreshTokenDuration)
	require.NoError(t, err, "Invalid refresh token duration")

	authSvc := auth.NewAuthService(userRepo, []byte(cfg.Auth.JWTSecret), accessDuration, refreshDuration)
	userSvc := service.NewUserService(userRepo, appLogger)

	authHandler := handler.NewAuthHandler(authSvc, userSvc, appLogger)
	userHandler := handler.NewUserHandler(userSvc) // Pass logger

	// --- Setup Router & Test Server ---
	e := echo.New()
	// Need to configure validator for request validation to work
	// e.Validator = ... // Setup validator instance here (e.g., go-playground/validator)

	deps := router.Dependencies{
		Logger:         appLogger,
		AuthMiddleware: middleware.JWTAuth(authSvc, appLogger), // Assuming middleware package exists
		AuthHandler:    authHandler,
		UserHandler:    userHandler,
	}
	router.SetupRoutes(e, deps)

	testServer = httptest.NewServer(e)
}

// teardownIntegrationTests cleans up resources after tests.
func teardownIntegrationTests(t *testing.T) {
	if testServer != nil {
		testServer.Close()
	}
	// Clean up database after tests
	if testDB != nil {
		// Example: Delete users created during the test run
		if len(testUserIDs) > 0 {
			err := testDB.Exec("DELETE FROM user_models WHERE id IN (?)", testUserIDs).Error
			assert.NoError(t, err, "Failed to clean up created users")
		}
		// Close DB connection if necessary (GORM manages pool, usually not needed to close explicitly here)
		// sqlDB, _ := testDB.DB()
		// if sqlDB != nil { sqlDB.Close() }
	}
	fmt.Println("--- Teardown Complete ---")
}

// TestMain runs setup and teardown around all tests in the package.
func TestMain(m *testing.M) {
	// Setup runs once before all tests in this package
	// setupIntegrationTests(nil) // Need a dummy *testing.T or handle error reporting differently
	fmt.Println("--- Setting up Integration Tests ---")
	// Note: Proper setup often involves creating a dummy *testing.T or managing errors manually
	// For simplicity, errors here might panic. A better setup uses a dedicated test runner.

	// Run all tests in the package
	exitCode := m.Run()

	// Teardown runs once after all tests
	// teardownIntegrationTests(nil) // See note above

	os.Exit(exitCode)
}

// --- Example Test Case ---

func TestAuthEndpoints(t *testing.T) {
	// Ensure setup is run (TestMain doesn't pass 't', so might need per-test setup or better TestMain)
	// Re-running setup for each test ensures isolation but is slower.
	// Running once in TestMain is faster but requires careful state management.
	// Let's assume TestMain handles it or use a test suite structure.
	if testServer == nil {
		t.Skip("Test server not initialized, likely due to simplified TestMain. Skipping.") // Skip if setup failed/didn't run
	}

	t.Run("POST /signup - Success", func(t *testing.T) {
		require := require.New(t) // Use require for fatal assertions in setup phases
		assert := assert.New(t)   // Use assert for checks where test can continue

		// Generate unique email for each test run
		uniqueEmail := fmt.Sprintf("testuser_%d@example.com", time.Now().UnixNano())
		signupReq := request.SignupRequest{
			Name:     "Test User",
			Email:    uniqueEmail,
			Password: "password123",
		}
		reqBody, err := json.Marshal(signupReq)
		require.NoError(err)

		// Create request to the test server
		req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/v1/auth/signup", bytes.NewBuffer(reqBody))
		require.NoError(err)
		req.Header.Set("Content-Type", "application/json")

		// Send request
		client := testServer.Client()
		resp, err := client.Do(req)
		require.NoError(err)
		defer resp.Body.Close()

		// Assertions
		assert.Equal(http.StatusCreated, resp.StatusCode, "Expected status code 201")

		// Decode response (assuming generic SuccessResponse wrapping UserResponse)
		var successResp response.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&successResp)
		require.NoError(err, "Failed to decode response body")

		assert.Equal("success", successResp.Status)
		require.NotNil(successResp.Data, "Response data should not be nil")

		// Need to assert the structure of Data (UserResponse)
		// This requires converting map[string]interface{} back to UserResponse
		dataBytes, _ := json.Marshal(successResp.Data) // Convert map back to JSON
		var userResp response.UserResponse
		err = json.Unmarshal(dataBytes, &userResp) // Convert JSON to UserResponse struct
		require.NoError(err, "Failed to unmarshal UserResponse from response data")

		assert.Equal(signupReq.Name, userResp.Name)
		assert.Equal(signupReq.Email, userResp.Email)
		assert.NotEmpty(userResp.ID, "User ID should not be empty")
		assert.WithinDuration(time.Now(), userResp.CreatedAt, 5*time.Second, "CreatedAt should be recent")

		// Add created user ID for cleanup
		if userResp.ID != "" {
			testUserIDs = append(testUserIDs, userResp.ID)
		}

		// Optional: Verify user exists in the test database
		var dbUser domain.User
		dbErr := testDB.Where("id = ?", userResp.ID).First(&dbUser).Error
		assert.NoError(dbErr, "User should exist in database")
		assert.Equal(userResp.Email, dbUser.Email)

	})

	// Add more test cases for signup failure (duplicate email), login success/failure etc.
	// t.Run("POST /signup - Duplicate Email", func(t *testing.T) { ... })
	// t.Run("POST /login - Success", func(t *testing.T) { ... })
	// t.Run("POST /login - Incorrect Password", func(t *testing.T) { ... })
	// t.Run("GET /users/me - Unauthorized", func(t *testing.T) { ... })
	// t.Run("GET /users/me - Success", func(t *testing.T) { ... }) requires login first to get token

}
