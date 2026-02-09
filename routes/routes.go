package routes

import (
	"todo-app/internal/handlers"
	"todo-app/internal/middleware"
	"todo-app/pkg/utils"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes with the provided handlers
func SetupRoutes(
	router *gin.Engine,
	authHandler *handlers.AuthHandler,
	todoHandler *handlers.TodoHandler,
	categoryHandler *handlers.CategoryHandler,
	jwtManager *utils.JWTManager,
) {
	// API group
	api := router.Group("/api")

	// Health check endpoint
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Todo API is running",
		})
	})

	// Headers demo (shows reading a custom request header and returning a custom response header)
	api.GET("/headers", handlers.Headers)

	// Auth routes (public)
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Todo routes (protected)
	todos := api.Group("/todos")
	todos.Use(middleware.AuthMiddleware(jwtManager))
	{
		todos.POST("", todoHandler.CreateTodo)
		todos.GET("", todoHandler.GetTodos)
		todos.GET("/grouped", todoHandler.GetTodosGroupedByCategory)
		todos.GET("/:id", todoHandler.GetTodo)
		todos.PUT("/:id", todoHandler.UpdateTodo)
		todos.DELETE("/:id", todoHandler.DeleteTodo)
	}

	// Category routes (protected)
	// Note: Categories are auto-created when creating todos
	// These endpoints are for managing existing categories and sharing
	categories := api.Group("/categories")
	categories.Use(middleware.AuthMiddleware(jwtManager))
	{
		categories.GET("", categoryHandler.GetCategories)
		categories.GET("/:id", categoryHandler.GetCategory)
		categories.PUT("/:id", categoryHandler.UpdateCategory)
		categories.DELETE("/:id", categoryHandler.DeleteCategory)

		// Category sharing
		categories.POST("/:id/share", categoryHandler.ShareCategory)
		categories.GET("/:id/shares", categoryHandler.GetShares)
		categories.PUT("/:id/shares/:user_id", categoryHandler.UpdateSharePermission)
		categories.DELETE("/:id/shares/:user_id", categoryHandler.UnshareCategory)
	}
}
