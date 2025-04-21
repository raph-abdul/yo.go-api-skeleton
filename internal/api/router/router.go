// Copyright 2025 raph-abdul
// Licensed under the Apache License, Version 2.0.
// Visit http://www.apache.org/licenses/LICENSE-2.0 for details

// Package router /youGo/internal/api/router/router.go
package router

import (
	echoSwagger "github.com/swaggo/echo-swagger"
	// Standard library imports
	"net/http"

	// External dependencies
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	_ "youGo/docs"

	"youGo/internal/api/handler"
)

// Dependencies holds the required components for setting up routes.
// This struct is populated in main.go and passed to SetupRoutes.
type Dependencies struct {
	Logger         *zap.Logger
	AuthMiddleware echo.MiddlewareFunc // The JWTAuth middleware instance configured in main.go

	// Handlers
	AuthHandler *handler.AuthHandler
	UserHandler *handler.UserHandler
	// Add other handlers here, e.g.:
	// ProductHandler *producthandler.ProductHandler
}

// SetupRoutes configures all the application routes and applies middleware.
func SetupRoutes(e *echo.Echo, deps Dependencies) {

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// --- Health Check / Root ---
	// A simple endpoint to check if the service is running
	e.GET("/", func(c echo.Context) error {
		// You can return status or version information here
		return c.JSON(http.StatusOK, map[string]string{
			"service": "youGo",
			"status":  "ok",
			"version": "1.0.0", // Consider adding dynamic versioning
		})
	})

	// --- API Versioning Group ---
	// Grouping routes under /api/v1 for future versioning
	api := e.Group("/api/v1")

	// --- Authentication Routes (Public) ---
	// These routes generally do not require the user to be logged in.
	authGroup := api.Group("/auth")
	{ // Braces for visual grouping, optional
		deps.Logger.Debug("Setting up /auth routes")
		authGroup.POST("/login", deps.AuthHandler.Login)
		authGroup.POST("/signup", deps.AuthHandler.Register)

		// Add other public auth routes if implemented:
		// authGroup.POST("/refresh", deps.AuthHandler.RefreshToken) // Needs careful consideration about auth state
		// authGroup.POST("/forgot-password", deps.AuthHandler.ForgotPassword)
		// authGroup.POST("/reset-password", deps.AuthHandler.ResetPassword)
	}

	// --- User Routes (Protected) ---
	// Routes related to the logged-in user's own data.
	// Apply the authentication middleware to this group.
	//userGroup := api.Group("/users")
	//userGroup.Use(deps.AuthMiddleware) // Apply JWT authentication to all routes below
	//{
	//	deps.Logger.Debug("Setting up protected /users routes")
	//	// Endpoint for the logged-in user to get their own profile
	//	userGroup.GET("/me", deps.UserHandler.GetMe)
	//	// Endpoint for the logged-in user to update their own profile
	//	userGroup.PUT("/me", deps.UserHandler.UpdateMe)
	//	// Endpoint for the logged-in user to change their password
	//	userGroup.PUT("/me/password", deps.UserHandler.ChangeMyPassword)
	//}

	// --- Admin User Routes (Example - Protected with Auth + Admin Middleware) ---
	// Routes for administrators managing users. Requires additional role checking.
	// NOTE: This requires an additional Admin middleware not defined yet.
	/*
	   adminUserGroup := api.Group("/admin/users")
	   adminUserGroup.Use(deps.AuthMiddleware) // Must be logged in
	   // adminUserGroup.Use(apimiddleware.RequireAdmin(deps.Logger)) // Apply admin check middleware <<<< NEEDS IMPLEMENTATION
	   {
	       deps.Logger.Debug("Setting up protected /admin/users routes")
	       adminUserGroup.GET("", deps.UserHandler.ListUsers) // Handler method needs implementation
	       adminUserGroup.POST("", deps.UserHandler.CreateUser) // Handler method needs implementation
	       adminUserGroup.GET("/:id", deps.UserHandler.GetUserByID) // Handler method needs implementation
	       adminUserGroup.PUT("/:id", deps.UserHandler.UpdateUser) // Handler method needs implementation
	       adminUserGroup.DELETE("/:id", deps.UserHandler.DeleteUser) // Handler method needs implementation
	   }
	*/

	// --- Other Resource Routes (Example: Products) ---
	/*
	   productGroup := api.Group("/products")
	   {
	       deps.Logger.Debug("Setting up /products routes")
	       // Public endpoints to view products
	       productGroup.GET("", deps.ProductHandler.ListProducts)
	       productGroup.GET("/:id", deps.ProductHandler.GetProductByID)

	       // Protected endpoints to manage products
	       // Apply middleware directly to routes or group specific sub-routes
	       productGroup.POST("", deps.ProductHandler.CreateProduct, deps.AuthMiddleware)
	       productGroup.PUT("/:id", deps.ProductHandler.UpdateProduct, deps.AuthMiddleware)
	       productGroup.DELETE("/:id", deps.ProductHandler.DeleteProduct, deps.AuthMiddleware)
	       // Example: Adding a review might also require auth
	       // productGroup.POST("/:id/reviews", deps.ProductHandler.AddReview, deps.AuthMiddleware)
	   }
	*/

	deps.Logger.Info("✅ API routes configured successfully")
}
