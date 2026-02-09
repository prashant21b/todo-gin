package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"todo-app/pkg/utils"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuthMiddleware(t *testing.T) {
	// Create JWT manager for testing
	jwtManager, err := utils.NewJWTManager("test-secret-key")
	if err != nil {
		t.Fatalf("Failed to create JWT manager: %v", err)
	}

	// Generate a valid token for testing
	validToken, _ := jwtManager.GenerateToken(1)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedUserID uint
		shouldPass     bool
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectedUserID: 1,
			shouldPass:     true,
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			shouldPass:     false,
		},
		{
			name:           "invalid format - no Bearer prefix",
			authHeader:     validToken,
			expectedStatus: http.StatusUnauthorized,
			shouldPass:     false,
		},
		{
			name:           "invalid format - wrong prefix",
			authHeader:     "Basic " + validToken,
			expectedStatus: http.StatusUnauthorized,
			shouldPass:     false,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			shouldPass:     false,
		},
		{
			name:           "empty token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			shouldPass:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(AuthMiddleware(jwtManager))
			router.GET("/protected", func(c *gin.Context) {
				userID, exists := c.Get("userID")
				if !exists {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "userID not set"})
					return
				}
				c.JSON(http.StatusOK, gin.H{"userID": userID})
			})

			req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("AuthMiddleware() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestAuthMiddleware_UserIDInContext(t *testing.T) {
	// Create JWT manager for testing
	jwtManager, err := utils.NewJWTManager("test-secret-key")
	if err != nil {
		t.Fatalf("Failed to create JWT manager: %v", err)
	}

	// Generate a token for user ID 42
	token, _ := jwtManager.GenerateToken(42)

	router := gin.New()
	router.Use(AuthMiddleware(jwtManager))

	var capturedUserID uint
	router.GET("/protected", func(c *gin.Context) {
		userID, _ := c.Get("userID")
		capturedUserID = userID.(uint)
		c.JSON(http.StatusOK, gin.H{"userID": capturedUserID})
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", w.Code)
	}

	if capturedUserID != 42 {
		t.Errorf("Expected userID 42, got %v", capturedUserID)
	}
}
