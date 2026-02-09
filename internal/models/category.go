package models

import (
	"time"
)

// Permission represents the permission level for a shared category
type Permission string

const (
	PermissionRead  Permission = "read"
	PermissionWrite Permission = "write"
)

// IsValid checks if the permission is valid
func (p Permission) IsValid() bool {
	return p == PermissionRead || p == PermissionWrite
}

// Category represents a category owned by a user
type Category struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	OwnerID   uint      `json:"owner_id"`
	Todos     []Todo    `json:"todos,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CategoryShare represents a category shared with a user
type CategoryShare struct {
	ID               uint       `json:"id"`
	CategoryID       uint       `json:"category_id"`
	SharedWithUserID uint       `json:"shared_with_user_id"`
	Permission       Permission `json:"permission"`
	CreatedAt        time.Time  `json:"created_at"`
}

// CategoryShareWithUser includes user info for the shared user
type CategoryShareWithUser struct {
	ID                  uint       `json:"id"`
	CategoryID          uint       `json:"category_id"`
	SharedWithUserID    uint       `json:"shared_with_user_id"`
	Permission          Permission `json:"permission"`
	CreatedAt           time.Time  `json:"created_at"`
	SharedWithUserName  string     `json:"shared_with_user_name"`
	SharedWithUserEmail string     `json:"shared_with_user_email"`
}

// SharedCategoryWithOwner includes owner info for a shared category
type SharedCategoryWithOwner struct {
	ID         uint       `json:"id"`
	Name       string     `json:"name"`
	OwnerID    uint       `json:"owner_id"`
	Todos      []Todo     `json:"todos,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	Permission Permission `json:"permission"`
	OwnerName  string     `json:"owner_name"`
	OwnerEmail string     `json:"owner_email"`
}

// CategoryWithTodosRow represents a flat row from the grouped query
// Each row contains one category with one todo (or no todo if category is empty)
type CategoryWithTodosRow struct {
	CategoryID        uint    `json:"category_id"`
	CategoryName      string  `json:"category_name"`
	CategoryOwnerID   uint    `json:"category_owner_id"`
	CategoryOwnerName string  `json:"category_owner_name"`
	UserPermission    string  `json:"user_permission"`
	TodoID            uint    `json:"todo_id"`
	TodoTitle         string  `json:"todo_title"`
	TodoDescription   string  `json:"todo_description"`
	TodoCompleted     bool    `json:"todo_completed"`
	TodoCreatedBy     uint    `json:"todo_created_by"`
	TodoCreatorName   string  `json:"todo_creator_name"`
	TodoCreatedAt     *string `json:"todo_created_at"`
	TodoUpdatedAt     *string `json:"todo_updated_at"`
}
