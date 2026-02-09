package mocks

import (
	"context"

	"todo-app/internal/dto"
	"todo-app/internal/models"
	"todo-app/internal/services"
)

// Ensure MockCategoryService implements CategoryService
var _ services.CategoryService = (*MockCategoryService)(nil)

// MockCategoryService is a mock implementation of CategoryService for testing
type MockCategoryService struct {
	CreateCategoryFunc               func(ctx context.Context, req dto.CreateCategoryRequest) (*models.Category, error)
	GetCategoriesFunc                func(ctx context.Context, userID uint) ([]models.Category, error)
	GetCategoryByIDFunc              func(ctx context.Context, categoryID, userID uint) (*models.Category, error)
	UpdateCategoryFunc               func(ctx context.Context, req dto.UpdateCategoryRequest) (*models.Category, error)
	DeleteCategoryFunc               func(ctx context.Context, categoryID, userID uint) error
	ShareCategoryFunc                func(ctx context.Context, req dto.ShareCategoryRequest) (*models.CategoryShare, error)
	UnshareCategoryFunc              func(ctx context.Context, req dto.UnshareCategoryRequest) error
	UpdateSharePermissionFunc        func(ctx context.Context, req dto.UpdateSharePermissionRequest) error
	GetSharesForCategoryFunc         func(ctx context.Context, categoryID, userID uint) ([]models.CategoryShareWithUser, error)
	GetSharedCategoriesFunc          func(ctx context.Context, userID uint) ([]models.SharedCategoryWithOwner, error)
	GetUserPermissionForCategoryFunc func(ctx context.Context, userID, categoryID uint) (string, error)
}

// CreateCategory calls the mock function
func (m *MockCategoryService) CreateCategory(ctx context.Context, req dto.CreateCategoryRequest) (*models.Category, error) {
	if m.CreateCategoryFunc != nil {
		return m.CreateCategoryFunc(ctx, req)
	}
	return &models.Category{}, nil
}

// GetCategories calls the mock function
func (m *MockCategoryService) GetCategories(ctx context.Context, userID uint) ([]models.Category, error) {
	if m.GetCategoriesFunc != nil {
		return m.GetCategoriesFunc(ctx, userID)
	}
	return []models.Category{}, nil
}

// GetCategoryByID calls the mock function
func (m *MockCategoryService) GetCategoryByID(ctx context.Context, categoryID, userID uint) (*models.Category, error) {
	if m.GetCategoryByIDFunc != nil {
		return m.GetCategoryByIDFunc(ctx, categoryID, userID)
	}
	return nil, nil
}

// UpdateCategory calls the mock function
func (m *MockCategoryService) UpdateCategory(ctx context.Context, req dto.UpdateCategoryRequest) (*models.Category, error) {
	if m.UpdateCategoryFunc != nil {
		return m.UpdateCategoryFunc(ctx, req)
	}
	return &models.Category{}, nil
}

// DeleteCategory calls the mock function
func (m *MockCategoryService) DeleteCategory(ctx context.Context, categoryID, userID uint) error {
	if m.DeleteCategoryFunc != nil {
		return m.DeleteCategoryFunc(ctx, categoryID, userID)
	}
	return nil
}

// ShareCategory calls the mock function
func (m *MockCategoryService) ShareCategory(ctx context.Context, req dto.ShareCategoryRequest) (*models.CategoryShare, error) {
	if m.ShareCategoryFunc != nil {
		return m.ShareCategoryFunc(ctx, req)
	}
	return &models.CategoryShare{}, nil
}

// UnshareCategory calls the mock function
func (m *MockCategoryService) UnshareCategory(ctx context.Context, req dto.UnshareCategoryRequest) error {
	if m.UnshareCategoryFunc != nil {
		return m.UnshareCategoryFunc(ctx, req)
	}
	return nil
}

// UpdateSharePermission calls the mock function
func (m *MockCategoryService) UpdateSharePermission(ctx context.Context, req dto.UpdateSharePermissionRequest) error {
	if m.UpdateSharePermissionFunc != nil {
		return m.UpdateSharePermissionFunc(ctx, req)
	}
	return nil
}

// GetSharesForCategory calls the mock function
func (m *MockCategoryService) GetSharesForCategory(ctx context.Context, categoryID, userID uint) ([]models.CategoryShareWithUser, error) {
	if m.GetSharesForCategoryFunc != nil {
		return m.GetSharesForCategoryFunc(ctx, categoryID, userID)
	}
	return []models.CategoryShareWithUser{}, nil
}

// GetSharedCategories calls the mock function
func (m *MockCategoryService) GetSharedCategories(ctx context.Context, userID uint) ([]models.SharedCategoryWithOwner, error) {
	if m.GetSharedCategoriesFunc != nil {
		return m.GetSharedCategoriesFunc(ctx, userID)
	}
	return []models.SharedCategoryWithOwner{}, nil
}

// GetUserPermissionForCategory calls the mock function
func (m *MockCategoryService) GetUserPermissionForCategory(ctx context.Context, userID, categoryID uint) (string, error) {
	if m.GetUserPermissionForCategoryFunc != nil {
		return m.GetUserPermissionForCategoryFunc(ctx, userID, categoryID)
	}
	return "none", nil
}
