// Copyright 2025 raph-abdul
// Licensed under the Apache License, Version 2.0.
// Visit http://www.apache.org/licenses/LICENSE-2.0 for details

// Package response /youGo/internal/api/response/auth_response.go
package response

// LoginResponse defines the structure returned after a successful login.
type LoginResponse struct {
	// Optionally embed the full UserResponse DTO
	User *UserResponse `json:"user,omitempty"`

	// Tokens for accessing protected resources
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"` // Refresh token might be handled differently (e.g., httpOnly cookie) or omitted sometimes
	TokenType    string `json:"token_type"`              // Typically "Bearer"
	// ExpiresIn int `json:"expires_in,omitempty"` // Optional: Seconds until access token expiry
}

// RefreshTokenResponse defines the structure returned after successfully refreshing a token.
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"` // Typically "Bearer"
	// ExpiresIn int `json:"expires_in,omitempty"`
}

// SignupResponse: Often, a successful signup might just return the created user.
// In that case, you would return a UserResponse directly (wrapped in SuccessResponse).
// Example: return c.JSON(http.StatusCreated, response.NewSuccessResponse(userResponseDTO))
// Alternatively, if signup also immediately logs the user in, it might return a LoginResponse.
// Or, it could be a simple success message if email verification is needed:
// type SignupResponse struct {
//     Message string `json:"message"` // e.g., "Signup successful, please check your email to verify your account."
// }
