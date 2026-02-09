package mocks

import (
	"context"

	"todo-app/internal/models"
	"todo-app/internal/repository"
)

// Ensure MockTodoRepository implements TodoRepository
var _ repository.TodoRepository = (*MockTodoRepository)(nil)

// MockTodoRepository is a mock implementation of TodoRepository for testing
type MockTodoRepository struct {
	CreateTodoFunc           func(ctx context.Context, todo *models.Todo) error
	GetTodosFunc             func(ctx context.Context, userID uint, page, pageSize int) ([]models.Todo, int64, error)
	GetTodosByCategoryIDFunc func(ctx context.Context, categoryID uint, page, pageSize int) ([]models.Todo, int64, error)
	GetTodoByIDFunc          func(ctx context.Context, id uint) (*models.Todo, error)
	UpdateTodoFunc           func(ctx context.Context, todo *models.Todo) error
	DeleteTodoFunc           func(ctx context.Context, id uint) error
}

// CreateTodo calls the mock function
func (m *MockTodoRepository) CreateTodo(ctx context.Context, todo *models.Todo) error {
	if m.CreateTodoFunc != nil {
		return m.CreateTodoFunc(ctx, todo)
	}
	return nil
}

// GetTodos calls the mock function
func (m *MockTodoRepository) GetTodos(ctx context.Context, userID uint, page, pageSize int) ([]models.Todo, int64, error) {
	if m.GetTodosFunc != nil {
		return m.GetTodosFunc(ctx, userID, page, pageSize)
	}
	return []models.Todo{}, 0, nil
}

// GetTodosByCategoryID calls the mock function
func (m *MockTodoRepository) GetTodosByCategoryID(ctx context.Context, categoryID uint, page, pageSize int) ([]models.Todo, int64, error) {
	if m.GetTodosByCategoryIDFunc != nil {
		return m.GetTodosByCategoryIDFunc(ctx, categoryID, page, pageSize)
	}
	return []models.Todo{}, 0, nil
}

// GetTodoByID calls the mock function
func (m *MockTodoRepository) GetTodoByID(ctx context.Context, id uint) (*models.Todo, error) {
	if m.GetTodoByIDFunc != nil {
		return m.GetTodoByIDFunc(ctx, id)
	}
	return nil, nil
}

// UpdateTodo calls the mock function
func (m *MockTodoRepository) UpdateTodo(ctx context.Context, todo *models.Todo) error {
	if m.UpdateTodoFunc != nil {
		return m.UpdateTodoFunc(ctx, todo)
	}
	return nil
}

// DeleteTodo calls the mock function
func (m *MockTodoRepository) DeleteTodo(ctx context.Context, id uint) error {
	if m.DeleteTodoFunc != nil {
		return m.DeleteTodoFunc(ctx, id)
	}
	return nil
}
