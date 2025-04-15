// Copyright 2025 raph-abdul
// Licensed under the Apache License, Version 2.0.
// Visit http://www.apache.org/licenses/LICENSE-2.0 for details

// Package auth /youGo/internal/auth/auth_service.go
package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	// Import domain package for User entity and Repository interface
	"youGo/internal/domain"
	// Import request DTO if Login needs it (though better to pass individual fields)
	"github.com/google/uuid" // Use consistent ID type
	"youGo/internal/api/request"
)

var (
	// Note: These errors might be better defined in the domain package if they are domain concepts
	// Or keep auth-specific ones like ErrInvalidCredentials here.
	// ErrUserNotFound is already in domain as ErrNotFound
	ErrInvalidCredentials = errors.New("invalid email or password")
)

// Service defines the interface for authentication operations.
// Register is REMOVED - it belongs in UserService.
type Service interface {
	Login(ctx context.Context, req *request.LoginRequest) (accessToken, refreshToken string, err error) // Accept DTO or email/password
	ValidateToken(tokenString string) (userID uuid.UUID, err error)                                     // Return uuid.UUID
	// RefreshToken(ctx context.Context, refreshTokenString string) (newAccessToken string, err error) // Optional
}

// authService implements the Service interface.
type authService struct {
	// CORRECT DEPENDENCY: Use the UserRepository interface from the domain package
	userRepo domain.UserRepository

	jwtSecret            []byte
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
	// No logger needed here? Or add if Login/Validate needs logging
}

// NewAuthService creates a new instance of the authentication service.
// It requires the user repository interface and JWT configuration values.
func NewAuthService(
	// CORRECT DEPENDENCY: Accept the interface
	repo domain.UserRepository,
	jwtSecret []byte,
	accessDuration time.Duration,
	refreshDuration time.Duration,
) Service { // Return the Service interface
	if len(jwtSecret) == 0 {
		panic("JWT secret cannot be empty")
	}
	return &authService{
		userRepo:             repo, // Store the interface implementation
		jwtSecret:            jwtSecret,
		accessTokenDuration:  accessDuration,
		refreshTokenDuration: refreshDuration,
	}
}

// Register method is REMOVED from AuthService.
// It is now implemented in internal/service/user_service.go

// Login handles user login attempts.
func (s *authService) Login(ctx context.Context, req *request.LoginRequest) (accessToken, refreshToken string, err error) {
	// 1. Find user by email using the UserRepository interface
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// Return specific auth error
			return "", "", ErrInvalidCredentials
		}
		// Log underlying error? Need logger dependency if so.
		return "", "", fmt.Errorf("error finding user by email: %w", err)
	}

	// 2. Check password hash using the helper from this package
	if !CheckPasswordHash(req.Password, user.PasswordHash) {
		return "", "", ErrInvalidCredentials
	}

	// 3. Generate tokens using helpers from this package
	// userID := user.ID // ID is uuid.UUID now

	accessToken, err = GenerateAccessToken(user.ID, s.jwtSecret, s.accessTokenDuration) // Pass uuid.UUID directly if generator handles it, or user.ID.String() if it expects string
	if err != nil {
		// Log error?
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err = GenerateRefreshToken(user.ID, s.jwtSecret, s.refreshTokenDuration) // Pass uuid.UUID or user.ID.String()
	if err != nil {
		// Log error?
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// 4. Return tokens
	return accessToken, refreshToken, nil
}

// ValidateToken is used by middleware to check token validity and get user ID.
func (s *authService) ValidateToken(tokenString string) (userID uuid.UUID, err error) {
	claims, err := ValidateToken(tokenString, s.jwtSecret) // Use helper from this package
	if err != nil {
		return uuid.Nil, fmt.Errorf("token validation failed: %w", err) // Return Nil UUID on error
	}

	// Token is valid, return the UserID stored in the Subject/Custom claim
	// Try parsing from custom claim first, then Subject
	userIDStr := claims.UserID
	if userIDStr == uuid.Nil {
		userIDStr = claims.UserID
	}

	if userID == uuid.Nil {
		return uuid.Nil, errors.New("invalid token: missing user identifier in claims")
	}

	return userID, nil
}

// RefreshToken implementation would go here if needed...
