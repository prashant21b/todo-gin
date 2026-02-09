package repository

import (
	"context"

	"todo-app/internal/models"
)

// TodoRepository defines persistence operations for todos
type TodoRepository interface {
	CreateTodo(ctx context.Context, todo *models.Todo) error
	GetTodos(ctx context.Context, userID uint, page, pageSize int) ([]models.Todo, int64, error)
	GetTodosByCategoryID(ctx context.Context, categoryID uint, page, pageSize int) ([]models.Todo, int64, error)
	GetTodoByID(ctx context.Context, id uint) (*models.Todo, error)
	UpdateTodo(ctx context.Context, todo *models.Todo) error
	DeleteTodo(ctx context.Context, id uint) error
}

// UserRepository defines persistence operations for users
type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id uint) (*models.User, error)
}

// CategoryRepository defines persistence operations for categories
type CategoryRepository interface {
	CreateCategory(ctx context.Context, category *models.Category) error
	GetCategoryByID(ctx context.Context, id uint) (*models.Category, error)
	GetCategoriesByOwnerID(ctx context.Context, ownerID uint) ([]models.Category, error)
	GetCategoryByNameAndOwner(ctx context.Context, ownerID uint, name string) (*models.Category, error)
	UpdateCategory(ctx context.Context, category *models.Category) error
	DeleteCategory(ctx context.Context, id uint) error
}

// CategoryShareRepository defines persistence operations for category shares
type CategoryShareRepository interface {
	CreateCategoryShare(ctx context.Context, share *models.CategoryShare) error
	GetCategoryShareByID(ctx context.Context, id uint) (*models.CategoryShare, error)
	GetCategoryShareByCategoryAndUser(ctx context.Context, categoryID, userID uint) (*models.CategoryShare, error)
	GetSharesForCategory(ctx context.Context, categoryID uint) ([]models.CategoryShareWithUser, error)
	GetSharedCategoriesForUser(ctx context.Context, userID uint) ([]models.SharedCategoryWithOwner, error)
	UpdateCategorySharePermission(ctx context.Context, id uint, permission models.Permission) error
	DeleteCategoryShare(ctx context.Context, id uint) error
	DeleteCategoryShareByUserAndCategory(ctx context.Context, categoryID, userID uint) error
	GetUserPermissionForCategory(ctx context.Context, userID, categoryID uint) (string, error)
	GetTodosGroupedByCategory(ctx context.Context, userID uint) ([]models.CategoryWithTodosRow, error)
}
