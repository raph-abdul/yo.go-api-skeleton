// Copyright 2025 raph-abdul
// Licensed under the Apache License, Version 2.0.
// Visit http://www.apache.org/licenses/LICENSE-2.0 for details

// Package response /youGo/internal/api/response/response.go
package response

import (
	"strings"

	"github.com/go-playground/validator/v10" // If using this validator for error details
	"github.com/labstack/echo/v4"            // To potentially handle echo.HTTPError
)

// SuccessResponse defines the structure for a standard successful API response.
type SuccessResponse struct {
	Status string      `json:"status"` // Typically "success"
	Data   interface{} `json:"data"`   // Holds the actual response DTO (e.g., UserResponse, LoginResponse)
}

// NewSuccessResponse creates a standard success response wrapper.
func NewSuccessResponse(data interface{}) SuccessResponse {
	return SuccessResponse{
		Status: "success",
		Data:   data,
	}
}

// --- Standard Error Response ---

// ErrorResponse defines the structure for a standard error API response.
type ErrorResponse struct {
	Status  string      `json:"status"`            // Typically "error" or "fail" (for validation)
	Message string      `json:"message"`           // User-friendly error message
	Details interface{} `json:"details,omitempty"` // Optional: Detailed error info (e.g., validation failures)
}

// NewErrorResponse creates a standard error response wrapper.
func NewErrorResponse(message string, details interface{}) ErrorResponse {
	status := "error"
	// If details indicate validation failure, potentially use "fail" status
	if details != nil {
		// Basic check, you might refine this
		if _, ok := details.(map[string]string); ok {
			status = "fail"
		}
		if _, ok := details.([]string); ok {
			status = "fail"
		}
	}

	return ErrorResponse{
		Status:  status,
		Message: message,
		Details: details,
	}
}

// --- Specific Error Handling Helpers (Optional but recommended) ---

// NewError wraps a standard Go error in an ErrorResponse.
func NewError(err error) ErrorResponse {
	// Check if it's an echo.HTTPError to get status code and specific message
	if he, ok := err.(*echo.HTTPError); ok {
		// Use the message from echo.HTTPError if it's informative
		// You might want to customize this logic further
		msg := he.Message
		if s, ok := msg.(string); ok {
			return NewErrorResponse(s, nil)
		}
		return NewErrorResponse("An unexpected error occurred", nil)

	}
	// Generic error
	return NewErrorResponse(err.Error(), nil)
}

// NewValidationError formats validation errors into a consistent structure.
// This assumes you are using 'go-playground/validator/v10'. Adjust if using a different library.
func NewValidationError(err error) ErrorResponse {
	details := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErrors {
			fieldName := strings.ToLower(fieldErr.Field()) // Use lowercase field name
			// Provide more user-friendly messages based on the 'tag'
			switch fieldErr.Tag() {
			case "required":
				details[fieldName] = fieldName + " is required"
			case "email":
				details[fieldName] = fieldName + " must be a valid email address"
			case "min":
				details[fieldName] = fieldName + " must be at least " + fieldErr.Param() + " characters long"
			case "max":
				details[fieldName] = fieldName + " must be at most " + fieldErr.Param() + " characters long"
			case "eqfield":
				details[fieldName] = fieldName + " must match the " + strings.ToLower(fieldErr.Param()) + " field"
			case "nefield":
				details[fieldName] = fieldName + " must not match the " + strings.ToLower(fieldErr.Param()) + " field"
			default:
				details[fieldName] = fieldName + " is invalid (" + fieldErr.Tag() + ")"
			}
		}
	} else {
		// If it's not validator.ValidationErrors, return a generic message
		return NewErrorResponse("Validation failed", err.Error())
	}

	return NewErrorResponse("Validation failed", details)
}
