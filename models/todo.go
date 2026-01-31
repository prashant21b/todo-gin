package models

import (
	"context"
	"time"

	"todo-app/config"
)

// Todo represents the todo model
type Todo struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"size:255;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Completed   bool      `json:"completed" gorm:"default:false"`
	UserID      uint      `json:"user_id" gorm:"not null;index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateTodo inserts a new todo into the database with context support
func CreateTodo(ctx context.Context, todo *Todo) error {
	return config.GetDBWithContext(ctx).Create(todo).Error
}

// GetTodosByUserID retrieves all todos for a specific user with context support
func GetTodosByUserID(ctx context.Context, userID uint) ([]Todo, error) {
	var todos []Todo
	err := config.GetDBWithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&todos).Error
	return todos, err
}

// GetTodoByID retrieves a single todo by its ID with context support
func GetTodoByID(ctx context.Context, id uint) (*Todo, error) {
	var todo Todo
	err := config.GetDBWithContext(ctx).First(&todo, id).Error
	if err != nil {
		return nil, err
	}
	return &todo, nil
}

// UpdateTodo updates an existing todo with context support
func UpdateTodo(ctx context.Context, todo *Todo) error {
	return config.GetDBWithContext(ctx).Save(todo).Error
}

// DeleteTodo removes a todo from the database with context support
func DeleteTodo(ctx context.Context, id uint) error {
	return config.GetDBWithContext(ctx).Delete(&Todo{}, id).Error
}
