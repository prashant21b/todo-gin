package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"todo-app/internal/dto"
	"todo-app/internal/services"
	"todo-app/pkg/utils"

	"github.com/gin-gonic/gin"
)

// TodoHandler handles HTTP requests for todos
type TodoHandler struct {
	todoService services.TodoService
}

// NewTodoHandler creates a new TodoHandler with the provided service
func NewTodoHandler(svc services.TodoService) *TodoHandler {
	return &TodoHandler{todoService: svc}
}

// CreateTodoInput represents the create todo request body
type CreateTodoInput struct {
	Title       string `json:"title" binding:"required,min=1,max=255"`
	Description string `json:"description" binding:"max=1000"`
	Category    string `json:"category" binding:"-"`            // Validated in Validate(); optional when category_id is set
	CategoryID  *uint  `json:"category_id" binding:"omitempty"` // ID: use this category (must have write access)
}

// Validate performs custom validation on CreateTodoInput
func (c *CreateTodoInput) Validate() error {
	c.Title = strings.TrimSpace(c.Title)
	if c.Title == "" {
		return errors.New("title cannot be empty or whitespace only")
	}
	c.Description = strings.TrimSpace(c.Description)
	c.Category = strings.TrimSpace(c.Category)
	// Require either category_id or category name
	hasID := c.CategoryID != nil && *c.CategoryID > 0
	if !hasID && c.Category == "" {
		return errors.New("either category or category_id is required")
	}
	return nil
}

// UpdateTodoInput represents the update todo request body
type UpdateTodoInput struct {
	Title       *string `json:"title" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
	CategoryID  *uint   `json:"category_id"`
	Completed   *bool   `json:"completed"`
}

// IsEmpty returns true if no fields are provided for update
func (u *UpdateTodoInput) IsEmpty() bool {
	return u.Title == nil && u.Description == nil && u.CategoryID == nil && u.Completed == nil
}

// Validate performs custom validation on UpdateTodoInput
func (u *UpdateTodoInput) Validate() error {
	if u.IsEmpty() {
		return errors.New("at least one field must be provided for update")
	}
	if u.Title != nil {
		trimmed := strings.TrimSpace(*u.Title)
		if trimmed == "" {
			return errors.New("title cannot be empty or whitespace only")
		}
		u.Title = &trimmed
	}
	if u.Description != nil {
		trimmed := strings.TrimSpace(*u.Description)
		u.Description = &trimmed
	}
	return nil
}

// handleTodoError maps service errors to HTTP responses
func (h *TodoHandler) handleTodoError(c *gin.Context, ctx context.Context, err error, operation string, userID uint, todoID uint) bool {
	if err == nil {
		return false
	}

	// Check for timeout
	if ctx.Err() != nil {
		respondTimeout(c)
		return true
	}

	// Handle specific business errors
	if errors.Is(err, services.ErrTodoNotFound) {
		respondNotFound(c, "Todo")
		return true
	}

	if errors.Is(err, services.ErrForbidden) {
		respondForbidden(c, "You don't have permission to access this todo")
		return true
	}

	if errors.Is(err, services.ErrCategoryNotFound) {
		respondNotFound(c, "Category")
		return true
	}

	if errors.Is(err, services.ErrCategoryRequired) {
		respondBadRequest(c, "Category is required", nil)
		return true
	}

	if errors.Is(err, services.ErrNoWritePermission) {
		respondForbidden(c, "You don't have write permission for this category")
		return true
	}

	// Log and return generic error
	rid := utils.GetRequestID(c.Request.Context())
	log.Printf("[%s] request=%s user=%v todo=%d error=%v", operation, rid, userID, todoID, err)

	respondInternalError(c, "Failed to "+operation, err)
	return true
}

// CreateTodo handles creating a new todo HTTP request
func (h *TodoHandler) CreateTodo(c *gin.Context) {
	var input CreateTodoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondBadRequest(c, "Validation failed", err)
		return
	}

	// Custom validation for whitespace trimming
	if err := input.Validate(); err != nil {
		respondBadRequest(c, err.Error(), nil)
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		respondUnauthorized(c)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	todo, err := h.todoService.CreateTodo(ctx, dto.CreateTodoRequest{
		Title:       input.Title,
		Description: input.Description,
		Category:    input.Category,
		CategoryID:  input.CategoryID,
		UserID:      userID,
	})

	if h.handleTodoError(c, ctx, err, "create todo", userID, 0) {
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Todo created successfully",
		"data":    todo,
	})
}

// GetTodos retrieves todos for the authenticated user HTTP request
func (h *TodoHandler) GetTodos(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		respondUnauthorized(c)
		return
	}

	// Parse pagination params (service handles validation)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	response, err := h.todoService.GetTodos(ctx, userID, page, pageSize)
	if h.handleTodoError(c, ctx, err, "fetch todos", userID, 0) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "Todos retrieved successfully",
		"data":        response.Todos,
		"count":       len(response.Todos),
		"total":       response.Total,
		"page":        response.Page,
		"page_size":   response.PageSize,
		"total_pages": response.TotalPages,
	})
}

// GetTodo retrieves a single todo by ID HTTP request
func (h *TodoHandler) GetTodo(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		respondBadRequest(c, "Invalid todo ID", nil)
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		respondUnauthorized(c)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	todo, err := h.todoService.GetTodoByID(ctx, dto.GetTodoRequest{
		ID:     id,
		UserID: userID,
	})

	if h.handleTodoError(c, ctx, err, "fetch todo", userID, id) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Todo retrieved successfully",
		"data":    todo,
	})
}

// UpdateTodo handles updating an existing todo HTTP request
func (h *TodoHandler) UpdateTodo(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		respondBadRequest(c, "Invalid todo ID", nil)
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		respondUnauthorized(c)
		return
	}

	var input UpdateTodoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondBadRequest(c, "Validation failed", err)
		return
	}

	// Custom validation for update-specific rules
	if err := input.Validate(); err != nil {
		respondBadRequest(c, err.Error(), nil)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	todo, err := h.todoService.UpdateTodo(ctx, dto.UpdateTodoRequest{
		ID:          id,
		UserID:      userID,
		Title:       input.Title,
		Description: input.Description,
		CategoryID:  input.CategoryID,
		Completed:   input.Completed,
	})

	if h.handleTodoError(c, ctx, err, "update todo", userID, id) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Todo updated successfully",
		"data":    todo,
	})
}

// DeleteTodo handles deleting a todo HTTP request
func (h *TodoHandler) DeleteTodo(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		respondBadRequest(c, "Invalid todo ID", nil)
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		respondUnauthorized(c)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	err = h.todoService.DeleteTodo(ctx, dto.DeleteTodoRequest{
		ID:     id,
		UserID: userID,
	})

	if h.handleTodoError(c, ctx, err, "delete todo", userID, id) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Todo deleted successfully",
	})
}

// GetTodosGroupedByCategory retrieves all accessible todos grouped by category
func (h *TodoHandler) GetTodosGroupedByCategory(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		respondUnauthorized(c)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	response, err := h.todoService.GetTodosGroupedByCategory(ctx, userID)
	if h.handleTodoError(c, ctx, err, "fetch todos by category", userID, 0) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Todos grouped by category retrieved successfully",
		"data":    response.Categories,
	})
}
