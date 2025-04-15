// Copyright 2025 raph-abdul
// Licensed under the Apache License, Version 2.0.
// Visit http://www.apache.org/licenses/LICENSE-2.0 for details

// Package auth /youGo/internal/auth/password.go
package auth

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword generates a bcrypt hash for the given password.
// It uses the default cost factor provided by the bcrypt library.
func HashPassword(password string) (string, error) {
	// GenerateFromPassword automatically handles salt generation
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// CheckPasswordHash compares a plaintext password with a stored bcrypt hash.
// Returns true if the password matches the hash, false otherwise.
func CheckPasswordHash(password, hash string) bool {
	// CompareHashAndPassword handles extracting the salt and cost from the hash
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	// err is nil if the password matches the hash
	return err == nil
}
