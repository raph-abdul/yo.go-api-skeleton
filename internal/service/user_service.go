// Copyright 2025 raph-abdul
// Licensed under the Apache License, Version 2.0.
// Visit http://www.apache.org/licenses/LICENSE-2.0 for details

// Package service /youGo/internal/repository/service/user_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"time"
	"youGo/internal/api/request"
	"youGo/internal/api/response"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"youGo/internal/auth"
	"youGo/internal/domain"
)

// UserService interface (signatures already use uuid.UUID)
type UserService interface {
	Create(ctx context.Context, req *request.CreateUserRequest) (*response.UserResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*response.UserResponse, error)
	Update(ctx context.Context, id uuid.UUID, req *request.UpdateUserRequest) (*response.UserResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// userService struct (remains the same)
type userService struct {
	userRepo domain.UserRepository
	logger   *zap.Logger
}

// NewUserService constructor (remains the same)
func NewUserService(repo domain.UserRepository, logger *zap.Logger) UserService {
	return &userService{
		userRepo: repo,
		logger:   logger,
	}
}

// Create implementation
func (s *userService) Create(ctx context.Context, req *request.CreateUserRequest) (*response.UserResponse, error) {
	s.logger.Debug("Attempting user creation", zap.String("email", req.Email))

	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	// ... (error checking remains same) ...
	if existingUser != nil {
		return nil, domain.ErrDuplicateEntry
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	// ... (error checking remains same) ...
	if err != nil {
		return nil, fmt.Errorf("internal server error processing creation")
	}

	newUUID, err := uuid.NewRandom()
	if err != nil {
		s.logger.Error("Failed to generate UUID for new user", zap.Error(err))
		return nil, fmt.Errorf("internal server error generating user ID")
	}

	now := time.Now().UTC()
	newUser := &domain.User{
		ID:           newUUID, // Assign uuid.UUID directly
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		IsActive:     true,
		Role:         "user",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err = s.userRepo.Create(ctx, newUser) // Pass user object with uuid.UUID ID
	// ... (error checking remains same) ...
	if err != nil {
		s.logger.Error("Failed to create user in repository", zap.String("email", req.Email), zap.Error(err))
		if errors.Is(err, domain.ErrDuplicateEntry) {
			return nil, domain.ErrDuplicateEntry
		}
		return nil, fmt.Errorf("failed to save user information")
	}

	s.logger.Info("User created successfully", zap.String("userID", newUser.ID.String()), zap.String("email", req.Email)) // Log string representation
	return mapUserToUserResponse(newUser), nil
}

// GetByID implementation
func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*response.UserResponse, error) {
	s.logger.Debug("Getting user by ID", zap.String("userID", id.String())) // Log string representation

	user, err := s.userRepo.FindByID(ctx, id) // Pass uuid.UUID directly to repo
	if err != nil {
		s.logger.Warn("Failed to get user by ID from repository", zap.String("userID", id.String()), zap.Error(err))
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed retrieving user data")
	}
	return mapUserToUserResponse(user), nil
}

// Update implementation
func (s *userService) Update(ctx context.Context, id uuid.UUID, req *request.UpdateUserRequest) (*response.UserResponse, error) {
	s.logger.Debug("Updating user profile", zap.String("userID", id.String())) // Log string representation

	user, err := s.userRepo.FindByID(ctx, id) // Pass uuid.UUID directly to repo
	if err != nil {
		// ... (error handling for not found remains same) ...
		s.logger.Warn("User not found for update", zap.String("userID", id.String()), zap.Error(err))
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed retrieving user for update")
	}

	updated := false
	// ... (logic for updating fields remains same) ...
	if req.Name != nil && *req.Name != user.Name {
		user.Name = *req.Name
		updated = true
	}
	if req.IsActive != nil && *req.IsActive != user.IsActive {
		user.IsActive = *req.IsActive
		updated = true
	}
	if req.Role != nil && *req.Role != user.Role {
		user.Role = *req.Role
		updated = true
	}

	if updated {
		user.UpdatedAt = time.Now().UTC()
		err = s.userRepo.Update(ctx, user) // Pass user object with uuid.UUID ID
		if err != nil {
			// ... (error handling remains same) ...
			s.logger.Error("Failed to update user profile in repository", zap.String("userID", id.String()), zap.Error(err))
			return nil, fmt.Errorf("failed saving updated user data")
		}
		s.logger.Info("User profile updated", zap.String("userID", id.String()))
	} else {
		s.logger.Debug("No changes detected for user update", zap.String("userID", id.String()))
	}

	return mapUserToUserResponse(user), nil
}

// Delete implementation
func (s *userService) Delete(ctx context.Context, id uuid.UUID) error {
	s.logger.Debug("Deleting user", zap.String("userID", id.String())) // Log string representation

	err := s.userRepo.Delete(ctx, id) // Pass uuid.UUID directly to repo
	if err != nil {
		s.logger.Error("Failed to delete user in repository", zap.String("userID", id.String()), zap.Error(err))
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("failed deleting user")
	}

	s.logger.Info("User deleted successfully", zap.String("userID", id.String()))
	return nil
}

// mapUserToUserResponse helper function
func mapUserToUserResponse(user *domain.User) *response.UserResponse {
	if user == nil {
		return nil
	}
	return &response.UserResponse{
		ID:        user.ID.String(), // Convert uuid.UUID to string for JSON response
		Name:      user.Name,
		Email:     user.Email,
		IsActive:  user.IsActive,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
