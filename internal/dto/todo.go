package dto

import "todo-app/internal/models"

// CreateTodoRequest represents the data needed to create a todo
type CreateTodoRequest struct {
	Title       string
	Description string
	Category    string // Category name (used only when CategoryID is not set; will be created if doesn't exist)
	CategoryID  *uint  // Optional: use this category when set (user must have write access)
	UserID      uint   // User creating the todo
}

// UpdateTodoRequest represents the data needed to update a todo
type UpdateTodoRequest struct {
	ID          uint
	UserID      uint // For permission verification
	Title       *string
	Description *string
	CategoryID  *uint
	Completed   *bool
}

// GetTodoRequest represents the data needed to get a single todo
type GetTodoRequest struct {
	ID     uint
	UserID uint // For permission verification
}

// DeleteTodoRequest represents the data needed to delete a todo
type DeleteTodoRequest struct {
	ID     uint
	UserID uint // For permission verification
}

// TodoListResponse represents paginated todo list response
type TodoListResponse struct {
	Todos      []models.Todo
	Total      int64
	Page       int
	PageSize   int
	TotalPages int64
}

// TodoInCategory represents a todo item within a category
type TodoInCategory struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	CreatedBy   uint   `json:"created_by"`
	CreatorName string `json:"creator_name"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// CategoryWithTodos represents a category and all its todos
type CategoryWithTodos struct {
	ID             uint             `json:"id"`
	Name           string           `json:"name"`
	OwnerID        uint             `json:"owner_id"`
	OwnerName      string           `json:"owner_name"`
	UserPermission string           `json:"user_permission"` // "owner", "read", or "write"
	Todos          []TodoInCategory `json:"todos"`
}

// TodosGroupedByCategoryResponse represents the full grouped response
type TodosGroupedByCategoryResponse struct {
	Categories []CategoryWithTodos `json:"categories"`
}
