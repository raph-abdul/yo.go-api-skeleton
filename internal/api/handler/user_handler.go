// Copyright 2025 raph-abdul
// Licensed under the Apache License, Version 2.0.
// Visit http://www.apache.org/licenses/LICENSE-2.0 for details

// Package handler /youGo/internal/api/handler/user_handler.go
package handler

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"youGo/internal/domain"

	"youGo/internal/api/request"
	// --- Internal Imports ---
	"youGo/internal/api/response"
	"youGo/internal/service"

	"github.com/google/uuid"
)

// UserHandler handles user resource related HTTP requests.
type UserHandler struct {
	userService service.UserService // Dependency: UserService interface
}

// NewUserHandler creates a new instance of UserHandler.
func NewUserHandler(userSvc service.UserService) *UserHandler {
	return &UserHandler{
		userService: userSvc,
	}
}

// CreateUser godoc
// @Summary      Create a new user
// @Description  Adds a new user to the system. Typically used by administrators.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        user body request.CreateUserRequest true "User details for creation" // Request uses domain type - OK
// @Success      201 {object} response.UserResponse "User created successfully"       // UPDATED: Reference response DTO
// @Failure      400 {object} response.ErrorResponse "Invalid input data"             // UPDATED: Reference response DTO
// @Failure      409 {object} response.ErrorResponse "User conflict (e.g., email exists)" // UPDATED: Reference response DTO
// @Failure      500 {object} response.ErrorResponse "Internal server error"          // UPDATED: Reference response DTO
// @Router       /users [post]
// @Security     ApiKeyAuth
func (h *UserHandler) CreateUser(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(request.CreateUserRequest)

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body", http.StatusBadRequest))
	}

	// Optional: Validate DTO
	// if err := c.Validate(&req); err != nil { ... }

	domainUser, err := h.userService.Create(ctx, req)
	if err != nil {
		// Handle errors (e.g., conflict, validation, internal)
		// Map service.ErrUserAlreadyExists -> http.StatusConflict
		// Map service.ErrValidation -> http.StatusBadRequest
		// ... etc ...
		c.Logger().Error("Create user service call failed:", err)
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Failed to create user", http.StatusInternalServerError)) // Placeholder
	}

	responseDto := response.NewSuccessResponse(domainUser)

	return c.JSON(http.StatusCreated, responseDto)
}

// GetUserByID godoc
// @Summary      Get a user by ID
// @Description  Retrieves details for a specific user by their ID.
// @Tags         Users
// @Produce      json
// @Param        id path string true "User ID" format(uuid) // Added format(uuid)
// @Success      200 {object} response.UserResponse "User details found"      // Corrected: domain. prefix
// @Failure      400 {object} response.ErrorResponse "Invalid User ID format" // Corrected: domain. prefix
// @Failure      404 {object} response.ErrorResponse "User not found"         // Corrected: domain. prefix
// @Failure      500 {object} response.ErrorResponse "Internal server error"  // Corrected: domain. prefix
// @Router       /users/{id} [get]
// @Security     ApiKeyAuth
func (h *UserHandler) GetUserByID(c echo.Context) error {
	ctx := c.Request().Context()
	idStr := c.Param("id") // Get ID string from URL path

	// 1. Parse the string into a UUID object
	parsedUUID, err := uuid.Parse(idStr) // Correctly parse to uuid.UUID type
	if err != nil {
		// Parsing failed, means idStr is not a valid UUID format
		return c.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid user ID format", http.StatusBadRequest))
	}

	// 2. Call service with the parsed uuid.UUID
	userResp, err := h.userService.GetByID(ctx, parsedUUID) // Pass the uuid.UUID type
	if err != nil {
		// Handle errors (map domain.ErrNotFound to http.StatusNotFound, etc.)
		if err == domain.ErrNotFound { // Example check
			return c.JSON(http.StatusNotFound, response.NewErrorResponse("User not found", http.StatusNotFound))
		}
		c.Logger().Error("Get user by ID service call failed:", err)
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Failed to retrieve user", http.StatusInternalServerError))
	}

	// 3. Return response
	return c.JSON(http.StatusOK, userResp)
}

// UpdateUser godoc
// @Summary      Update a user
// @Description  Updates details for an existing user.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID" format(uuid) // Added format(uuid)
// @Param        user body request.UpdateUserRequest true "User details to update" // Corrected: domain. prefix
// @Success      200 {object} response.UserResponse "User updated successfully"    // Corrected: domain. prefix
// @Failure      400 {object} response.ErrorResponse "Invalid input data or User ID format" // Corrected: domain. prefix
// @Failure      404 {object} response.ErrorResponse "User not found"              // Corrected: domain. prefix
// @Failure      500 {object} response.ErrorResponse "Internal server error"       // Corrected: domain. prefix
// @Router       /users/{id} [put] // Or PATCH
// @Security     ApiKeyAuth
func (h *UserHandler) UpdateUser(c echo.Context) error {
	ctx := c.Request().Context()
	idStr := c.Param("id")
	req := new(request.UpdateUserRequest)

	// 1. Validate/Parse ID
	userID, err := uuid.Parse(idStr) // Correctly parse to uuid.UUID type

	if err != nil {
		return c.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid user ID format", http.StatusBadRequest))
	}

	// 2. Bind request body
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body", http.StatusBadRequest))
	}

	// 3. Optional: Validate DTO
	// if err := c.Validate(&req); err != nil { ... }

	// 4. Call service
	userResp, err := h.userService.Update(ctx, userID, req) // Pass ID and request DTO
	if err != nil {
		// Handle errors (e.g., not found, validation, internal)
		// Map service.ErrNotFound -> http.StatusNotFound
		// Map service.ErrValidation -> http.StatusBadRequest
		// ... etc ...
		c.Logger().Error("Update user service call failed:", err)
		return c.JSON(http.StatusInternalServerError, response.NewErrorResponse("Failed to update user", http.StatusInternalServerError)) // Placeholder
	}

	// 5. Return updated user data
	return c.JSON(http.StatusOK, userResp) // Use your UserResponse DTO
}

// Add other user-related handlers if needed (e.g., GetCurrentUser, ListUsers with pagination/filtering)
