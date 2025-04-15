// Copyright 2025 raph-abdul
// Licensed under the Apache License, Version 2.0.
// Visit http://www.apache.org/licenses/LICENSE-2.0 for details

// Package auth /youGo/internal/auth/jwt.go
package auth

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"time"

	"github.com/golang-jwt/jwt/v5" // Using v5
)

// CustomClaims defines the structure of the JWT claims used in this application.
// It includes standard registered claims and custom claims like UserID.
type CustomClaims struct {
	UserID uuid.UUID `json:"user_id"`
	// Add other custom claims if needed (e.g., role, email)
	// Role string `json:"role,omitempty"`
	jwt.RegisteredClaims // Embeds standard claims like ExpiresAt, IssuedAt, Subject etc.
}

// GenerateAccessToken creates a new JWT access token for the given user ID.
func GenerateAccessToken(userID uuid.UUID, secret []byte, expiryDuration time.Duration) (string, error) {
	// Create the claims
	claims := CustomClaims{
		UserID: userID,
		// Role: role, // Add role if needed
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),                                    // Subject identifies the principal that is the subject of the JWT.
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiryDuration)), // Token expiration time
			IssuedAt:  jwt.NewNumericDate(time.Now()),                     // Time when the token was issued
			NotBefore: jwt.NewNumericDate(time.Now()),                     // Token is valid starting now
			// Issuer:    "you-go",                           // Optional: Issuer of the token
			// Audience:  []string{"you-go-clients"},           // Optional: Intended audience
		},
	}

	// Create a new token object, specifying signing method and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) // Using HMAC SHA-256

	// Sign the token with the secret key
	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return signedToken, nil
}

// GenerateRefreshToken creates a new JWT refresh token. Often has a longer expiry.
// Note: For higher security, consider using a separate secret for refresh tokens
// and potentially storing refresh token validity state server-side.
func GenerateRefreshToken(userID uuid.UUID, secret []byte, expiryDuration time.Duration) (string, error) {
	claims := CustomClaims{
		UserID: userID, // Keep UserID for identification
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiryDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			// Issuer:    "you-go-refresh", // Different issuer maybe?
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}
	return signedToken, nil
}

// ValidateToken parses and validates a JWT token string.
// It checks the signature, expiration, and other standard claims.
// Returns the custom claims if the token is valid, otherwise returns an error.
func ValidateToken(tokenString string, secret []byte) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the secret key for validation
		return secret, nil
	})

	if err != nil {
		// Handle specific JWT errors
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, fmt.Errorf("malformed token: %w", err)
		} else if errors.Is(err, jwt.ErrTokenExpired) {
			// Potentially handle refresh logic if ErrTokenExpired occurs
			return nil, fmt.Errorf("token expired: %w", err)
		} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, fmt.Errorf("token not valid yet: %w", err)
		} else if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return nil, fmt.Errorf("invalid token signature: %w", err)
		}
		// Other parsing errors
		return nil, fmt.Errorf("could not parse token: %w", err)
	}

	// Check if the token is valid and extract claims
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
