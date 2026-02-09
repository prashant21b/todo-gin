package mocks

import (
	"context"

	"todo-app/internal/models"
	"todo-app/internal/repository"
)

// Ensure MockCategoryShareRepository implements CategoryShareRepository
var _ repository.CategoryShareRepository = (*MockCategoryShareRepository)(nil)

// MockCategoryShareRepository is a mock implementation of CategoryShareRepository for testing
type MockCategoryShareRepository struct {
	CreateCategoryShareFunc                  func(ctx context.Context, share *models.CategoryShare) error
	GetCategoryShareByIDFunc                 func(ctx context.Context, id uint) (*models.CategoryShare, error)
	GetCategoryShareByCategoryAndUserFunc    func(ctx context.Context, categoryID, userID uint) (*models.CategoryShare, error)
	GetSharesForCategoryFunc                 func(ctx context.Context, categoryID uint) ([]models.CategoryShareWithUser, error)
	GetSharedCategoriesForUserFunc           func(ctx context.Context, userID uint) ([]models.SharedCategoryWithOwner, error)
	UpdateCategorySharePermissionFunc        func(ctx context.Context, id uint, permission models.Permission) error
	DeleteCategoryShareFunc                  func(ctx context.Context, id uint) error
	DeleteCategoryShareByUserAndCategoryFunc func(ctx context.Context, categoryID, userID uint) error
	GetUserPermissionForCategoryFunc         func(ctx context.Context, userID, categoryID uint) (string, error)
	GetTodosGroupedByCategoryFunc            func(ctx context.Context, userID uint) ([]models.CategoryWithTodosRow, error)
}

// CreateCategoryShare calls the mock function
func (m *MockCategoryShareRepository) CreateCategoryShare(ctx context.Context, share *models.CategoryShare) error {
	if m.CreateCategoryShareFunc != nil {
		return m.CreateCategoryShareFunc(ctx, share)
	}
	return nil
}

// GetCategoryShareByID calls the mock function
func (m *MockCategoryShareRepository) GetCategoryShareByID(ctx context.Context, id uint) (*models.CategoryShare, error) {
	if m.GetCategoryShareByIDFunc != nil {
		return m.GetCategoryShareByIDFunc(ctx, id)
	}
	return nil, nil
}

// GetCategoryShareByCategoryAndUser calls the mock function
func (m *MockCategoryShareRepository) GetCategoryShareByCategoryAndUser(ctx context.Context, categoryID, userID uint) (*models.CategoryShare, error) {
	if m.GetCategoryShareByCategoryAndUserFunc != nil {
		return m.GetCategoryShareByCategoryAndUserFunc(ctx, categoryID, userID)
	}
	return nil, nil
}

// GetSharesForCategory calls the mock function
func (m *MockCategoryShareRepository) GetSharesForCategory(ctx context.Context, categoryID uint) ([]models.CategoryShareWithUser, error) {
	if m.GetSharesForCategoryFunc != nil {
		return m.GetSharesForCategoryFunc(ctx, categoryID)
	}
	return []models.CategoryShareWithUser{}, nil
}

// GetSharedCategoriesForUser calls the mock function
func (m *MockCategoryShareRepository) GetSharedCategoriesForUser(ctx context.Context, userID uint) ([]models.SharedCategoryWithOwner, error) {
	if m.GetSharedCategoriesForUserFunc != nil {
		return m.GetSharedCategoriesForUserFunc(ctx, userID)
	}
	return []models.SharedCategoryWithOwner{}, nil
}

// UpdateCategorySharePermission calls the mock function
func (m *MockCategoryShareRepository) UpdateCategorySharePermission(ctx context.Context, id uint, permission models.Permission) error {
	if m.UpdateCategorySharePermissionFunc != nil {
		return m.UpdateCategorySharePermissionFunc(ctx, id, permission)
	}
	return nil
}

// DeleteCategoryShare calls the mock function
func (m *MockCategoryShareRepository) DeleteCategoryShare(ctx context.Context, id uint) error {
	if m.DeleteCategoryShareFunc != nil {
		return m.DeleteCategoryShareFunc(ctx, id)
	}
	return nil
}

// DeleteCategoryShareByUserAndCategory calls the mock function
func (m *MockCategoryShareRepository) DeleteCategoryShareByUserAndCategory(ctx context.Context, categoryID, userID uint) error {
	if m.DeleteCategoryShareByUserAndCategoryFunc != nil {
		return m.DeleteCategoryShareByUserAndCategoryFunc(ctx, categoryID, userID)
	}
	return nil
}

// GetUserPermissionForCategory calls the mock function
func (m *MockCategoryShareRepository) GetUserPermissionForCategory(ctx context.Context, userID, categoryID uint) (string, error) {
	if m.GetUserPermissionForCategoryFunc != nil {
		return m.GetUserPermissionForCategoryFunc(ctx, userID, categoryID)
	}
	return "none", nil
}

// GetTodosGroupedByCategory calls the mock function
func (m *MockCategoryShareRepository) GetTodosGroupedByCategory(ctx context.Context, userID uint) ([]models.CategoryWithTodosRow, error) {
	if m.GetTodosGroupedByCategoryFunc != nil {
		return m.GetTodosGroupedByCategoryFunc(ctx, userID)
	}
	return []models.CategoryWithTodosRow{}, nil
}
