package dto

import "todo-app/internal/models"

// RegisterRequest represents user registration data
type RegisterRequest struct {
	Name     string
	Email    string
	Password string
}

// LoginRequest represents user login credentials
type LoginRequest struct {
	Email    string
	Password string
}

// AuthResponse represents the authentication response with user and token
type AuthResponse struct {
	User  *models.User
	Token string
}
