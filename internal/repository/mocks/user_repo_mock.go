package mocks

import (
	"context"

	"todo-app/internal/models"
	"todo-app/internal/repository"
)

// Ensure MockUserRepository implements UserRepository
var _ repository.UserRepository = (*MockUserRepository)(nil)

// MockUserRepository is a mock implementation of UserRepository for testing
type MockUserRepository struct {
	CreateUserFunc     func(ctx context.Context, user *models.User) error
	GetUserByEmailFunc func(ctx context.Context, email string) (*models.User, error)
	GetUserByIDFunc    func(ctx context.Context, id uint) (*models.User, error)
}

// CreateUser calls the mock function
func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, user)
	}
	return nil
}

// GetUserByEmail calls the mock function
func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	return nil, nil
}

// GetUserByID calls the mock function
func (m *MockUserRepository) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, id)
	}
	return nil, nil
}
