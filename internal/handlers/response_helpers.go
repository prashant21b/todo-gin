package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Shared HTTP response helpers used across all handlers

// getUserID extracts userID from gin context, returns 0 and false if not found
func getUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, false
	}
	return userID.(uint), true
}

// parseIDParam parses ID from URL parameter
func parseIDParam(c *gin.Context, paramName string) (uint, error) {
	idParam := c.Param(paramName)
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// respondUnauthorized sends unauthorized response
func respondUnauthorized(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"success": false,
		"message": "User not authenticated",
	})
}

// respondBadRequest sends bad request response
func respondBadRequest(c *gin.Context, message string, err error) {
	response := gin.H{
		"success": false,
		"message": message,
	}
	if err != nil {
		response["error"] = err.Error()
	}
	c.JSON(http.StatusBadRequest, response)
}

// respondTimeout sends request timeout response
func respondTimeout(c *gin.Context) {
	c.JSON(http.StatusRequestTimeout, gin.H{
		"success": false,
		"message": "Request timeout",
	})
}

// respondNotFound sends not found response
func respondNotFound(c *gin.Context, resource string) {
	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"message": resource + " not found",
	})
}

// respondForbidden sends forbidden response
func respondForbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, gin.H{
		"success": false,
		"message": message,
	})
}

// respondInternalError sends internal server error response
func respondInternalError(c *gin.Context, message string, err error) {
	response := gin.H{
		"success": false,
		"message": message,
	}
	if err != nil {
		response["error"] = err.Error()
	}
	c.JSON(http.StatusInternalServerError, response)
}

// respondConflict sends conflict response (e.g., duplicate resource)
func respondConflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, gin.H{
		"success": false,
		"message": message,
	})
}

// respondUnauthorizedWithMessage sends unauthorized response with custom message
func respondUnauthorizedWithMessage(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"success": false,
		"message": message,
	})
}
