package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"todo-app/internal/dto"
	"todo-app/internal/models"
	"todo-app/internal/services"
	"todo-app/internal/services/mocks"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestTodoHandler_CreateTodo(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		userID         uint
		mockFunc       func(ctx context.Context, req dto.CreateTodoRequest) (*models.Todo, error)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "successful creation",
			requestBody: map[string]interface{}{
				"title":       "Test Todo",
				"description": "Test Description",
				"category":    "Work",
			},
			userID: 1,
			mockFunc: func(ctx context.Context, req dto.CreateTodoRequest) (*models.Todo, error) {
				return &models.Todo{
					ID:          1,
					Title:       req.Title,
					Description: req.Description,
					CategoryID:  1,
					UserID:      req.UserID,
				}, nil
			},
			expectedStatus: http.StatusCreated,
			expectedMsg:    "Todo created successfully",
		},
		{
			name: "validation error - missing title",
			requestBody: map[string]interface{}{
				"description": "Test Description",
				"category":    "Work",
			},
			userID:         1,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Validation failed",
		},
		{
			name: "validation error - missing category",
			requestBody: map[string]interface{}{
				"title":       "Test Todo",
				"description": "Test Description",
			},
			userID:         1,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "either category or category_id is required",
		},
		{
			name: "service error",
			requestBody: map[string]interface{}{
				"title":    "Test Todo",
				"category": "Work",
			},
			userID: 1,
			mockFunc: func(ctx context.Context, req dto.CreateTodoRequest) (*models.Todo, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "Failed to create todo",
		},
		{
			name: "validation error - whitespace only title",
			requestBody: map[string]interface{}{
				"title":    "   ",
				"category": "Work",
			},
			userID:         1,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "title cannot be empty or whitespace only",
		},
		{
			name: "validation error - title too long",
			requestBody: map[string]interface{}{
				"title":    string(make([]byte, 300)),
				"category": "Work",
			},
			userID:         1,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.MockTodoService{
				CreateTodoFunc: tt.mockFunc,
			}
			handler := NewTodoHandler(mockService)

			router := gin.New()
			router.POST("/todos", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.CreateTodo(c)
			})

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("CreateTodo() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			if msg, ok := response["message"].(string); ok {
				if msg != tt.expectedMsg {
					t.Errorf("CreateTodo() message = %v, want %v", msg, tt.expectedMsg)
				}
			}
		})
	}
}

func TestTodoHandler_GetTodos(t *testing.T) {
	tests := []struct {
		name           string
		userID         uint
		queryParams    string
		mockFunc       func(ctx context.Context, userID uint, page, pageSize int) (*dto.TodoListResponse, error)
		expectedStatus int
		expectedCount  int
	}{
		{
			name:        "successful retrieval",
			userID:      1,
			queryParams: "",
			mockFunc: func(ctx context.Context, userID uint, page, pageSize int) (*dto.TodoListResponse, error) {
				return &dto.TodoListResponse{
					Todos: []models.Todo{
						{ID: 1, Title: "Todo 1", CategoryID: 1, UserID: userID},
						{ID: 2, Title: "Todo 2", CategoryID: 2, UserID: userID},
					},
					Total:      2,
					Page:       page,
					PageSize:   pageSize,
					TotalPages: 1,
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:        "with pagination",
			userID:      1,
			queryParams: "?page=1&page_size=5",
			mockFunc: func(ctx context.Context, userID uint, page, pageSize int) (*dto.TodoListResponse, error) {
				if page != 1 || pageSize != 5 {
					t.Errorf("Expected page=1, pageSize=5, got page=%d, pageSize=%d", page, pageSize)
				}
				return &dto.TodoListResponse{
					Todos: []models.Todo{
						{ID: 1, Title: "Todo 1", CategoryID: 1, UserID: userID},
					},
					Total:      10,
					Page:       page,
					PageSize:   pageSize,
					TotalPages: 2,
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:        "service error",
			userID:      1,
			queryParams: "",
			mockFunc: func(ctx context.Context, userID uint, page, pageSize int) (*dto.TodoListResponse, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.MockTodoService{
				GetTodosFunc: tt.mockFunc,
			}
			handler := NewTodoHandler(mockService)

			router := gin.New()
			router.GET("/todos", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.GetTodos(c)
			})

			req, _ := http.NewRequest(http.MethodGet, "/todos"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("GetTodos() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)

				count := int(response["count"].(float64))
				if count != tt.expectedCount {
					t.Errorf("GetTodos() count = %v, want %v", count, tt.expectedCount)
				}
			}
		})
	}
}

func TestTodoHandler_GetTodo(t *testing.T) {
	tests := []struct {
		name           string
		todoID         string
		userID         uint
		mockFunc       func(ctx context.Context, req dto.GetTodoRequest) (*models.Todo, error)
		expectedStatus int
	}{
		{
			name:   "successful retrieval",
			todoID: "1",
			userID: 1,
			mockFunc: func(ctx context.Context, req dto.GetTodoRequest) (*models.Todo, error) {
				return &models.Todo{
					ID:         req.ID,
					Title:      "Test Todo",
					CategoryID: 1,
					UserID:     1,
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "invalid id",
			todoID: "invalid",
			userID: 1,
			mockFunc: func(ctx context.Context, req dto.GetTodoRequest) (*models.Todo, error) {
				return nil, nil
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "not found",
			todoID: "999",
			userID: 1,
			mockFunc: func(ctx context.Context, req dto.GetTodoRequest) (*models.Todo, error) {
				return nil, services.ErrTodoNotFound
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "forbidden - different user",
			todoID: "1",
			userID: 2,
			mockFunc: func(ctx context.Context, req dto.GetTodoRequest) (*models.Todo, error) {
				return nil, services.ErrForbidden
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.MockTodoService{
				GetTodoByIDFunc: tt.mockFunc,
			}
			handler := NewTodoHandler(mockService)

			router := gin.New()
			router.GET("/todos/:id", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.GetTodo(c)
			})

			req, _ := http.NewRequest(http.MethodGet, "/todos/"+tt.todoID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("GetTodo() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestTodoHandler_UpdateTodo(t *testing.T) {
	tests := []struct {
		name           string
		todoID         string
		userID         uint
		requestBody    map[string]interface{}
		updateFunc     func(ctx context.Context, req dto.UpdateTodoRequest) (*models.Todo, error)
		expectedStatus int
	}{
		{
			name:   "successful update",
			todoID: "1",
			userID: 1,
			requestBody: map[string]interface{}{
				"title":     "Updated Title",
				"completed": true,
			},
			updateFunc: func(ctx context.Context, req dto.UpdateTodoRequest) (*models.Todo, error) {
				return &models.Todo{
					ID:         req.ID,
					Title:      *req.Title,
					CategoryID: 1,
					Completed:  *req.Completed,
					UserID:     req.UserID,
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "successful category_id update",
			todoID: "1",
			userID: 1,
			requestBody: map[string]interface{}{
				"category_id": 2,
			},
			updateFunc: func(ctx context.Context, req dto.UpdateTodoRequest) (*models.Todo, error) {
				if req.CategoryID == nil || *req.CategoryID != 2 {
					t.Errorf("Expected category_id to be 2, got %v", req.CategoryID)
				}
				return &models.Todo{
					ID:         req.ID,
					Title:      "Original Title",
					CategoryID: *req.CategoryID,
					UserID:     req.UserID,
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "successful update with all fields",
			todoID: "1",
			userID: 1,
			requestBody: map[string]interface{}{
				"title":       "Updated Title",
				"description": "Updated Description",
				"category_id": 3,
				"completed":   true,
			},
			updateFunc: func(ctx context.Context, req dto.UpdateTodoRequest) (*models.Todo, error) {
				if req.CategoryID == nil || *req.CategoryID != 3 {
					t.Errorf("Expected category_id to be 3, got %v", req.CategoryID)
				}
				return &models.Todo{
					ID:          req.ID,
					Title:       *req.Title,
					Description: *req.Description,
					CategoryID:  *req.CategoryID,
					Completed:   *req.Completed,
					UserID:      req.UserID,
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "not found",
			todoID: "999",
			userID: 1,
			requestBody: map[string]interface{}{
				"title": "Updated Title",
			},
			updateFunc: func(ctx context.Context, req dto.UpdateTodoRequest) (*models.Todo, error) {
				return nil, services.ErrTodoNotFound
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "forbidden - different user",
			todoID: "1",
			userID: 2,
			requestBody: map[string]interface{}{
				"title": "Updated Title",
			},
			updateFunc: func(ctx context.Context, req dto.UpdateTodoRequest) (*models.Todo, error) {
				return nil, services.ErrForbidden
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "validation error - empty body",
			todoID:         "1",
			userID:         1,
			requestBody:    map[string]interface{}{},
			updateFunc:     nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "validation error - whitespace only title",
			todoID: "1",
			userID: 1,
			requestBody: map[string]interface{}{
				"title": "   ",
			},
			updateFunc:     nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "validation error - title too long",
			todoID: "1",
			userID: 1,
			requestBody: map[string]interface{}{
				"title": string(make([]byte, 300)),
			},
			updateFunc:     nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.MockTodoService{
				UpdateTodoFunc: tt.updateFunc,
			}
			handler := NewTodoHandler(mockService)

			router := gin.New()
			router.PUT("/todos/:id", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.UpdateTodo(c)
			})

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPut, "/todos/"+tt.todoID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("UpdateTodo() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestTodoHandler_DeleteTodo(t *testing.T) {
	tests := []struct {
		name           string
		todoID         string
		userID         uint
		deleteFunc     func(ctx context.Context, req dto.DeleteTodoRequest) error
		expectedStatus int
	}{
		{
			name:   "successful deletion",
			todoID: "1",
			userID: 1,
			deleteFunc: func(ctx context.Context, req dto.DeleteTodoRequest) error {
				return nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "not found",
			todoID: "999",
			userID: 1,
			deleteFunc: func(ctx context.Context, req dto.DeleteTodoRequest) error {
				return services.ErrTodoNotFound
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "forbidden - different user",
			todoID: "1",
			userID: 2,
			deleteFunc: func(ctx context.Context, req dto.DeleteTodoRequest) error {
				return services.ErrForbidden
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.MockTodoService{
				DeleteTodoFunc: tt.deleteFunc,
			}
			handler := NewTodoHandler(mockService)

			router := gin.New()
			router.DELETE("/todos/:id", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.DeleteTodo(c)
			})

			req, _ := http.NewRequest(http.MethodDelete, "/todos/"+tt.todoID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("DeleteTodo() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}
