// Copyright 2025 raph-abdul
// Licensed under the Apache License, Version 2.0.
// Visit http://www.apache.org/licenses/LICENSE-2.0 for details

// Package middleware /youGo/internal/api/middleware/logger_middleware.go
package middleware

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// RequestLogger creates an Echo middleware function that logs details about each request using Zap.
// It logs method, path, status, latency, IP, user agent, response size, and request ID.
func RequestLogger(log *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Process the request by calling the next handler
			err := next(c)

			// Log after the request is handled
			req := c.Request()
			res := c.Response()
			stop := time.Now()
			latency := stop.Sub(start)
			// Try to get Request ID (assuming RequestID middleware runs before this)
			requestID := res.Header().Get(echo.HeaderXRequestID)
			if requestID == "" {
				requestID = req.Header.Get(echo.HeaderXRequestID) // Fallback if not in response yet
			}

			// Prepare base log fields
			fields := []zap.Field{
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path),
				zap.Int("status", res.Status), // Get status after handler execution
				zap.Duration("latency", latency),
				zap.String("ip", c.RealIP()),
				zap.String("user_agent", req.UserAgent()),
				zap.Int64("response_size", res.Size),
				zap.String("request_id", requestID),
			}

			// Handle potential errors returned by handlers/downstream middleware
			statusCode := res.Status
			if err != nil {
				// Include the error in the log fields
				fields = append(fields, zap.Error(err))

				// Try to get status code from echo.HTTPError if available
				var httpError *echo.HTTPError
				if errors.As(err, &httpError) {
					statusCode = httpError.Code
					// Update status field if it differs from response status somehow
					if res.Status != statusCode {
						fields[2] = zap.Int("status", statusCode) // fields[2] is status field index
					}
				} else if statusCode < 400 {
					// If it's a non-HTTP error and status wasn't set to error level, default to 500
					statusCode = http.StatusInternalServerError
					fields[2] = zap.Int("status", statusCode) // Update status field index
				}
			}

			// Choose log level based on final status code
			switch {
			case statusCode >= 500:
				log.Error("Server error", fields...)
			case statusCode >= 400:
				log.Warn("Client error", fields...)
			default:
				log.Info("Request handled", fields...) // Use Info or Debug
			}

			// Return the original error so Echo's error handling can process it
			return err
		}
	}
}
