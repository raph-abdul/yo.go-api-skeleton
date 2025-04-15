// Copyright 2025 raph-abdul
// Licensed under the Apache License, Version 2.0.
// Visit http://www.apache.org/licenses/LICENSE-2.0 for details

// Package request /youGo/internal/api/request/auth_request.go
package request

// LoginRequest defines the structure for a login request body.
// Validation tags depend on the validator library used (e.g., go-playground/validator).
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"` // Example: require password, min 8 chars
}

// SignupRequest defines the structure for a user registration request body.
type SignupRequest struct {
	Name     string `json:"name" validate:"required,min=2"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	// Optional: Add password confirmation if needed by your logic/UI
	// PasswordConfirm string `json:"password_confirm" validate:"required,eqfield=Password"`
}

// Add other auth-related request structs if needed, e.g., for password reset, token refresh, etc.
// type RefreshTokenRequest struct {
//     RefreshToken string `json:"refresh_token" validate:"required"`
// }
//
// type ForgotPasswordRequest struct {
//    Email string `json:"email" validate:"required,email"`
// }
//
// type ResetPasswordRequest struct {
//     Token           string `json:"token" validate:"required"`
//     Password        string `json:"password" validate:"required,min=8"`
//     PasswordConfirm string `json:"password_confirm" validate:"required,eqfield=Password"`
// }
