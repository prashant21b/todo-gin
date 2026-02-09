package services

import (
	"context"

	"todo-app/internal/dto"
	"todo-app/internal/models"
)

// TodoService defines the contract for todo business logic
type TodoService interface {
	// CreateTodo handles todo creation workflow
	CreateTodo(ctx context.Context, req dto.CreateTodoRequest) (*models.Todo, error)

	// GetTodos retrieves todos for a user with pagination
	GetTodos(ctx context.Context, userID uint, page, pageSize int) (*dto.TodoListResponse, error)

	// GetTodosByCategoryID retrieves todos filtered by category ID with pagination
	GetTodosByCategoryID(ctx context.Context, categoryID uint, page, pageSize int) (*dto.TodoListResponse, error)

	// GetTodosGroupedByCategory retrieves all accessible todos grouped by category
	GetTodosGroupedByCategory(ctx context.Context, userID uint) (*dto.TodosGroupedByCategoryResponse, error)

	// GetTodoByID retrieves a single todo with ownership/permission verification
	GetTodoByID(ctx context.Context, req dto.GetTodoRequest) (*models.Todo, error)

	// UpdateTodo handles todo update with ownership/permission verification
	UpdateTodo(ctx context.Context, req dto.UpdateTodoRequest) (*models.Todo, error)

	// DeleteTodo handles todo soft deletion with ownership/permission verification
	DeleteTodo(ctx context.Context, req dto.DeleteTodoRequest) error
}

// AuthService defines the contract for auth business logic
type AuthService interface {
	// RegisterUser handles complete user registration including validation, hashing, and token generation
	RegisterUser(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error)

	// LoginUser handles user authentication including password verification and token generation
	LoginUser(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)

	// GetByID retrieves a user by ID (for internal use)
	GetByID(ctx context.Context, id uint) (*models.User, error)
}

// CategoryService defines the contract for category business logic
type CategoryService interface {
	// CreateCategory creates a new category for a user
	CreateCategory(ctx context.Context, req dto.CreateCategoryRequest) (*models.Category, error)

	// GetCategories retrieves all categories owned by a user
	GetCategories(ctx context.Context, userID uint) ([]models.Category, error)

	// GetCategoryByID retrieves a category by ID with ownership verification
	GetCategoryByID(ctx context.Context, categoryID, userID uint) (*models.Category, error)

	// UpdateCategory updates a category with ownership verification
	UpdateCategory(ctx context.Context, req dto.UpdateCategoryRequest) (*models.Category, error)

	// DeleteCategory deletes a category with ownership verification
	DeleteCategory(ctx context.Context, categoryID, userID uint) error

	// ShareCategory shares a category with another user
	ShareCategory(ctx context.Context, req dto.ShareCategoryRequest) (*models.CategoryShare, error)

	// UnshareCategory removes sharing of a category with a user
	UnshareCategory(ctx context.Context, req dto.UnshareCategoryRequest) error

	// UpdateSharePermission changes the permission of a shared category
	UpdateSharePermission(ctx context.Context, req dto.UpdateSharePermissionRequest) error

	// GetSharesForCategory gets all shares for a category (owner only)
	GetSharesForCategory(ctx context.Context, categoryID, userID uint) ([]models.CategoryShareWithUser, error)

	// GetSharedCategories gets all categories shared with a user
	GetSharedCategories(ctx context.Context, userID uint) ([]models.SharedCategoryWithOwner, error)

	// GetUserPermissionForCategory checks what permission a user has for a category
	GetUserPermissionForCategory(ctx context.Context, userID, categoryID uint) (string, error)
}
