// Copyright 2025 raph-abdul
// Licensed under the Apache License, Version 2.0.
// Visit http://www.apache.org/licenses/LICENSE-2.0 for details

// Package middleware /youGo/internal/api/middleware/auth_middleware.go
package middleware

import (
	"net/http"
	"strings"

	"github.com/google/uuid" // Use UUID for consistency
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"youGo/internal/auth" // Import your auth service package
)

// UserIDContextKey is the key used to store the authenticated user's ID in the Echo context.
// Using a custom type avoids collisions.
type contextKey string

const UserIDContextKey = contextKey("userID")

// JWTAuth creates an Echo middleware function that verifies a JWT token.
// It expects the token in the "Authorization: Bearer <token>" header.
// Dependencies (AuthService, Logger) are passed in.
func JWTAuth(authSvc auth.Service, log *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				log.Warn("AuthMiddleware: Missing authorization header")
				// Return standard Echo HTTP error
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing or malformed authorization header")
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				log.Warn("AuthMiddleware: Invalid Authorization header format", zap.String("header", authHeader))
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing or malformed authorization header")
			}

			tokenString := parts[1]
			if tokenString == "" {
				log.Warn("AuthMiddleware: Empty token in authorization header")
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing or malformed authorization header")
			}

			// Validate the token using the auth service
			// Assumes ValidateToken returns userID (uuid.UUID) and error
			userID, err := authSvc.ValidateToken(tokenString)
			if err != nil {
				log.Warn("AuthMiddleware: Token validation failed", zap.Error(err))
				// Check for specific token errors if needed (e.g., expired)
				// For now, return a generic unauthorized error
				// Consider mapping specific validation errors to different messages/codes
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token") // Use error message if suitable: err.Error()
			}

			// --- Token is valid ---
			log.Debug("AuthMiddleware: Token validated successfully", zap.String("userID", userID.String()))

			// Store the user ID (as uuid.UUID) in the Echo context
			c.Set(string(UserIDContextKey), userID) // Use string(key) when setting

			// Proceed to the next handler in the chain
			return next(c)
		}
	}
}

// GetUserIDFromContext is a helper function to retrieve the user ID from the Echo context.
// Call this from your handlers that run *after* the JWTAuth middleware.
func GetUserIDFromContext(c echo.Context) (uuid.UUID, bool) {
	// Retrieve using string(key)
	val := c.Get(string(UserIDContextKey))
	if val == nil {
		return uuid.Nil, false // Not found
	}

	userID, ok := val.(uuid.UUID) // Type assertion
	if !ok {
		return uuid.Nil, false // Found but wrong type somehow
	}

	return userID, true
}
