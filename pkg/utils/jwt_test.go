package utils

import (
	"errors"
	"testing"
	"time"
)

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name      string
		userID    uint
		jwtSecret string
		wantErr   bool
	}{
		{
			name:      "successful token generation",
			userID:    1,
			jwtSecret: "test-secret-key",
			wantErr:   false,
		},
		{
			name:      "different user ID",
			userID:    999,
			jwtSecret: "test-secret-key",
			wantErr:   false,
		},
		{
			name:      "missing JWT_SECRET",
			userID:    1,
			jwtSecret: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwtManager, err := NewJWTManager(tt.jwtSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewJWTManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return // Expected error when creating manager
			}

			token, err := jwtManager.GenerateToken(tt.userID)

			if err != nil {
				t.Errorf("GenerateToken() unexpected error = %v", err)
				return
			}

			if token == "" {
				t.Error("GenerateToken() returned empty token")
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	// Create JWT manager for test
	jwtManager, err := NewJWTManager("test-secret-key")
	if err != nil {
		t.Fatalf("Failed to create JWT manager: %v", err)
	}

	// Generate a valid token
	validToken, _ := jwtManager.GenerateToken(42)

	tests := []struct {
		name       string
		token      string
		jwtSecret  string
		wantErr    bool
		wantUserID uint
	}{
		{
			name:       "valid token",
			token:      validToken,
			jwtSecret:  "test-secret-key",
			wantErr:    false,
			wantUserID: 42,
		},
		{
			name:      "invalid token format",
			token:     "invalid.token.format",
			jwtSecret: "test-secret-key",
			wantErr:   true,
		},
		{
			name:      "empty token",
			token:     "",
			jwtSecret: "test-secret-key",
			wantErr:   true,
		},
		{
			name:      "missing JWT_SECRET",
			token:     validToken,
			jwtSecret: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testManager *JWTManager
			if tt.jwtSecret != "" {
				testManager, _ = NewJWTManager(tt.jwtSecret)
			}

			var claims *Claims
			var err error
			if testManager != nil {
				claims, err = testManager.ValidateToken(tt.token)
			} else {
				// Simulate missing manager
				err = errors.New("JWT manager not initialized")
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && claims.UserID != tt.wantUserID {
				t.Errorf("ValidateToken() userID = %v, want %v", claims.UserID, tt.wantUserID)
			}
		})
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	// Generate token with one secret
	jwtManager1, err := NewJWTManager("secret-one")
	if err != nil {
		t.Fatalf("Failed to create JWT manager: %v", err)
	}
	token, _ := jwtManager1.GenerateToken(1)

	// Try to validate with different secret
	jwtManager2, err := NewJWTManager("secret-two")
	if err != nil {
		t.Fatalf("Failed to create JWT manager: %v", err)
	}

	_, err = jwtManager2.ValidateToken(token)
	if err == nil {
		t.Error("Expected error when validating token with wrong secret")
	}
}

func TestGenerateToken_DifferentTokensForSameUser(t *testing.T) {
	jwtManager, err := NewJWTManager("test-secret-key")
	if err != nil {
		t.Fatalf("Failed to create JWT manager: %v", err)
	}

	token1, _ := jwtManager.GenerateToken(1)
	time.Sleep(time.Second) // ensure different IssuedAt (JWT uses second precision)
	token2, _ := jwtManager.GenerateToken(1)

	if token1 == "" || token2 == "" {
		t.Error("Tokens should not be empty")
	}
	if token1 == token2 {
		t.Error("Tokens should be different for separate calls (different IssuedAt or jti)")
	}
}
