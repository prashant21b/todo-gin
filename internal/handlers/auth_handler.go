package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"todo-app/internal/dto"
	"todo-app/internal/services"
	"todo-app/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles HTTP requests for authentication
type AuthHandler struct {
	authService services.AuthService
}

// NewAuthHandler creates a new AuthHandler with the provided service
func NewAuthHandler(svc services.AuthService) *AuthHandler {
	return &AuthHandler{authService: svc}
}

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

// handleAuthError maps service errors to HTTP responses
func (h *AuthHandler) handleAuthError(c *gin.Context, ctx context.Context, err error, operation string, email string) bool {
	if err == nil {
		return false
	}

	// Check for timeout
	if ctx.Err() != nil {
		respondTimeout(c)
		return true
	}

	// Handle specific business errors
	if errors.Is(err, services.ErrEmailAlreadyRegistered) {
		respondConflict(c, err.Error())
		return true
	}

	if errors.Is(err, services.ErrInvalidCredentials) {
		respondUnauthorizedWithMessage(c, err.Error())
		return true
	}

	// Log and return generic error
	rid := utils.GetRequestID(c.Request.Context())
	log.Printf("[%s] request=%s email=%s error=%v", operation, rid, email, err)

	respondInternalError(c, "Failed to "+operation, err)
	return true
}

// Register handles user registration HTTP request
func (h *AuthHandler) Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondBadRequest(c, "Validation failed", err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	response, err := h.authService.RegisterUser(ctx, dto.RegisterRequest{
		Name:     input.Name,
		Email:    input.Email,
		Password: input.Password,
	})

	if h.handleAuthError(c, ctx, err, "register", input.Email) {
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User registered successfully",
		"data": gin.H{
			"user":  response.User,
			"token": response.Token,
		},
	})
}

// Login handles user authentication HTTP request
func (h *AuthHandler) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondBadRequest(c, "Validation failed", err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	response, err := h.authService.LoginUser(ctx, dto.LoginRequest{
		Email:    input.Email,
		Password: input.Password,
	})

	if h.handleAuthError(c, ctx, err, "login", input.Email) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
		"data": gin.H{
			"user":  response.User,
			"token": response.Token,
		},
	})
}
