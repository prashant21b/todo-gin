package middleware

import (
	"context"
	"log"

	"todo-app/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware injects a request ID into the context and response header
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate UUID
		rid := uuid.New().String()

		// Add to request context using typed key to avoid collisions
		ctx := context.WithValue(c.Request.Context(), utils.RequestIDKey, rid)
		c.Request = c.Request.WithContext(ctx)

		// Set in Gin context for convenience
		c.Set("requestID", rid)

		// Add response header
		c.Writer.Header().Set("X-Request-Id", rid)

		log.Printf("[RequestID] %s %s %s", rid, c.Request.Method, c.Request.URL.Path)

		c.Next()
	}
}
