package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"todo-app/internal/dto"
	"todo-app/internal/models"
	"todo-app/internal/services"
	"todo-app/pkg/utils"

	"github.com/gin-gonic/gin"
)

// CategoryHandler handles HTTP requests for categories
type CategoryHandler struct {
	categoryService services.CategoryService
}

// NewCategoryHandler creates a new CategoryHandler with the provided service
func NewCategoryHandler(svc services.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: svc}
}

// CreateCategoryInput represents the create category request body
type CreateCategoryInput struct {
	Name string `json:"name" binding:"required,min=1,max=255"`
}

// Validate performs custom validation on CreateCategoryInput
func (c *CreateCategoryInput) Validate() error {
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" {
		return errors.New("name cannot be empty or whitespace only")
	}
	return nil
}

// UpdateCategoryInput represents the update category request body
type UpdateCategoryInput struct {
	Name string `json:"name" binding:"required,min=1,max=255"`
}

// Validate performs custom validation on UpdateCategoryInput
func (u *UpdateCategoryInput) Validate() error {
	u.Name = strings.TrimSpace(u.Name)
	if u.Name == "" {
		return errors.New("name cannot be empty or whitespace only")
	}
	return nil
}

// ShareCategoryInput represents the share category request body
type ShareCategoryInput struct {
	Email      string `json:"email" binding:"required,email"`
	Permission string `json:"permission" binding:"required,oneof=read write"`
}

// UpdateSharePermissionInput represents the update share permission request body
type UpdateSharePermissionInput struct {
	Permission string `json:"permission" binding:"required,oneof=read write"`
}

// handleCategoryError maps service errors to HTTP responses
func (h *CategoryHandler) handleCategoryError(c *gin.Context, ctx context.Context, err error, operation string, userID uint, categoryID uint) bool {
	if err == nil {
		return false
	}

	// Check for timeout
	if ctx.Err() != nil {
		respondTimeout(c)
		return true
	}

	// Handle specific business errors
	if errors.Is(err, services.ErrCategoryNotFound) {
		respondNotFound(c, "Category")
		return true
	}

	if errors.Is(err, services.ErrCategoryForbidden) {
		respondForbidden(c, "You don't have permission to access this category")
		return true
	}

	if errors.Is(err, services.ErrCategoryNameExists) {
		respondConflict(c, "Category with this name already exists")
		return true
	}

	if errors.Is(err, services.ErrUserNotFound) {
		respondNotFound(c, "User")
		return true
	}

	if errors.Is(err, services.ErrCannotShareWithSelf) {
		respondBadRequest(c, "Cannot share category with yourself", nil)
		return true
	}

	if errors.Is(err, services.ErrShareAlreadyExists) {
		respondConflict(c, "Category is already shared with this user")
		return true
	}

	if errors.Is(err, services.ErrShareNotFound) {
		respondNotFound(c, "Share")
		return true
	}

	// Log and return generic error
	rid := utils.GetRequestID(c.Request.Context())
	log.Printf("[%s] request=%s user=%v category=%d error=%v", operation, rid, userID, categoryID, err)

	respondInternalError(c, "Failed to "+operation, err)
	return true
}

// CreateCategory handles creating a new category HTTP request
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var input CreateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondBadRequest(c, "Validation failed", err)
		return
	}

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

	category, err := h.categoryService.CreateCategory(ctx, dto.CreateCategoryRequest{
		Name:    input.Name,
		OwnerID: userID,
	})

	if h.handleCategoryError(c, ctx, err, "create category", userID, 0) {
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Category created successfully",
		"data":    category,
	})
}

// GetCategories retrieves all categories for the authenticated user
func (h *CategoryHandler) GetCategories(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		respondUnauthorized(c)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Get owned categories
	ownedCategories, err := h.categoryService.GetCategories(ctx, userID)
	if h.handleCategoryError(c, ctx, err, "fetch categories", userID, 0) {
		return
	}

	// Get shared categories
	sharedCategories, err := h.categoryService.GetSharedCategories(ctx, userID)
	if h.handleCategoryError(c, ctx, err, "fetch shared categories", userID, 0) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Categories retrieved successfully",
		"data": dto.CategoryListResponse{
			OwnedCategories:  ownedCategories,
			SharedCategories: sharedCategories,
		},
	})
}

// GetCategory retrieves a single category by ID
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		respondBadRequest(c, "Invalid category ID", nil)
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		respondUnauthorized(c)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	category, err := h.categoryService.GetCategoryByID(ctx, id, userID)
	if h.handleCategoryError(c, ctx, err, "fetch category", userID, id) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Category retrieved successfully",
		"data":    category,
	})
}

// UpdateCategory handles updating an existing category
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		respondBadRequest(c, "Invalid category ID", nil)
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		respondUnauthorized(c)
		return
	}

	var input UpdateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondBadRequest(c, "Validation failed", err)
		return
	}

	if err := input.Validate(); err != nil {
		respondBadRequest(c, err.Error(), nil)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	category, err := h.categoryService.UpdateCategory(ctx, dto.UpdateCategoryRequest{
		ID:     id,
		UserID: userID,
		Name:   input.Name,
	})

	if h.handleCategoryError(c, ctx, err, "update category", userID, id) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Category updated successfully",
		"data":    category,
	})
}

// DeleteCategory handles deleting a category
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		respondBadRequest(c, "Invalid category ID", nil)
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		respondUnauthorized(c)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	err = h.categoryService.DeleteCategory(ctx, id, userID)
	if h.handleCategoryError(c, ctx, err, "delete category", userID, id) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Category deleted successfully",
	})
}

// ShareCategory handles sharing a category with another user
func (h *CategoryHandler) ShareCategory(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		respondBadRequest(c, "Invalid category ID", nil)
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		respondUnauthorized(c)
		return
	}

	var input ShareCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondBadRequest(c, "Validation failed", err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	share, err := h.categoryService.ShareCategory(ctx, dto.ShareCategoryRequest{
		CategoryID:     id,
		OwnerID:        userID,
		ShareWithEmail: input.Email,
		Permission:     models.Permission(input.Permission),
	})

	if h.handleCategoryError(c, ctx, err, "share category", userID, id) {
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Category shared successfully",
		"data":    share,
	})
}

// UnshareCategory handles removing sharing of a category
func (h *CategoryHandler) UnshareCategory(c *gin.Context) {
	categoryID, err := parseIDParam(c, "id")
	if err != nil {
		respondBadRequest(c, "Invalid category ID", nil)
		return
	}

	shareUserID, err := parseIDParam(c, "user_id")
	if err != nil {
		respondBadRequest(c, "Invalid user ID", nil)
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		respondUnauthorized(c)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	err = h.categoryService.UnshareCategory(ctx, dto.UnshareCategoryRequest{
		CategoryID:       categoryID,
		OwnerID:          userID,
		SharedWithUserID: shareUserID,
	})

	if h.handleCategoryError(c, ctx, err, "unshare category", userID, categoryID) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Category share removed successfully",
	})
}

// UpdateSharePermission handles updating the permission of a share
func (h *CategoryHandler) UpdateSharePermission(c *gin.Context) {
	categoryID, err := parseIDParam(c, "id")
	if err != nil {
		respondBadRequest(c, "Invalid category ID", nil)
		return
	}

	shareUserID, err := parseIDParam(c, "user_id")
	if err != nil {
		respondBadRequest(c, "Invalid user ID", nil)
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		respondUnauthorized(c)
		return
	}

	var input UpdateSharePermissionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondBadRequest(c, "Validation failed", err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	err = h.categoryService.UpdateSharePermission(ctx, dto.UpdateSharePermissionRequest{
		CategoryID:       categoryID,
		OwnerID:          userID,
		SharedWithUserID: shareUserID,
		Permission:       models.Permission(input.Permission),
	})

	if h.handleCategoryError(c, ctx, err, "update share permission", userID, categoryID) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Share permission updated successfully",
	})
}

// GetShares retrieves all shares for a category
func (h *CategoryHandler) GetShares(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		respondBadRequest(c, "Invalid category ID", nil)
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		respondUnauthorized(c)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	shares, err := h.categoryService.GetSharesForCategory(ctx, id, userID)
	if h.handleCategoryError(c, ctx, err, "fetch shares", userID, id) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Shares retrieved successfully",
		"data":    shares,
		"count":   len(shares),
	})
}
