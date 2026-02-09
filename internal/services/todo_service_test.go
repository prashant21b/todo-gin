package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"todo-app/internal/dto"
	"todo-app/internal/models"
	"todo-app/internal/repository/mocks"
)

// Helper to create a TodoService with all required mocks
func createTestTodoService(
	todoRepo *mocks.MockTodoRepository,
	categoryRepo *mocks.MockCategoryRepository,
	categoryShareRepo *mocks.MockCategoryShareRepository,
) TodoService {
	if categoryRepo == nil {
		categoryRepo = &mocks.MockCategoryRepository{}
	}
	if categoryShareRepo == nil {
		categoryShareRepo = &mocks.MockCategoryShareRepository{}
	}
	return NewTodoService(todoRepo, categoryRepo, categoryShareRepo, PaginationConfig{DefaultPageSize: 10, MaxPageSize: 100})
}

// Default category mock that returns owner permission
func defaultCategoryMock(ownerID uint) *mocks.MockCategoryRepository {
	return &mocks.MockCategoryRepository{
		GetCategoryByIDFunc: func(ctx context.Context, id uint) (*models.Category, error) {
			return &models.Category{
				ID:      id,
				Name:    "Test Category",
				OwnerID: ownerID,
			}, nil
		},
	}
}

func TestTodoService_CreateTodo(t *testing.T) {
	tests := []struct {
		name           string
		req            dto.CreateTodoRequest
		todoMockFunc   func(ctx context.Context, todo *models.Todo) error
		categoryExists bool
		wantErr        bool
		wantID         uint
	}{
		{
			name: "successful creation - existing category",
			req: dto.CreateTodoRequest{
				Title:       "Test Todo",
				Description: "Test Description",
				Category:    "Work",
				UserID:      1,
			},
			todoMockFunc: func(ctx context.Context, todo *models.Todo) error {
				todo.ID = 1
				todo.CreatedAt = time.Now()
				todo.UpdatedAt = time.Now()
				return nil
			},
			categoryExists: true,
			wantErr:        false,
			wantID:         1,
		},
		{
			name: "successful creation - new category created",
			req: dto.CreateTodoRequest{
				Title:       "Test Todo",
				Description: "Test Description",
				Category:    "NewCategory",
				UserID:      1,
			},
			todoMockFunc: func(ctx context.Context, todo *models.Todo) error {
				todo.ID = 2
				todo.CreatedAt = time.Now()
				todo.UpdatedAt = time.Now()
				return nil
			},
			categoryExists: false,
			wantErr:        false,
			wantID:         2,
		},
		{
			name: "category required",
			req: dto.CreateTodoRequest{
				Title:  "Test Todo",
				UserID: 1,
			},
			wantErr: true,
		},
		{
			name: "repository error",
			req: dto.CreateTodoRequest{
				Title:    "Test Todo",
				Category: "Work",
				UserID:   1,
			},
			todoMockFunc: func(ctx context.Context, todo *models.Todo) error {
				return errors.New("database error")
			},
			categoryExists: true,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todoRepo := &mocks.MockTodoRepository{
				CreateTodoFunc: tt.todoMockFunc,
			}

			categoryRepo := &mocks.MockCategoryRepository{
				GetCategoryByNameAndOwnerFunc: func(ctx context.Context, ownerID uint, name string) (*models.Category, error) {
					if tt.categoryExists {
						return &models.Category{
							ID:      1,
							Name:    name,
							OwnerID: ownerID,
						}, nil
					}
					return nil, sql.ErrNoRows
				},
				CreateCategoryFunc: func(ctx context.Context, category *models.Category) error {
					category.ID = 2
					return nil
				},
			}

			service := createTestTodoService(todoRepo, categoryRepo, nil)

			todo, err := service.CreateTodo(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTodo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if todo.ID != tt.wantID {
					t.Errorf("CreateTodo() todo.ID = %v, want %v", todo.ID, tt.wantID)
				}
				if todo.UserID != tt.req.UserID {
					t.Errorf("CreateTodo() todo.UserID = %v, want %v", todo.UserID, tt.req.UserID)
				}
				if todo.CreatedBy != tt.req.UserID {
					t.Errorf("CreateTodo() todo.CreatedBy = %v, want %v", todo.CreatedBy, tt.req.UserID)
				}
			}
		})
	}
}

func TestTodoService_GetTodos(t *testing.T) {
	tests := []struct {
		name      string
		userID    uint
		page      int
		pageSize  int
		mockFunc  func(ctx context.Context, userID uint, page, pageSize int) ([]models.Todo, int64, error)
		wantCount int
		wantErr   bool
	}{
		{
			name:     "successful retrieval",
			userID:   1,
			page:     1,
			pageSize: 10,
			mockFunc: func(ctx context.Context, userID uint, page, pageSize int) ([]models.Todo, int64, error) {
				return []models.Todo{
					{ID: 1, Title: "Todo 1", UserID: userID, CategoryID: 1},
					{ID: 2, Title: "Todo 2", UserID: userID, CategoryID: 1},
				}, 2, nil
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:     "empty list",
			userID:   1,
			page:     1,
			pageSize: 10,
			mockFunc: func(ctx context.Context, userID uint, page, pageSize int) ([]models.Todo, int64, error) {
				return []models.Todo{}, 0, nil
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:     "repository error",
			userID:   1,
			page:     1,
			pageSize: 10,
			mockFunc: func(ctx context.Context, userID uint, page, pageSize int) ([]models.Todo, int64, error) {
				return nil, 0, errors.New("database error")
			},
			wantErr: true,
		},
		{
			name:     "pagination normalization - negative page",
			userID:   1,
			page:     -1,
			pageSize: 10,
			mockFunc: func(ctx context.Context, userID uint, page, pageSize int) ([]models.Todo, int64, error) {
				if page != 1 {
					t.Errorf("Expected page to be normalized to 1, got %d", page)
				}
				return []models.Todo{}, 0, nil
			},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockTodoRepository{
				GetTodosFunc: tt.mockFunc,
			}
			service := createTestTodoService(repo, nil, nil)

			result, err := service.GetTodos(context.Background(), tt.userID, tt.page, tt.pageSize)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetTodos() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(result.Todos) != tt.wantCount {
				t.Errorf("GetTodos() got %d todos, want %d", len(result.Todos), tt.wantCount)
			}
		})
	}
}

func TestTodoService_GetTodoByID(t *testing.T) {
	tests := []struct {
		name             string
		req              dto.GetTodoRequest
		todoMockFunc     func(ctx context.Context, id uint) (*models.Todo, error)
		categoryOwnerID  uint
		sharedPermission string
		wantErr          bool
		expectedErr      error
	}{
		{
			name: "successful retrieval - owner",
			req:  dto.GetTodoRequest{ID: 1, UserID: 1},
			todoMockFunc: func(ctx context.Context, id uint) (*models.Todo, error) {
				return &models.Todo{
					ID:         id,
					Title:      "Test Todo",
					UserID:     1,
					CategoryID: 1,
				}, nil
			},
			categoryOwnerID: 1,
			wantErr:         false,
		},
		{
			name: "successful retrieval - shared read",
			req:  dto.GetTodoRequest{ID: 1, UserID: 2},
			todoMockFunc: func(ctx context.Context, id uint) (*models.Todo, error) {
				return &models.Todo{
					ID:         id,
					Title:      "Test Todo",
					UserID:     1,
					CategoryID: 1,
				}, nil
			},
			categoryOwnerID:  1,
			sharedPermission: "read",
			wantErr:          false,
		},
		{
			name: "not found",
			req:  dto.GetTodoRequest{ID: 999, UserID: 1},
			todoMockFunc: func(ctx context.Context, id uint) (*models.Todo, error) {
				return nil, sql.ErrNoRows
			},
			wantErr:     true,
			expectedErr: ErrTodoNotFound,
		},
		{
			name: "forbidden - no permission",
			req:  dto.GetTodoRequest{ID: 1, UserID: 3},
			todoMockFunc: func(ctx context.Context, id uint) (*models.Todo, error) {
				return &models.Todo{
					ID:         id,
					Title:      "Test Todo",
					UserID:     1,
					CategoryID: 1,
				}, nil
			},
			categoryOwnerID:  1,
			sharedPermission: "none",
			wantErr:          true,
			expectedErr:      ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todoRepo := &mocks.MockTodoRepository{
				GetTodoByIDFunc: tt.todoMockFunc,
			}

			categoryRepo := &mocks.MockCategoryRepository{
				GetCategoryByIDFunc: func(ctx context.Context, id uint) (*models.Category, error) {
					return &models.Category{
						ID:      id,
						Name:    "Test Category",
						OwnerID: tt.categoryOwnerID,
					}, nil
				},
			}

			categoryShareRepo := &mocks.MockCategoryShareRepository{
				GetUserPermissionForCategoryFunc: func(ctx context.Context, userID, categoryID uint) (string, error) {
					if userID == tt.categoryOwnerID {
						return "owner", nil
					}
					return tt.sharedPermission, nil
				},
			}

			service := createTestTodoService(todoRepo, categoryRepo, categoryShareRepo)

			todo, err := service.GetTodoByID(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetTodoByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
				t.Errorf("GetTodoByID() error = %v, expected %v", err, tt.expectedErr)
			}

			if !tt.wantErr && todo.ID != tt.req.ID {
				t.Errorf("GetTodoByID() got ID %d, want %d", todo.ID, tt.req.ID)
			}
		})
	}
}

func TestTodoService_UpdateTodo(t *testing.T) {
	title := "Updated Title"
	completed := true

	tests := []struct {
		name             string
		req              dto.UpdateTodoRequest
		existingTodo     *models.Todo
		categoryOwnerID  uint
		sharedPermission string
		updateErr        error
		wantErr          bool
		expectedErr      error
	}{
		{
			name: "successful update - owner",
			req: dto.UpdateTodoRequest{
				ID:        1,
				UserID:    1,
				Title:     &title,
				Completed: &completed,
			},
			existingTodo: &models.Todo{
				ID:         1,
				Title:      "Original",
				UserID:     1,
				CategoryID: 1,
			},
			categoryOwnerID: 1,
			wantErr:         false,
		},
		{
			name: "successful update - shared write",
			req: dto.UpdateTodoRequest{
				ID:        1,
				UserID:    2,
				Title:     &title,
				Completed: &completed,
			},
			existingTodo: &models.Todo{
				ID:         1,
				Title:      "Original",
				UserID:     1,
				CategoryID: 1,
			},
			categoryOwnerID:  1,
			sharedPermission: "write",
			wantErr:          false,
		},
		{
			name: "forbidden - read only",
			req: dto.UpdateTodoRequest{
				ID:     1,
				UserID: 2,
				Title:  &title,
			},
			existingTodo: &models.Todo{
				ID:         1,
				Title:      "Original",
				UserID:     1,
				CategoryID: 1,
			},
			categoryOwnerID:  1,
			sharedPermission: "read",
			wantErr:          true,
			expectedErr:      ErrNoWritePermission,
		},
		{
			name: "not found",
			req: dto.UpdateTodoRequest{
				ID:     999,
				UserID: 1,
				Title:  &title,
			},
			existingTodo: nil,
			wantErr:      true,
			expectedErr:  ErrTodoNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todoRepo := &mocks.MockTodoRepository{
				GetTodoByIDFunc: func(ctx context.Context, id uint) (*models.Todo, error) {
					if tt.existingTodo == nil {
						return nil, sql.ErrNoRows
					}
					return tt.existingTodo, nil
				},
				UpdateTodoFunc: func(ctx context.Context, todo *models.Todo) error {
					return tt.updateErr
				},
			}

			categoryRepo := &mocks.MockCategoryRepository{
				GetCategoryByIDFunc: func(ctx context.Context, id uint) (*models.Category, error) {
					return &models.Category{
						ID:      id,
						Name:    "Test Category",
						OwnerID: tt.categoryOwnerID,
					}, nil
				},
			}

			categoryShareRepo := &mocks.MockCategoryShareRepository{
				GetUserPermissionForCategoryFunc: func(ctx context.Context, userID, categoryID uint) (string, error) {
					return tt.sharedPermission, nil
				},
			}

			service := createTestTodoService(todoRepo, categoryRepo, categoryShareRepo)

			todo, err := service.UpdateTodo(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateTodo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
				t.Errorf("UpdateTodo() error = %v, expected %v", err, tt.expectedErr)
			}

			if !tt.wantErr {
				if todo.Title != *tt.req.Title {
					t.Errorf("UpdateTodo() title = %v, want %v", todo.Title, *tt.req.Title)
				}
				if tt.req.Completed != nil && todo.Completed != *tt.req.Completed {
					t.Errorf("UpdateTodo() completed = %v, want %v", todo.Completed, *tt.req.Completed)
				}
			}
		})
	}
}

func TestTodoService_DeleteTodo(t *testing.T) {
	tests := []struct {
		name             string
		req              dto.DeleteTodoRequest
		existingTodo     *models.Todo
		categoryOwnerID  uint
		sharedPermission string
		deleteErr        error
		wantErr          bool
		expectedErr      error
	}{
		{
			name: "successful delete - owner",
			req:  dto.DeleteTodoRequest{ID: 1, UserID: 1},
			existingTodo: &models.Todo{
				ID:         1,
				Title:      "Test",
				UserID:     1,
				CategoryID: 1,
			},
			categoryOwnerID: 1,
			wantErr:         false,
		},
		{
			name: "successful delete - shared write",
			req:  dto.DeleteTodoRequest{ID: 1, UserID: 2},
			existingTodo: &models.Todo{
				ID:         1,
				Title:      "Test",
				UserID:     1,
				CategoryID: 1,
			},
			categoryOwnerID:  1,
			sharedPermission: "write",
			wantErr:          false,
		},
		{
			name: "forbidden - read only",
			req:  dto.DeleteTodoRequest{ID: 1, UserID: 2},
			existingTodo: &models.Todo{
				ID:         1,
				Title:      "Test",
				UserID:     1,
				CategoryID: 1,
			},
			categoryOwnerID:  1,
			sharedPermission: "read",
			wantErr:          true,
			expectedErr:      ErrNoWritePermission,
		},
		{
			name:         "not found",
			req:          dto.DeleteTodoRequest{ID: 999, UserID: 1},
			existingTodo: nil,
			wantErr:      true,
			expectedErr:  ErrTodoNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todoRepo := &mocks.MockTodoRepository{
				GetTodoByIDFunc: func(ctx context.Context, id uint) (*models.Todo, error) {
					if tt.existingTodo == nil {
						return nil, sql.ErrNoRows
					}
					return tt.existingTodo, nil
				},
				DeleteTodoFunc: func(ctx context.Context, id uint) error {
					return tt.deleteErr
				},
			}

			categoryRepo := &mocks.MockCategoryRepository{
				GetCategoryByIDFunc: func(ctx context.Context, id uint) (*models.Category, error) {
					return &models.Category{
						ID:      id,
						Name:    "Test Category",
						OwnerID: tt.categoryOwnerID,
					}, nil
				},
			}

			categoryShareRepo := &mocks.MockCategoryShareRepository{
				GetUserPermissionForCategoryFunc: func(ctx context.Context, userID, categoryID uint) (string, error) {
					return tt.sharedPermission, nil
				},
			}

			service := createTestTodoService(todoRepo, categoryRepo, categoryShareRepo)

			err := service.DeleteTodo(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteTodo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
				t.Errorf("DeleteTodo() error = %v, expected %v", err, tt.expectedErr)
			}
		})
	}
}

func TestTodoService_GetOrCreateCategory(t *testing.T) {
	tests := []struct {
		name               string
		categoryExists     bool
		existingCategoryID uint
		createCategoryErr  error
		newCategoryID      uint
		wantErr            bool
		wantCategoryID     uint
	}{
		{
			name:               "returns existing category",
			categoryExists:     true,
			existingCategoryID: 1,
			wantErr:            false,
			wantCategoryID:     1,
		},
		{
			name:              "creates new category if not exists",
			categoryExists:    false,
			newCategoryID:     5,
			createCategoryErr: nil,
			wantErr:           false,
			wantCategoryID:    5,
		},
		{
			name:              "handles category creation error",
			categoryExists:    false,
			createCategoryErr: errors.New("database error"),
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			categoryRepo := &mocks.MockCategoryRepository{
				GetCategoryByNameAndOwnerFunc: func(ctx context.Context, ownerID uint, name string) (*models.Category, error) {
					if tt.categoryExists {
						return &models.Category{
							ID:      tt.existingCategoryID,
							Name:    name,
							OwnerID: ownerID,
						}, nil
					}
					return nil, sql.ErrNoRows
				},
				CreateCategoryFunc: func(ctx context.Context, category *models.Category) error {
					if tt.createCategoryErr != nil {
						return tt.createCategoryErr
					}
					category.ID = tt.newCategoryID
					return nil
				},
			}

			todoRepo := &mocks.MockTodoRepository{
				CreateTodoFunc: func(ctx context.Context, todo *models.Todo) error {
					todo.ID = 1
					return nil
				},
			}

			service := createTestTodoService(todoRepo, categoryRepo, nil)

			todo, err := service.CreateTodo(context.Background(), dto.CreateTodoRequest{
				Title:    "Test",
				Category: "Work",
				UserID:   1,
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTodo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if todo.CategoryID != tt.wantCategoryID {
					t.Errorf("CreateTodo() todo.CategoryID = %v, want %v", todo.CategoryID, tt.wantCategoryID)
				}
			}
		})
	}
}
