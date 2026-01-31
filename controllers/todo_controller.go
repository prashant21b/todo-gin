package controllers

import (
	"net/http"
	"strconv"
	"time"

	"todo-app/models"
	"todo-app/utils"

	"github.com/gin-gonic/gin"
)

// CreateTodoInput represents the create todo request body
type CreateTodoInput struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

// UpdateTodoInput represents the update todo request body
type UpdateTodoInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   *bool  `json:"completed"` // Pointer to distinguish between false and not provided
}

// CreateTodo handles creating a new todo with context support
func CreateTodo(c *gin.Context) {
	var input CreateTodoInput

	// Validate input
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Validation failed",
			"error":   err.Error(),
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// Create context with timeout for database operations
	ctx, cancel := utils.CreateRequestContext(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Create todo
	todo := models.Todo{
		Title:       input.Title,
		Description: input.Description,
		UserID:      userID.(uint),
	}

	if err := models.CreateTodo(ctx, &todo); err != nil {
		// Check if context was cancelled
		if ctx.Err() != nil {
			c.JSON(http.StatusRequestTimeout, gin.H{
				"success": false,
				"message": "Request timeout",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create todo",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Todo created successfully",
		"data":    todo,
	})
}

// GetTodos retrieves all todos for the authenticated user with context support
func GetTodos(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// Create context with timeout for database operations
	ctx, cancel := utils.CreateRequestContext(c.Request.Context(), 5*time.Second)
	defer cancel()

	todos, err := models.GetTodosByUserID(ctx, userID.(uint))
	if err != nil {
		// Check if context was cancelled
		if ctx.Err() != nil {
			c.JSON(http.StatusRequestTimeout, gin.H{
				"success": false,
				"message": "Request timeout",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch todos",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Todos retrieved successfully",
		"data":    todos,
		"count":   len(todos),
	})
}

// GetTodo retrieves a single todo by ID with context support
func GetTodo(c *gin.Context) {
	// Parse todo ID from URL
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid todo ID",
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// Create context with timeout for database operations
	ctx, cancel := utils.CreateRequestContext(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Fetch the todo
	todo, err := models.GetTodoByID(ctx, uint(id))
	if err != nil {
		// Check if context was cancelled
		if ctx.Err() != nil {
			c.JSON(http.StatusRequestTimeout, gin.H{
				"success": false,
				"message": "Request timeout",
			})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Todo not found",
		})
		return
	}

	// Check ownership - Authorization
	if todo.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "You don't have permission to access this todo",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Todo retrieved successfully",
		"data":    todo,
	})
}

// UpdateTodo handles updating an existing todo with context support
func UpdateTodo(c *gin.Context) {
	// Parse todo ID from URL
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid todo ID",
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// Create context with timeout for database operations
	ctx, cancel := utils.CreateRequestContext(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Fetch the todo
	todo, err := models.GetTodoByID(ctx, uint(id))
	if err != nil {
		// Check if context was cancelled
		if ctx.Err() != nil {
			c.JSON(http.StatusRequestTimeout, gin.H{
				"success": false,
				"message": "Request timeout",
			})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Todo not found",
		})
		return
	}

	// Check ownership - Authorization
	if todo.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "You don't have permission to update this todo",
		})
		return
	}

	// Parse update input
	var input UpdateTodoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Validation failed",
			"error":   err.Error(),
		})
		return
	}

	// Update fields if provided
	if input.Title != "" {
		todo.Title = input.Title
	}
	if input.Description != "" {
		todo.Description = input.Description
	}
	if input.Completed != nil {
		todo.Completed = *input.Completed
	}

	// Save updates
	if err := models.UpdateTodo(ctx, todo); err != nil {
		// Check if context was cancelled
		if ctx.Err() != nil {
			c.JSON(http.StatusRequestTimeout, gin.H{
				"success": false,
				"message": "Request timeout",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update todo",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Todo updated successfully",
		"data":    todo,
	})
}

// DeleteTodo handles deleting a todo with context support
func DeleteTodo(c *gin.Context) {
	// Parse todo ID from URL
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid todo ID",
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// Create context with timeout for database operations
	ctx, cancel := utils.CreateRequestContext(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Fetch the todo
	todo, err := models.GetTodoByID(ctx, uint(id))
	if err != nil {
		// Check if context was cancelled
		if ctx.Err() != nil {
			c.JSON(http.StatusRequestTimeout, gin.H{
				"success": false,
				"message": "Request timeout",
			})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Todo not found",
		})
		return
	}

	// Check ownership - Authorization
	if todo.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "You don't have permission to delete this todo",
		})
		return
	}

	// Delete the todo
	if err := models.DeleteTodo(ctx, uint(id)); err != nil {
		// Check if context was cancelled
		if ctx.Err() != nil {
			c.JSON(http.StatusRequestTimeout, gin.H{
				"success": false,
				"message": "Request timeout",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to delete todo",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Todo deleted successfully",
	})
}
