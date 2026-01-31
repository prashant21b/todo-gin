package controllers

import (
	"net/http"
	"time"

	"todo-app/models"
	"todo-app/utils"

	"github.com/gin-gonic/gin"
)

// RegisterInput represents the registration request body
type RegisterInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginInput represents the login request body
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register handles user registration with context and timeout
func Register(c *gin.Context) {
	var input RegisterInput

	// Validate input
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Validation failed",
			"error":   err.Error(),
		})
		return
	}

	// Create context with timeout for database operations
	ctx, cancel := utils.CreateRequestContext(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Check if user already exists
	existingUser, _ := models.GetUserByEmail(ctx, input.Email)
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"message": "Email already registered",
		})
		return
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to hash password",
		})
		return
	}

	// Create user
	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: hashedPassword,
	}

	if err := models.CreateUser(ctx, &user); err != nil {
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
			"message": "Failed to create user",
			"error":   err.Error(),
		})
		return
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User registered successfully",
		"data": gin.H{
			"user":  user,
			"token": token,
		},
	})
}

// Login handles user authentication with context and timeout
func Login(c *gin.Context) {
	var input LoginInput

	// Validate input
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Validation failed",
			"error":   err.Error(),
		})
		return
	}

	// Create context with timeout for database operations
	ctx, cancel := utils.CreateRequestContext(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Find user by email
	user, err := models.GetUserByEmail(ctx, input.Email)
	if err != nil {
		// Check if context was cancelled
		if ctx.Err() != nil {
			c.JSON(http.StatusRequestTimeout, gin.H{
				"success": false,
				"message": "Request timeout",
			})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Invalid email or password",
		})
		return
	}

	// Verify password
	if !utils.CheckPassword(input.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Invalid email or password",
		})
		return
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
		"data": gin.H{
			"user":  user,
			"token": token,
		},
	})
}
