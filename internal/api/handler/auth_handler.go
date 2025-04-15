// Copyright 2025 raph-abdul
// Licensed under the Apache License, Version 2.0.
// Visit http://www.apache.org/licenses/LICENSE-2.0 for details

// Package handler /youGo/internal/api/handler/auth_handler.go
package handler

import (
	"youGo/internal/api/request"  // Request DTOs
	"youGo/internal/api/response" // Response DTOs
	"youGo/internal/auth"         // Interfaces for Auth Service
	"youGo/internal/domain"       // Import for potential domain-specific errors
	"youGo/internal/service"      // Interfaces for Services lives here

	"errors" // For error checking (errors.Is)

	"github.com/labstack/echo/v4"
	"go.uber.org/zap" // Zap logger
	"net/http"
)

// AuthHandler handles HTTP requests related to authentication.

type AuthHandler struct {
	authService auth.Service        // Interface for auth operations (Login, Refresh, etc.)
	userService service.UserService // Interface for user operations (Register)
	logger      *zap.Logger
}

// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(authSvc auth.Service, userSvc service.UserService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authSvc,
		userService: userSvc,
		logger:      logger.Named("AuthHandler"),
	}
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        register body request.CreateUserRequest true "User Registration Details"  // Correct: Matches code binding request.CreateUserRequest
// @Success      201 {object} response.SuccessResponse{data=response.UserResponse} "User registered successfully" // Correct: Matches code returning wrapped response.UserResponse (assuming registerResp is compatible)
// @Failure      400 {object} response.ErrorResponse "Invalid input data (validation error)"
// @Failure      409 {object} response.ErrorResponse "User with this email already exists"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /auth/signup [post]
func (h *AuthHandler) Register(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(request.CreateUserRequest)

	// 1. Bind Request Body (now binds directly to request.CreateUserRequest)
	if err := c.Bind(req); err != nil {
		h.logger.Warn("Failed to bind registration request", zap.Error(err))
		// Consider using domain.NewErrorResponse if defined and appropriate
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format: "+err.Error())
	}

	// 2. Validate Request Data (ensure validation tags exist on request.CreateUserRequest)
	if err := c.Validate(req); err != nil {
		h.logger.Warn("Registration request validation failed", zap.Error(err))
		// validationDetails := response.NewValidationError(err) // This might need adjustment if error format changes
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "Input validation failed") // Keep simple or adjust error reporting
	}

	// 3. Call Service Layer - 'req' is now the correct type (*request.CreateUserRequest)
	registerResp, err := h.userService.Create(ctx, req)
	if err != nil {
		// 4. Handle Service Errors (remains mostly the same, ensure errors match domain errors)
		switch {
		case errors.Is(err, domain.ErrDuplicateEntry):
			h.logger.Warn("Registration attempt failed: user already exists", zap.String("email", req.Email))
			// Consider using domain.NewErrorResponse
			return echo.NewHTTPError(http.StatusConflict, domain.ErrDuplicateEntry.Error()) // Use domain error message
		default:
			h.logger.Error("Internal error during user registration", zap.Error(err), zap.String("email", req.Email))
			// Consider using response.NewErrorResponse
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user")
		}
	}

	// 5. Return Successful Response (registerResp is now *response.UserResponse)
	// Remove the response.NewUserResponse mapping if registerResp is already the correct structure
	// userDto := response.NewUserResponse(registerResp) // MAYBE NOT NEEDED if registerResp is already response.UserResponse
	h.logger.Info("User registered successfully", zap.String("userID", registerResp.ID)) // Log ID from service response DTO
	return c.JSON(http.StatusCreated, response.NewSuccessResponse(registerResp))         // Wrap service response DTO
}

// Login godoc
// @Summary      Log in a user
// @Description  Authenticates a user and returns access/refresh tokens.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        login body request.LoginRequest true "Login credentials" // Correct: Matches code binding request.LoginRequest
// @Success      200 {object} response.SuccessResponse{data=response.LoginResponse} "Login successful, tokens provided" // Corrected: Matches code returning wrapped response.LoginResponse
// @Failure      400 {object} response.ErrorResponse "Invalid input data (validation error)" // Note: Code returns 422 for validation
// @Failure      401 {object} response.ErrorResponse "Invalid credentials"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(request.LoginRequest) // Use pointer
	// 1. Bind Request Body
	if err := c.Bind(req); err != nil {
		h.logger.Warn("Failed to bind login request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format: "+err.Error()) // Include binding error detail
	}

	// 2. Validate Request Data
	if err := c.Validate(req); err != nil {
		h.logger.Warn("Login request validation failed", zap.Error(err))
		validationDetails := response.NewValidationError(err) // Assume returns map[string]string or similar
		// Option 1: Pass details if your custom error handler can use them
		return echo.NewHTTPError(http.StatusUnprocessableEntity, validationDetails)

		// Option 2: Convert details to a simple string message (loses structure)
		// return echo.NewHTTPError(http.StatusUnprocessableEntity, fmt.Sprintf("Validation failed: %v", validationDetails))
		// Option 3: Keep simple message if details aren't critical for the client
		// return echo.NewHTTPError(http.StatusUnprocessableEntity, "Input validation failed") // Keep generic but clear
	}

	// 3. Call Service Layer
	// Capture all return values from the authService.Login
	accessToken, refreshToken, err := h.authService.Login(ctx, req) // <-- Fix: Capture 3 values

	if err != nil {
		switch {
		// TODO: Confirm domain/service error before using as condition.
		case errors.Is(err, auth.ErrInvalidCredentials): // Use error from auth package
			h.logger.Warn("Login attempt failed: invalid credentials", zap.String("email", req.Email))
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
		default:
			h.logger.Error("Internal error during user login", zap.Error(err), zap.String("email", req.Email))
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to login due to an internal error")
		}
	}

	// 4. Construct the successful response DTO using the returned tokens
	loginResp := response.LoginResponse{
		// User field is optional - depends if your Login service method also returns user details
		// User: &response.UserResponse{ ID: userID, ... }, // Populate if needed
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer", // Standard token type
		// 200 OK
	}

	// 5. Return Successful Response
	// Assuming your LoginResponse doesn't include UserID directly, maybe log differently or fetch user details if needed for logging
	h.logger.Info("User logged in successfully", zap.String("email", req.Email)) // Log email instead of UserID if not readily available
	return c.JSON(http.StatusOK, response.NewSuccessResponse(loginResp))
}
