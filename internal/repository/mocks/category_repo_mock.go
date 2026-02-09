package mocks

import (
	"context"

	"todo-app/internal/models"
	"todo-app/internal/repository"
)

// Ensure MockCategoryRepository implements CategoryRepository
var _ repository.CategoryRepository = (*MockCategoryRepository)(nil)

// MockCategoryRepository is a mock implementation of CategoryRepository for testing
type MockCategoryRepository struct {
	CreateCategoryFunc          func(ctx context.Context, category *models.Category) error
	GetCategoryByIDFunc         func(ctx context.Context, id uint) (*models.Category, error)
	GetCategoriesByOwnerIDFunc  func(ctx context.Context, ownerID uint) ([]models.Category, error)
	GetCategoryByNameAndOwnerFunc func(ctx context.Context, ownerID uint, name string) (*models.Category, error)
	UpdateCategoryFunc          func(ctx context.Context, category *models.Category) error
	DeleteCategoryFunc          func(ctx context.Context, id uint) error
}

// CreateCategory calls the mock function
func (m *MockCategoryRepository) CreateCategory(ctx context.Context, category *models.Category) error {
	if m.CreateCategoryFunc != nil {
		return m.CreateCategoryFunc(ctx, category)
	}
	return nil
}

// GetCategoryByID calls the mock function
func (m *MockCategoryRepository) GetCategoryByID(ctx context.Context, id uint) (*models.Category, error) {
	if m.GetCategoryByIDFunc != nil {
		return m.GetCategoryByIDFunc(ctx, id)
	}
	return nil, nil
}

// GetCategoriesByOwnerID calls the mock function
func (m *MockCategoryRepository) GetCategoriesByOwnerID(ctx context.Context, ownerID uint) ([]models.Category, error) {
	if m.GetCategoriesByOwnerIDFunc != nil {
		return m.GetCategoriesByOwnerIDFunc(ctx, ownerID)
	}
	return []models.Category{}, nil
}

// GetCategoryByNameAndOwner calls the mock function
func (m *MockCategoryRepository) GetCategoryByNameAndOwner(ctx context.Context, ownerID uint, name string) (*models.Category, error) {
	if m.GetCategoryByNameAndOwnerFunc != nil {
		return m.GetCategoryByNameAndOwnerFunc(ctx, ownerID, name)
	}
	return nil, nil
}

// UpdateCategory calls the mock function
func (m *MockCategoryRepository) UpdateCategory(ctx context.Context, category *models.Category) error {
	if m.UpdateCategoryFunc != nil {
		return m.UpdateCategoryFunc(ctx, category)
	}
	return nil
}

// DeleteCategory calls the mock function
func (m *MockCategoryRepository) DeleteCategory(ctx context.Context, id uint) error {
	if m.DeleteCategoryFunc != nil {
		return m.DeleteCategoryFunc(ctx, id)
	}
	return nil
}
