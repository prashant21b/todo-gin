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

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]any
		mockFunc       func(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "successful registration",
			requestBody: map[string]any{
				"name":     "John Doe",
				"email":    "john@example.com",
				"password": "password123",
			},
			mockFunc: func(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
				return &dto.AuthResponse{
					User: &models.User{
						ID:    1,
						Name:  req.Name,
						Email: req.Email,
					},
					Token: "test-token-123",
				}, nil
			},
			expectedStatus: http.StatusCreated,
			expectedMsg:    "User registered successfully",
		},
		{
			name: "email already exists",
			requestBody: map[string]any{
				"name":     "John Doe",
				"email":    "existing@example.com",
				"password": "password123",
			},
			mockFunc: func(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
				return nil, services.ErrEmailAlreadyRegistered
			},
			expectedStatus: http.StatusConflict,
			expectedMsg:    "email already registered",
		},
		{
			name: "invalid input - missing name",
			requestBody: map[string]any{
				"email":    "john@example.com",
				"password": "password123",
			},
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Validation failed",
		},
		{
			name: "invalid input - invalid email",
			requestBody: map[string]any{
				"name":     "John Doe",
				"email":    "invalid-email",
				"password": "password123",
			},
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Validation failed",
		},
		{
			name: "invalid input - short password",
			requestBody: map[string]any{
				"name":     "John Doe",
				"email":    "john@example.com",
				"password": "123",
			},
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Validation failed",
		},
		{
			name: "service error",
			requestBody: map[string]any{
				"name":     "John Doe",
				"email":    "john@example.com",
				"password": "password123",
			},
			mockFunc: func(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
				return nil, errors.New("internal server error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "Failed to register",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.MockAuthService{
				RegisterUserFunc: tt.mockFunc,
			}
			handler := NewAuthHandler(mockService)

			router := gin.New()
			router.POST("/register", handler.Register)

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Register() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			var response map[string]any
			json.Unmarshal(w.Body.Bytes(), &response)

			if msg, ok := response["message"].(string); ok {
				if msg != tt.expectedMsg {
					t.Errorf("Register() message = %v, want %v", msg, tt.expectedMsg)
				}
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]any
		mockFunc       func(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "successful login",
			requestBody: map[string]any{
				"email":    "john@example.com",
				"password": "password123",
			},
			mockFunc: func(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
				return &dto.AuthResponse{
					User: &models.User{
						ID:    1,
						Name:  "John Doe",
						Email: req.Email,
					},
					Token: "test-token-123",
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "Login successful",
		},
		{
			name: "invalid credentials",
			requestBody: map[string]any{
				"email":    "john@example.com",
				"password": "wrongpassword",
			},
			mockFunc: func(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
				return nil, services.ErrInvalidCredentials
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "invalid email or password",
		},
		{
			name: "invalid input - missing email",
			requestBody: map[string]any{
				"password": "password123",
			},
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Validation failed",
		},
		{
			name: "invalid input - invalid email format",
			requestBody: map[string]any{
				"email":    "invalid-email",
				"password": "password123",
			},
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Validation failed",
		},
		{
			name: "service error",
			requestBody: map[string]any{
				"email":    "john@example.com",
				"password": "password123",
			},
			mockFunc: func(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "Failed to login",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.MockAuthService{
				LoginUserFunc: tt.mockFunc,
			}
			handler := NewAuthHandler(mockService)

			router := gin.New()
			router.POST("/login", handler.Login)

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Login() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			var response map[string]any
			json.Unmarshal(w.Body.Bytes(), &response)

			if msg, ok := response["message"].(string); ok {
				if msg != tt.expectedMsg {
					t.Errorf("Login() message = %v, want %v", msg, tt.expectedMsg)
				}
			}
		})
	}
}
