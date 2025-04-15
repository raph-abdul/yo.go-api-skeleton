// Copyright 2025 raph-abdul
// Licensed under the Apache License, Version 2.0.
// Visit http://www.apache.org/licenses/LICENSE-2.0 for details

// Package request /youGo/internal/api/request/user_request.go
package request

// UpdateUserProfileRequest defines the structure for updating user profile data.
// Typically, sensitive fields like email or role are not updated here.
// Use 'omitempty' if fields are optional

// CreateUserRequest (remains the same)
type CreateUserRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	// Role string `json:"role"`
}

type UpdateUserProfileRequest struct {
	Name string `json:"name" validate:"omitempty,min=2"` // Optional: If provided, must be at least 2 chars
	// Add other fields that users can update, e.g.:
	// Phone   string `json:"phone" validate:"omitempty,e164"` // Example: optional international phone number
	// Bio     string `json:"bio" validate:"omitempty,max=500"` // Example: optional bio, max 500 chars
}

// UpdateUserRequest (remains the same)
type UpdateUserRequest struct {
	Name     *string `json:"name,omitempty"`
	IsActive *bool   `json:"isActive,omitempty"`
	Role     *string `json:"role,omitempty"`
}

// ChangePasswordRequest defines the structure for a user changing their own password.
type ChangePasswordRequest struct {
	OldPassword        string `json:"old_password" validate:"required"`
	NewPassword        string `json:"new_password" validate:"required,min=8,nefield=OldPassword"` // Example: Min 8 chars, different from old password
	NewPasswordConfirm string `json:"new_password_confirm" validate:"required,eqfield=NewPassword"`
}

// CreateUserRequest defines the structure for an admin creating a user (example).
// Might be different from SignupRequest (e.g., password might be auto-generated or set differently).
// type CreateUserRequest struct {
//  Name  string `json:"name" validate:"required,min=2"`
//  Email string `json:"email" validate:"required,email"`
//  Role  string `json:"role" validate:"required,oneof=admin user guest"` // Example role validation
//  // Optionally set an initial password or trigger a password setup flow
// }

// Add other user-related request structs if needed.
