package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"todo-app/pkg/utils"

	"github.com/gin-gonic/gin"
)

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response header contains X-Request-Id
	requestID := w.Header().Get("X-Request-Id")
	if requestID == "" {
		t.Error("Expected X-Request-Id header to be set")
	}

	// Verify it's a valid UUID format (36 characters including hyphens)
	if len(requestID) != 36 {
		t.Errorf("Expected UUID format (36 chars), got %d chars: %s", len(requestID), requestID)
	}
}

func TestRequestIDMiddleware_InContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())

	var capturedRequestID string
	router.GET("/test", func(c *gin.Context) {
		// Get from Gin context
		if rid, exists := c.Get("requestID"); exists {
			capturedRequestID = rid.(string)
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that request ID was captured from context
	if capturedRequestID == "" {
		t.Error("Expected requestID to be set in Gin context")
	}

	// Check that it matches the header
	headerRequestID := w.Header().Get("X-Request-Id")
	if capturedRequestID != headerRequestID {
		t.Errorf("Context requestID (%s) doesn't match header (%s)", capturedRequestID, headerRequestID)
	}
}

func TestRequestIDMiddleware_InRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())

	var capturedRequestID string
	router.GET("/test", func(c *gin.Context) {
		// Get from request context using utils.GetRequestID
		capturedRequestID = utils.GetRequestID(c.Request.Context())
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that request ID was captured from request context
	if capturedRequestID == "" {
		t.Error("Expected requestID to be set in request context")
	}

	// Check that it matches the header
	headerRequestID := w.Header().Get("X-Request-Id")
	if capturedRequestID != headerRequestID {
		t.Errorf("Request context requestID (%s) doesn't match header (%s)", capturedRequestID, headerRequestID)
	}
}

func TestRequestIDMiddleware_UniqueIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make multiple requests and check that each gets a unique ID
	requestIDs := make(map[string]bool)
	for i := 0; i < 10; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		requestID := w.Header().Get("X-Request-Id")
		if requestIDs[requestID] {
			t.Errorf("Duplicate request ID found: %s", requestID)
		}
		requestIDs[requestID] = true
	}
}
