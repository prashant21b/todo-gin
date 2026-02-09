package models

import (
	"time"
)

// Todo represents the todo model (pure data structure)
type Todo struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	CategoryID  uint       `json:"category_id"`
	Completed   bool       `json:"completed"`
	UserID      uint       `json:"user_id"`
	CreatedBy   uint       `json:"created_by"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
