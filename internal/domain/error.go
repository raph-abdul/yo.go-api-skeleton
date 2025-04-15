// Copyright 2025 raph-abdul
// Licensed under the Apache License, Version 2.0.
// Visit http://www.apache.org/licenses/LICENSE-2.0 for details

// Package domain /youGo/internal/domain/error.go
package domain

import (
	"fmt"
	"strings"
)

// --- Sentinel Errors ---
// These are common, fixed errors returned by services or repositories.
// Use errors.Is() to check for these.
var ErrNotFound = fmt.Errorf("domain: entity not found")
var ErrDuplicateEntry = fmt.Errorf("domain: duplicate entry")
var ErrPermissionDenied = fmt.Errorf("domain: permission denied")
var ErrInsufficientStock = fmt.Errorf("domain: insufficient stock")                       // Example if needed later
var ErrOptimisticLock = fmt.Errorf("domain: edit conflict, please refresh and try again") // Example for optimistic locking

// --- Custom Error Structs ---

// InvalidArgumentError indicates an error due to an invalid value for a specific argument.
type InvalidArgumentError struct {
	ArgumentName string
	Reason       string
}

// Error implements the error interface for InvalidArgumentError.
func (e *InvalidArgumentError) Error() string {
	return fmt.Sprintf("domain: invalid argument %q: %s", e.ArgumentName, e.Reason)
}

// ValidationError holds details about multiple validation failures.
type ValidationError struct {
	Failures map[string][]string // Map of field name to list of validation error messages
}

// NewValidationError creates a new ValidationError instance.
func NewValidationError() *ValidationError {
	return &ValidationError{Failures: make(map[string][]string)}
}

// Error implements the error interface for ValidationError.
func (e *ValidationError) Error() string {
	if len(e.Failures) == 0 {
		return "domain: validation error occurred (no details specified)"
	}
	// Build a summary string
	var sb strings.Builder
	sb.WriteString("domain: validation failed for fields: ")
	first := true
	for field := range e.Failures {
		if !first {
			sb.WriteString(", ")
		}
		sb.WriteString(field)
		first = false
	}
	return sb.String()
}

// Add records a validation failure for a specific field.
func (e *ValidationError) Add(field, message string) {
	if e.Failures == nil {
		e.Failures = make(map[string][]string)
	}
	e.Failures[field] = append(e.Failures[field], message)
}

// HasErrors returns true if any validation failures have been recorded.
func (e *ValidationError) HasErrors() bool {
	return len(e.Failures) > 0
}
