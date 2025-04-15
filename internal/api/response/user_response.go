// Copyright 2025 raph-abdul
// Licensed under the Apache License, Version 2.0.
// Visit http://www.apache.org/licenses/LICENSE-2.0 for details

// Package response /youGo/internal/api/response/user_response.go
package response

import (
	"time"
	"youGo/internal/domain"
	// "github.com/google/uuid" // Needed if ID is UUID
)

// UserResponse represents user data returned by the API. ID is string for JSON
type UserResponse struct {
	ID        string    `json:"id"` // String for JSON compatibility
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	IsActive  bool      `json:"isActive"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	// Role string    `json:"role,omitempty"`
}

// NewUserResponse creates a UserResponse DTO from a domain.User object.
// It takes the *domain.User returned by the service layer.
func NewUserResponse(user *domain.User) UserResponse {
	if user == nil {
		// Return an empty struct or handle as appropriate
		return UserResponse{}
	}
	return UserResponse{
		ID:        user.ID.String(), // String for JSON compatibility
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		// Role:      user.Role, // Map other relevant fields
	}
}

// NewUserListResponse creates a slice of UserResponse DTOs from a slice of domain.User objects.
// func NewUserListResponse(users []*domain.User) []UserResponse {
//  list := make([]UserResponse, len(users))
//  for i, u := range users {
//      list[i] = NewUserResponse(u)
//  }
//  return list
// }
