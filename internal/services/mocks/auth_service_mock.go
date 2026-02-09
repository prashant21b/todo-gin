package mocks

import (
	"context"

	"todo-app/internal/dto"
	"todo-app/internal/models"
	"todo-app/internal/services"
)

// Ensure MockAuthService implements AuthService
var _ services.AuthService = (*MockAuthService)(nil)

// MockAuthService is a mock implementation of AuthService for testing
type MockAuthService struct {
	RegisterUserFunc func(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error)
	LoginUserFunc    func(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
	GetByIDFunc      func(ctx context.Context, id uint) (*models.User, error)
}

// RegisterUser calls the mock function
func (m *MockAuthService) RegisterUser(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	if m.RegisterUserFunc != nil {
		return m.RegisterUserFunc(ctx, req)
	}
	return nil, nil
}

// LoginUser calls the mock function
func (m *MockAuthService) LoginUser(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	if m.LoginUserFunc != nil {
		return m.LoginUserFunc(ctx, req)
	}
	return nil, nil
}

// GetByID calls the mock function
func (m *MockAuthService) GetByID(ctx context.Context, id uint) (*models.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}
