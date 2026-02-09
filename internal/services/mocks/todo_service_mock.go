package mocks

import (
	"context"

	"todo-app/internal/dto"
	"todo-app/internal/models"
	"todo-app/internal/services"
)

// Ensure MockTodoService implements TodoService
var _ services.TodoService = (*MockTodoService)(nil)

// MockTodoService is a mock implementation of TodoService for testing
type MockTodoService struct {
	CreateTodoFunc                func(ctx context.Context, req dto.CreateTodoRequest) (*models.Todo, error)
	GetTodosFunc                  func(ctx context.Context, userID uint, page, pageSize int) (*dto.TodoListResponse, error)
	GetTodosByCategoryIDFunc      func(ctx context.Context, categoryID uint, page, pageSize int) (*dto.TodoListResponse, error)
	GetTodosGroupedByCategoryFunc func(ctx context.Context, userID uint) (*dto.TodosGroupedByCategoryResponse, error)
	GetTodoByIDFunc               func(ctx context.Context, req dto.GetTodoRequest) (*models.Todo, error)
	UpdateTodoFunc                func(ctx context.Context, req dto.UpdateTodoRequest) (*models.Todo, error)
	DeleteTodoFunc                func(ctx context.Context, req dto.DeleteTodoRequest) error
}

// CreateTodo calls the mock function
func (m *MockTodoService) CreateTodo(ctx context.Context, req dto.CreateTodoRequest) (*models.Todo, error) {
	if m.CreateTodoFunc != nil {
		return m.CreateTodoFunc(ctx, req)
	}
	return &models.Todo{}, nil
}

// GetTodos calls the mock function
func (m *MockTodoService) GetTodos(ctx context.Context, userID uint, page, pageSize int) (*dto.TodoListResponse, error) {
	if m.GetTodosFunc != nil {
		return m.GetTodosFunc(ctx, userID, page, pageSize)
	}
	return &dto.TodoListResponse{
		Todos:      []models.Todo{},
		Total:      0,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: 0,
	}, nil
}

// GetTodosByCategoryID calls the mock function
func (m *MockTodoService) GetTodosByCategoryID(ctx context.Context, categoryID uint, page, pageSize int) (*dto.TodoListResponse, error) {
	if m.GetTodosByCategoryIDFunc != nil {
		return m.GetTodosByCategoryIDFunc(ctx, categoryID, page, pageSize)
	}
	return &dto.TodoListResponse{
		Todos:      []models.Todo{},
		Total:      0,
		Page:       1,
		PageSize:   10,
		TotalPages: 0,
	}, nil
}

// GetTodoByID calls the mock function
func (m *MockTodoService) GetTodoByID(ctx context.Context, req dto.GetTodoRequest) (*models.Todo, error) {
	if m.GetTodoByIDFunc != nil {
		return m.GetTodoByIDFunc(ctx, req)
	}
	return nil, nil
}

// UpdateTodo calls the mock function
func (m *MockTodoService) UpdateTodo(ctx context.Context, req dto.UpdateTodoRequest) (*models.Todo, error) {
	if m.UpdateTodoFunc != nil {
		return m.UpdateTodoFunc(ctx, req)
	}
	return &models.Todo{}, nil
}

// DeleteTodo calls the mock function
func (m *MockTodoService) DeleteTodo(ctx context.Context, req dto.DeleteTodoRequest) error {
	if m.DeleteTodoFunc != nil {
		return m.DeleteTodoFunc(ctx, req)
	}
	return nil
}

// GetTodosGroupedByCategory calls the mock function
func (m *MockTodoService) GetTodosGroupedByCategory(ctx context.Context, userID uint) (*dto.TodosGroupedByCategoryResponse, error) {
	if m.GetTodosGroupedByCategoryFunc != nil {
		return m.GetTodosGroupedByCategoryFunc(ctx, userID)
	}
	return &dto.TodosGroupedByCategoryResponse{
		Categories: []dto.CategoryWithTodos{},
	}, nil
}
