package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"todo-app/internal/dto"
	"todo-app/internal/models"
	"todo-app/internal/repository"
)

// Common errors for category operations
var (
	ErrCategoryNotFound    = errors.New("category not found")
	ErrCategoryForbidden   = errors.New("you don't have permission to access this category")
	ErrCategoryNameExists  = errors.New("category with this name already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrCannotShareWithSelf = errors.New("cannot share category with yourself")
	ErrShareAlreadyExists  = errors.New("category is already shared with this user")
	ErrShareNotFound       = errors.New("share not found")
)

// Ensure CategoryServiceImpl implements CategoryService
var _ CategoryService = (*CategoryServiceImpl)(nil)

// CategoryServiceImpl provides business logic for categories
type CategoryServiceImpl struct {
	categoryRepo      repository.CategoryRepository
	categoryShareRepo repository.CategoryShareRepository
	userRepo          repository.UserRepository
	todoRepo          repository.TodoRepository
}

// NewCategoryService creates a new CategoryService with the provided repositories
func NewCategoryService(
	categoryRepo repository.CategoryRepository,
	categoryShareRepo repository.CategoryShareRepository,
	userRepo repository.UserRepository,
	todoRepo repository.TodoRepository,
) CategoryService {
	return &CategoryServiceImpl{
		categoryRepo:      categoryRepo,
		categoryShareRepo: categoryShareRepo,
		userRepo:          userRepo,
		todoRepo:          todoRepo,
	}
}

// CreateCategory creates a new category for a user
func (s *CategoryServiceImpl) CreateCategory(ctx context.Context, req dto.CreateCategoryRequest) (*models.Category, error) {
	// Check if category with same name exists for this user
	existing, err := s.categoryRepo.GetCategoryByNameAndOwner(ctx, req.OwnerID, req.Name)
	if err == nil && existing != nil {
		return nil, ErrCategoryNameExists
	}
	// Only treat sql.ErrNoRows as "not found", other errors should be returned
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to check existing category: %w", err)
	}

	category := &models.Category{
		Name:    req.Name,
		OwnerID: req.OwnerID,
	}

	if err := s.categoryRepo.CreateCategory(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return category, nil
}

// GetCategories retrieves all categories owned by a user
func (s *CategoryServiceImpl) GetCategories(ctx context.Context, userID uint) ([]models.Category, error) {
	categories, err := s.categoryRepo.GetCategoriesByOwnerID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}

	// For each category, fetch todos belonging to that category (owner-created todos)
	for i := range categories {
		todos, _, err := s.todoRepo.GetTodosByCategoryID(ctx, categories[i].ID, 1, 1000)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch todos for category %d: %w", categories[i].ID, err)
		}
		categories[i].Todos = todos
	}

	return categories, nil
}

// GetCategoryByID retrieves a category by ID with ownership verification
func (s *CategoryServiceImpl) GetCategoryByID(ctx context.Context, categoryID, userID uint) (*models.Category, error) {
	category, err := s.categoryRepo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to fetch category: %w", err)
	}

	// Check ownership - for now only owner can access their categories
	// In Phase 3+, we'll also allow users with shared access
	if category.OwnerID != userID {
		// Check if user has shared access
		permission, _ := s.categoryShareRepo.GetUserPermissionForCategory(ctx, userID, categoryID)
		if permission == "none" || permission == "" {
			return nil, ErrCategoryForbidden
		}
	}

	return category, nil
}

// UpdateCategory updates a category with ownership verification
func (s *CategoryServiceImpl) UpdateCategory(ctx context.Context, req dto.UpdateCategoryRequest) (*models.Category, error) {
	// Fetch existing category
	category, err := s.categoryRepo.GetCategoryByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to fetch category: %w", err)
	}

	// Verify ownership - only owner can update
	if category.OwnerID != req.UserID {
		return nil, ErrCategoryForbidden
	}

	// Check if new name conflicts with existing category
	if req.Name != category.Name {
		existing, err := s.categoryRepo.GetCategoryByNameAndOwner(ctx, req.UserID, req.Name)
		if err == nil && existing != nil {
			return nil, ErrCategoryNameExists
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to check existing category: %w", err)
		}
	}

	// Update the category
	category.Name = req.Name
	if err := s.categoryRepo.UpdateCategory(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return category, nil
}

// DeleteCategory deletes a category with ownership verification
func (s *CategoryServiceImpl) DeleteCategory(ctx context.Context, categoryID, userID uint) error {
	// Fetch existing category
	category, err := s.categoryRepo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCategoryNotFound
		}
		return fmt.Errorf("failed to fetch category: %w", err)
	}

	// Verify ownership - only owner can delete
	if category.OwnerID != userID {
		return ErrCategoryForbidden
	}

	// Delete the category (cascades to shares and todos via FK)
	if err := s.categoryRepo.DeleteCategory(ctx, categoryID); err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}

// ShareCategory shares a category with another user
func (s *CategoryServiceImpl) ShareCategory(ctx context.Context, req dto.ShareCategoryRequest) (*models.CategoryShare, error) {
	// Verify category exists and user is owner
	category, err := s.categoryRepo.GetCategoryByID(ctx, req.CategoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to fetch category: %w", err)
	}

	if category.OwnerID != req.OwnerID {
		return nil, ErrCategoryForbidden
	}

	// Find user to share with by email
	shareWithUser, err := s.userRepo.GetUserByEmail(ctx, req.ShareWithEmail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Cannot share with yourself
	if shareWithUser.ID == req.OwnerID {
		return nil, ErrCannotShareWithSelf
	}

	// Check if share already exists
	existing, err := s.categoryShareRepo.GetCategoryShareByCategoryAndUser(ctx, req.CategoryID, shareWithUser.ID)
	if err == nil && existing != nil {
		return nil, ErrShareAlreadyExists
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to check existing share: %w", err)
	}

	// Create the share
	share := &models.CategoryShare{
		CategoryID:       req.CategoryID,
		SharedWithUserID: shareWithUser.ID,
		Permission:       req.Permission,
	}

	if err := s.categoryShareRepo.CreateCategoryShare(ctx, share); err != nil {
		return nil, fmt.Errorf("failed to create share: %w", err)
	}

	return share, nil
}

// UnshareCategory removes sharing of a category with a user
func (s *CategoryServiceImpl) UnshareCategory(ctx context.Context, req dto.UnshareCategoryRequest) error {
	// Verify category exists and user is owner
	category, err := s.categoryRepo.GetCategoryByID(ctx, req.CategoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCategoryNotFound
		}
		return fmt.Errorf("failed to fetch category: %w", err)
	}

	if category.OwnerID != req.OwnerID {
		return ErrCategoryForbidden
	}

	// Verify share exists
	_, err = s.categoryShareRepo.GetCategoryShareByCategoryAndUser(ctx, req.CategoryID, req.SharedWithUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrShareNotFound
		}
		return fmt.Errorf("failed to fetch share: %w", err)
	}

	// Delete the share
	if err := s.categoryShareRepo.DeleteCategoryShareByUserAndCategory(ctx, req.CategoryID, req.SharedWithUserID); err != nil {
		return fmt.Errorf("failed to delete share: %w", err)
	}

	return nil
}

// UpdateSharePermission changes the permission of a shared category
func (s *CategoryServiceImpl) UpdateSharePermission(ctx context.Context, req dto.UpdateSharePermissionRequest) error {
	// Verify category exists and user is owner
	category, err := s.categoryRepo.GetCategoryByID(ctx, req.CategoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCategoryNotFound
		}
		return fmt.Errorf("failed to fetch category: %w", err)
	}

	if category.OwnerID != req.OwnerID {
		return ErrCategoryForbidden
	}

	// Verify share exists
	share, err := s.categoryShareRepo.GetCategoryShareByCategoryAndUser(ctx, req.CategoryID, req.SharedWithUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrShareNotFound
		}
		return fmt.Errorf("failed to fetch share: %w", err)
	}

	// Update the permission
	if err := s.categoryShareRepo.UpdateCategorySharePermission(ctx, share.ID, req.Permission); err != nil {
		return fmt.Errorf("failed to update share permission: %w", err)
	}

	return nil
}

// GetSharesForCategory gets all shares for a category (owner only)
func (s *CategoryServiceImpl) GetSharesForCategory(ctx context.Context, categoryID, userID uint) ([]models.CategoryShareWithUser, error) {
	// Verify category exists and user is owner
	category, err := s.categoryRepo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to fetch category: %w", err)
	}

	if category.OwnerID != userID {
		return nil, ErrCategoryForbidden
	}

	shares, err := s.categoryShareRepo.GetSharesForCategory(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shares: %w", err)
	}

	return shares, nil
}

// GetSharedCategories gets all categories shared with a user
func (s *CategoryServiceImpl) GetSharedCategories(ctx context.Context, userID uint) ([]models.SharedCategoryWithOwner, error) {
	categories, err := s.categoryShareRepo.GetSharedCategoriesForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shared categories: %w", err)
	}

	// Populate todos for each shared category
	for i := range categories {
		todos, _, err := s.todoRepo.GetTodosByCategoryID(ctx, categories[i].ID, 1, 1000)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch todos for shared category %d: %w", categories[i].ID, err)
		}
		categories[i].Todos = todos
	}

	return categories, nil
}

// GetUserPermissionForCategory checks what permission a user has for a category
func (s *CategoryServiceImpl) GetUserPermissionForCategory(ctx context.Context, userID, categoryID uint) (string, error) {
	permission, err := s.categoryShareRepo.GetUserPermissionForCategory(ctx, userID, categoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "none", nil
		}
		return "", fmt.Errorf("failed to fetch permission: %w", err)
	}
	return permission, nil
}
