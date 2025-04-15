// Copyright 2025 raph-abdul
// Licensed under the Apache License, Version 2.0.
// Visit http://www.apache.org/licenses/LICENSE-2.0 for details

// Package domain /youGo/internal/domain/user.go
package domain

import (
	"context"
	"github.com/google/uuid" // Use consistent ID type
	"time"
)

// User represents the core user entity within the business domain.
// Agnostic of database or API implementation details.
type User struct {
	ID           uuid.UUID
	Name         string
	Email        string // Assumed to be unique
	PasswordHash string // The securely hashed password
	IsActive     bool
	Role         string // e.g., "admin", "customer"
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserRepository defines the contract for persistence operations related to Users.
// Implementations reside in the infrastructure/repository layer.
type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	// List(ctx context.Context /*, filters, pagination */) ([]*User, error) // Optional
}
