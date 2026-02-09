package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Headers demonstrates reading a custom request header and setting a custom response header
func Headers(c *gin.Context) {
	// Read a custom request header
	val := c.GetHeader("X-Custom-Header")

	// Echo back in response header and JSON
	c.Header("X-Echo-Custom", val)
	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"message":         "Headers received",
		"x_custom_header": val,
	})
}
