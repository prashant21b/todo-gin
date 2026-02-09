package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"todo-app/internal/dto"
	"todo-app/internal/models"
	"todo-app/internal/repository/mocks"
	"todo-app/pkg/utils"
)

func TestAuthService_RegisterUser(t *testing.T) {
	// Create JWT manager for testing
	jwtManager, err := utils.NewJWTManager("test-secret-key")
	if err != nil {
		t.Fatalf("Failed to create JWT manager: %v", err)
	}

	tests := []struct {
		name             string
		request          dto.RegisterRequest
		getByEmailFunc   func(ctx context.Context, email string) (*models.User, error)
		createUserFunc   func(ctx context.Context, user *models.User) error
		wantErr          bool
		expectedErrorMsg string
	}{
		{
			name: "successful registration",
			request: dto.RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			getByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return nil, errors.New("not found") // User doesn't exist
			},
			createUserFunc: func(ctx context.Context, user *models.User) error {
				user.ID = 1
				user.CreatedAt = time.Now()
				user.UpdatedAt = time.Now()
				return nil
			},
			wantErr: false,
		},
		{
			name: "email already registered",
			request: dto.RegisterRequest{
				Name:     "John Doe",
				Email:    "existing@example.com",
				Password: "password123",
			},
			getByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return &models.User{ID: 1, Email: email}, nil // User exists
			},
			createUserFunc:   nil,
			wantErr:          true,
			expectedErrorMsg: "email already registered",
		},
		{
			name: "database error",
			request: dto.RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			getByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return nil, errors.New("not found")
			},
			createUserFunc: func(ctx context.Context, user *models.User) error {
				return errors.New("database error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockUserRepository{
				GetUserByEmailFunc: tt.getByEmailFunc,
				CreateUserFunc:     tt.createUserFunc,
			}
			service := NewAuthService(mockRepo, jwtManager)

			response, err := service.RegisterUser(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.expectedErrorMsg != "" && err.Error() != tt.expectedErrorMsg {
					t.Errorf("RegisterUser() error = %v, want %v", err.Error(), tt.expectedErrorMsg)
				}
				return
			}

			if response == nil {
				t.Error("RegisterUser() returned nil response")
				return
			}

			if response.User == nil {
				t.Error("RegisterUser() returned nil user")
				return
			}

			if response.Token == "" {
				t.Error("RegisterUser() returned empty token")
			}
		})
	}
}

func TestAuthService_LoginUser(t *testing.T) {
	// Create JWT manager for testing
	jwtManager, err := utils.NewJWTManager("test-secret-key")
	if err != nil {
		t.Fatalf("Failed to create JWT manager: %v", err)
	}

	// Hash a test password
	hashedPassword, _ := utils.HashPassword("password123")

	tests := []struct {
		name             string
		request          dto.LoginRequest
		getByEmailFunc   func(ctx context.Context, email string) (*models.User, error)
		wantErr          bool
		expectedErrorMsg string
	}{
		{
			name: "successful login",
			request: dto.LoginRequest{
				Email:    "john@example.com",
				Password: "password123",
			},
			getByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return &models.User{
					ID:       1,
					Name:     "John Doe",
					Email:    email,
					Password: hashedPassword,
				}, nil
			},
			wantErr: false,
		},
		{
			name: "user not found",
			request: dto.LoginRequest{
				Email:    "notfound@example.com",
				Password: "password123",
			},
			getByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return nil, errors.New("not found")
			},
			wantErr:          true,
			expectedErrorMsg: "invalid email or password",
		},
		{
			name: "wrong password",
			request: dto.LoginRequest{
				Email:    "john@example.com",
				Password: "wrongpassword",
			},
			getByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return &models.User{
					ID:       1,
					Name:     "John Doe",
					Email:    email,
					Password: hashedPassword,
				}, nil
			},
			wantErr:          true,
			expectedErrorMsg: "invalid email or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockUserRepository{
				GetUserByEmailFunc: tt.getByEmailFunc,
			}
			service := NewAuthService(mockRepo, jwtManager)

			response, err := service.LoginUser(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("LoginUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err.Error() != tt.expectedErrorMsg {
					t.Errorf("LoginUser() error = %v, want %v", err.Error(), tt.expectedErrorMsg)
				}
				return
			}

			if response == nil {
				t.Error("LoginUser() returned nil response")
				return
			}

			if response.User == nil {
				t.Error("LoginUser() returned nil user")
				return
			}

			if response.Token == "" {
				t.Error("LoginUser() returned empty token")
			}
		})
	}
}

func TestAuthService_GetByID(t *testing.T) {
	// Create JWT manager for testing
	jwtManager, err := utils.NewJWTManager("test-secret-key")
	if err != nil {
		t.Fatalf("Failed to create JWT manager: %v", err)
	}

	tests := []struct {
		name     string
		userID   uint
		mockFunc func(ctx context.Context, id uint) (*models.User, error)
		wantErr  bool
	}{
		{
			name:   "user found",
			userID: 1,
			mockFunc: func(ctx context.Context, id uint) (*models.User, error) {
				return &models.User{
					ID:    id,
					Name:  "John Doe",
					Email: "john@example.com",
				}, nil
			},
			wantErr: false,
		},
		{
			name:   "user not found",
			userID: 999,
			mockFunc: func(ctx context.Context, id uint) (*models.User, error) {
				return nil, errors.New("not found")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.MockUserRepository{
				GetUserByIDFunc: tt.mockFunc,
			}
			service := NewAuthService(mockRepo, jwtManager)

			user, err := service.GetByID(context.Background(), tt.userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && user.ID != tt.userID {
				t.Errorf("GetByID() user.ID = %v, want %v", user.ID, tt.userID)
			}
		})
	}
}
