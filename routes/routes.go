package routes

import (
	"todo-app/controllers"
	"todo-app/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine) {
	// API group
	api := router.Group("/api")

	// Health check endpoint
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Todo API is running",
		})
	})

	// Auth routes (public)
	auth := api.Group("/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
	}

	// Todo routes (protected)
	todos := api.Group("/todos")
	todos.Use(middlewares.AuthMiddleware())
	{
		todos.POST("", controllers.CreateTodo)
		todos.GET("", controllers.GetTodos)
		todos.GET("/:id", controllers.GetTodo)
		todos.PUT("/:id", controllers.UpdateTodo)
		todos.DELETE("/:id", controllers.DeleteTodo)
	}
}
