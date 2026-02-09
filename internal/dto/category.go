package dto

import "todo-app/internal/models"

// CreateCategoryRequest represents the data needed to create a category
type CreateCategoryRequest struct {
	Name    string
	OwnerID uint
}

// UpdateCategoryRequest represents the data needed to update a category
type UpdateCategoryRequest struct {
	ID      uint
	UserID  uint // For ownership verification
	Name    string
}

// ShareCategoryRequest represents the data needed to share a category
type ShareCategoryRequest struct {
	CategoryID      uint
	OwnerID         uint   // User sharing the category (must be owner)
	ShareWithEmail  string // Email of user to share with
	Permission      models.Permission
}

// UnshareCategoryRequest represents the data needed to unshare a category
type UnshareCategoryRequest struct {
	CategoryID       uint
	OwnerID          uint // User unsharing (must be owner)
	SharedWithUserID uint
}

// UpdateSharePermissionRequest represents the data needed to update share permission
type UpdateSharePermissionRequest struct {
	CategoryID       uint
	OwnerID          uint // User updating (must be owner)
	SharedWithUserID uint
	Permission       models.Permission
}

// CategoryListResponse represents a list of categories
type CategoryListResponse struct {
	OwnedCategories  []models.Category             `json:"owned_categories"`
	SharedCategories []models.SharedCategoryWithOwner `json:"shared_categories"`
}
