package services

import (
	"context"
	"errors"
	"fmt"

	"todo-app/internal/dto"
	"todo-app/internal/models"
	"todo-app/internal/repository"
	"todo-app/pkg/utils"
)

// Common errors for auth operations
var (
	ErrEmailAlreadyRegistered = errors.New("email already registered")
	ErrInvalidCredentials     = errors.New("invalid email or password")
)

// Ensure AuthServiceImpl implements AuthService
var _ AuthService = (*AuthServiceImpl)(nil)

// AuthServiceImpl handles auth business logic
type AuthServiceImpl struct {
	repo       repository.UserRepository
	jwtManager *utils.JWTManager
}

// NewAuthService creates a new AuthService with the provided repository and JWT manager
func NewAuthService(repo repository.UserRepository, jwtManager *utils.JWTManager) AuthService {
	return &AuthServiceImpl{
		repo:       repo,
		jwtManager: jwtManager,
	}
}

// RegisterUser handles complete user registration workflow
func (s *AuthServiceImpl) RegisterUser(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	// Check if user already exists
	existingUser, _ := s.repo.GetUserByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, ErrEmailAlreadyRegistered
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &dto.AuthResponse{
		User:  user,
		Token: token,
	}, nil
}

// LoginUser handles user authentication workflow
func (s *AuthServiceImpl) LoginUser(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	// Find user by email
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &dto.AuthResponse{
		User:  user,
		Token: token,
	}, nil
}

// GetByID retrieves a user by ID
func (s *AuthServiceImpl) GetByID(ctx context.Context, id uint) (*models.User, error) {
	return s.repo.GetUserByID(ctx, id)
}
