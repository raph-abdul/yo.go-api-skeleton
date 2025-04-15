// Copyright 2025 raph-abdul
// Licensed under the Apache License, Version 2.0.
// Visit http://www.apache.org/licenses/LICENSE-2.0 for details

// Package validator /youGo/internal/platform/validator/validator.go
package validator

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator wraps the validator library
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator creates a new instance of CustomValidator
func NewValidator() *CustomValidator {
	return &CustomValidator{validator: validator.New()}
}

// Validate implements the echo.Validator interface
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you can return echo.NewHTTPError to provide specific HTTP errors
		// Here, we return a generic validation error message, or the specific error
		// You might want to customize error formatting here later
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "Input validation failed: "+err.Error())
		// return err // Alternatively, return the raw validator error
	}
	return nil
}
