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

// Common errors for todo operations
var (
	ErrTodoNotFound      = errors.New("todo not found")
	ErrForbidden         = errors.New("you don't have permission to access this todo")
	ErrInvalidTodoID     = errors.New("invalid todo id")
	ErrCategoryRequired  = errors.New("category is required")
	ErrNoWritePermission = errors.New("you don't have write permission for this category")
)

// PaginationConfig holds pagination settings
type PaginationConfig struct {
	DefaultPageSize int
	MaxPageSize     int
}

// Ensure TodoServiceImpl implements TodoService
var _ TodoService = (*TodoServiceImpl)(nil)

// TodoServiceImpl provides business logic for todos
type TodoServiceImpl struct {
	repo              repository.TodoRepository
	categoryRepo      repository.CategoryRepository
	categoryShareRepo repository.CategoryShareRepository
	pagination        PaginationConfig
}

// NewTodoService creates a new TodoService with the provided repositories and pagination config
func NewTodoService(
	repo repository.TodoRepository,
	categoryRepo repository.CategoryRepository,
	categoryShareRepo repository.CategoryShareRepository,
	pagination PaginationConfig,
) TodoService {
	return &TodoServiceImpl{
		repo:              repo,
		categoryRepo:      categoryRepo,
		categoryShareRepo: categoryShareRepo,
		pagination:        pagination,
	}
}

// checkCategoryPermission checks if user has at least the required permission for a category
// Returns the permission level ("owner", "write", "read", "none") and any error
func (s *TodoServiceImpl) checkCategoryPermission(ctx context.Context, userID, categoryID uint, requireWrite bool) error {
	// First check if category exists
	category, err := s.categoryRepo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCategoryNotFound
		}
		return fmt.Errorf("failed to fetch category: %w", err)
	}

	// If user is owner, they have full access
	if category.OwnerID == userID {
		return nil
	}

	// Check shared permission
	permission, err := s.categoryShareRepo.GetUserPermissionForCategory(ctx, userID, categoryID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to check permission: %w", err)
	}

	// Check if user has any access
	if permission == "none" || permission == "" {
		return ErrForbidden
	}

	// If write is required, check for write permission
	if requireWrite && permission != "write" {
		return ErrNoWritePermission
	}

	return nil
}

// getOrCreateCategory finds an existing category by name for the user, or creates a new one
func (s *TodoServiceImpl) getOrCreateCategory(ctx context.Context, userID uint, categoryName string) (*models.Category, error) {
	// Try to find existing category by name
	category, err := s.categoryRepo.GetCategoryByNameAndOwner(ctx, userID, categoryName)
	if err == nil {
		// Category exists, return it
		return category, nil
	}

	// Category doesn't exist, create it
	newCategory := &models.Category{
		Name:    categoryName,
		OwnerID: userID,
	}

	if err := s.categoryRepo.CreateCategory(ctx, newCategory); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return newCategory, nil
}

// CreateTodo handles todo creation workflow
func (s *TodoServiceImpl) CreateTodo(ctx context.Context, req dto.CreateTodoRequest) (*models.Todo, error) {
	var category *models.Category

	if req.CategoryID != nil && *req.CategoryID > 0 {
		// Use existing category by ID: require write permission (owner or shared with write)
		if err := s.checkCategoryPermission(ctx, req.UserID, *req.CategoryID, true); err != nil {
			return nil, err
		}
		var err error
		category, err = s.categoryRepo.GetCategoryByID(ctx, *req.CategoryID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrCategoryNotFound
			}
			return nil, fmt.Errorf("failed to fetch category: %w", err)
		}
	} else {
		// Use category name: get-or-create for the user (owner only)
		if req.Category == "" {
			return nil, ErrCategoryRequired
		}
		var err error
		category, err = s.getOrCreateCategory(ctx, req.UserID, req.Category)
		if err != nil {
			return nil, err
		}
	}

	todo := &models.Todo{
		Title:       req.Title,
		Description: req.Description,
		CategoryID:  category.ID,
		UserID:      req.UserID,
		CreatedBy:   req.UserID,
	}

	if err := s.repo.CreateTodo(ctx, todo); err != nil {
		return nil, fmt.Errorf("failed to create todo: %w", err)
	}

	return todo, nil
}

// GetTodos retrieves todos for a user with pagination
func (s *TodoServiceImpl) GetTodos(ctx context.Context, userID uint, page, pageSize int) (*dto.TodoListResponse, error) {
	// Normalize pagination parameters using config values
	page = max(page, 1)
	if pageSize < 1 {
		pageSize = s.pagination.DefaultPageSize
	}
	pageSize = min(pageSize, s.pagination.MaxPageSize)

	todos, total, err := s.repo.GetTodos(ctx, userID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch todos: %w", err)
	}

	// Calculate total pages
	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	return &dto.TodoListResponse{
		Todos:      todos,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetTodosByCategoryID retrieves todos filtered by category ID with pagination
func (s *TodoServiceImpl) GetTodosByCategoryID(ctx context.Context, categoryID uint, page, pageSize int) (*dto.TodoListResponse, error) {
	// Normalize pagination parameters using config values
	page = max(page, 1)
	if pageSize < 1 {
		pageSize = s.pagination.DefaultPageSize
	}
	pageSize = min(pageSize, s.pagination.MaxPageSize)

	todos, total, err := s.repo.GetTodosByCategoryID(ctx, categoryID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch todos by category: %w", err)
	}

	// Calculate total pages
	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	return &dto.TodoListResponse{
		Todos:      todos,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetTodoByID retrieves a single todo with ownership/permission verification
func (s *TodoServiceImpl) GetTodoByID(ctx context.Context, req dto.GetTodoRequest) (*models.Todo, error) {
	todo, err := s.repo.GetTodoByID(ctx, req.ID)
	if err != nil {
		return nil, ErrTodoNotFound
	}

	// Check if user has at least read permission for the todo's category
	if err := s.checkCategoryPermission(ctx, req.UserID, todo.CategoryID, false); err != nil {
		return nil, err
	}

	return todo, nil
}

// UpdateTodo handles todo update with ownership/permission verification
func (s *TodoServiceImpl) UpdateTodo(ctx context.Context, req dto.UpdateTodoRequest) (*models.Todo, error) {
	// Fetch existing todo
	todo, err := s.repo.GetTodoByID(ctx, req.ID)
	if err != nil {
		return nil, ErrTodoNotFound
	}

	// Check if user has write permission for the current category
	if err := s.checkCategoryPermission(ctx, req.UserID, todo.CategoryID, true); err != nil {
		return nil, err
	}

	// If changing category, check write permission for the new category
	if req.CategoryID != nil && *req.CategoryID != todo.CategoryID {
		if err := s.checkCategoryPermission(ctx, req.UserID, *req.CategoryID, true); err != nil {
			return nil, err
		}
		// Get new category to update UserID (todo belongs to category owner)
		newCategory, err := s.categoryRepo.GetCategoryByID(ctx, *req.CategoryID)
		if err != nil {
			return nil, ErrCategoryNotFound
		}
		todo.CategoryID = *req.CategoryID
		todo.UserID = newCategory.OwnerID
	}

	// Apply updates (only update fields that are provided)
	if req.Title != nil && *req.Title != "" {
		todo.Title = *req.Title
	}
	if req.Description != nil {
		todo.Description = *req.Description
	}
	if req.Completed != nil {
		todo.Completed = *req.Completed
	}

	// Save updates
	if err := s.repo.UpdateTodo(ctx, todo); err != nil {
		return nil, fmt.Errorf("failed to update todo: %w", err)
	}

	return todo, nil
}

// DeleteTodo handles todo soft deletion with ownership/permission verification
func (s *TodoServiceImpl) DeleteTodo(ctx context.Context, req dto.DeleteTodoRequest) error {
	// Fetch existing todo
	todo, err := s.repo.GetTodoByID(ctx, req.ID)
	if err != nil {
		return ErrTodoNotFound
	}

	// Check if user has write permission for the category
	if err := s.checkCategoryPermission(ctx, req.UserID, todo.CategoryID, true); err != nil {
		return err
	}

	// Soft delete the todo
	if err := s.repo.DeleteTodo(ctx, req.ID); err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	return nil
}

// GetTodosGroupedByCategory retrieves all accessible todos grouped by category
func (s *TodoServiceImpl) GetTodosGroupedByCategory(ctx context.Context, userID uint) (*dto.TodosGroupedByCategoryResponse, error) {
	// Get flat rows from repository
	rows, err := s.categoryShareRepo.GetTodosGroupedByCategory(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch todos grouped by category: %w", err)
	}

	// Group the flat rows by category
	categoryMap := make(map[uint]*dto.CategoryWithTodos)
	categoryOrder := make([]uint, 0)

	for _, row := range rows {
		// Check if we've already seen this category
		cat, exists := categoryMap[row.CategoryID]
		if !exists {
			// Create new category entry
			cat = &dto.CategoryWithTodos{
				ID:             row.CategoryID,
				Name:           row.CategoryName,
				OwnerID:        row.CategoryOwnerID,
				OwnerName:      row.CategoryOwnerName,
				UserPermission: row.UserPermission,
				Todos:          []dto.TodoInCategory{},
			}
			categoryMap[row.CategoryID] = cat
			categoryOrder = append(categoryOrder, row.CategoryID)
		}

		// Add todo to category (only if there is a todo - todo_id > 0)
		if row.TodoID > 0 {
			todoItem := dto.TodoInCategory{
				ID:          row.TodoID,
				Title:       row.TodoTitle,
				Description: row.TodoDescription,
				Completed:   row.TodoCompleted,
				CreatedBy:   row.TodoCreatedBy,
				CreatorName: row.TodoCreatorName,
			}
			if row.TodoCreatedAt != nil {
				todoItem.CreatedAt = *row.TodoCreatedAt
			}
			if row.TodoUpdatedAt != nil {
				todoItem.UpdatedAt = *row.TodoUpdatedAt
			}
			cat.Todos = append(cat.Todos, todoItem)
		}
	}

	// Build response maintaining category order
	categories := make([]dto.CategoryWithTodos, 0, len(categoryOrder))
	for _, catID := range categoryOrder {
		categories = append(categories, *categoryMap[catID])
	}

	return &dto.TodosGroupedByCategoryResponse{
		Categories: categories,
	}, nil
}
