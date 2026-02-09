package repository

import (
	"context"
	"database/sql"
	"time"

	"todo-app/db"
	"todo-app/internal/models"
)

// Ensure SQLTodoRepository implements TodoRepository
var _ TodoRepository = (*SQLTodoRepository)(nil)

// SQLTodoRepository implements TodoRepository using sqlc-generated queries
type SQLTodoRepository struct {
	queries *db.Queries
}

// NewSQLTodoRepository creates a new TodoRepository with the provided queries instance
func NewSQLTodoRepository(queries *db.Queries) TodoRepository {
	return &SQLTodoRepository{queries: queries}
}

// toModelTodo converts db.Todo to models.Todo
func toModelTodo(t db.Todo) models.Todo {
	d := ""
	if t.Description.Valid {
		d = t.Description.String
	}
	var deletedAt *time.Time
	if t.DeletedAt.Valid {
		deletedAt = &t.DeletedAt.Time
	}
	return models.Todo{
		ID:          uint(t.ID),
		Title:       t.Title,
		Description: d,
		CategoryID:  uint(t.CategoryID),
		Completed:   t.Completed,
		UserID:      uint(t.UserID),
		CreatedBy:   uint(t.CreatedBy),
		DeletedAt:   deletedAt,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// CreateTodo inserts a new todo into the database
func (r *SQLTodoRepository) CreateTodo(ctx context.Context, todo *models.Todo) error {
	if r.queries == nil {
		return sql.ErrConnDone
	}

	// Insert and get the new ID atomically (no race condition)
	id, err := r.queries.CreateTodo(ctx, db.CreateTodoParams{
		Title:       todo.Title,
		Description: sql.NullString{String: todo.Description, Valid: todo.Description != ""},
		CategoryID:  uint64(todo.CategoryID),
		Completed:   todo.Completed,
		UserID:      uint64(todo.UserID),
		CreatedBy:   uint64(todo.CreatedBy),
	})
	if err != nil {
		return err
	}

	// Fetch by exact ID (safe, no race condition)
	created, err := r.queries.GetTodoByID(ctx, uint64(id))
	if err != nil {
		return err
	}
	*todo = toModelTodo(created)
	return nil
}

// GetTodos retrieves todos created by the specific user with pagination
func (r *SQLTodoRepository) GetTodos(ctx context.Context, userID uint, page, pageSize int) ([]models.Todo, int64, error) {
	if r.queries == nil {
		return nil, 0, sql.ErrConnDone
	}

	// Count total todos owned/created by the user
	total, err := r.queries.CountTodosByUserID(ctx, uint64(userID))
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []models.Todo{}, total, nil
	}

	// Calculate offset
	offset := int32((page - 1) * pageSize)
	limit := int32(pageSize)

	// Get todos where user_id == userID
	items, err := r.queries.GetTodosByUserIDWithPagination(ctx, db.GetTodosByUserIDWithPaginationParams{
		UserID: uint64(userID),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, err
	}

	todos := make([]models.Todo, 0, len(items))
	for _, it := range items {
		todos = append(todos, toModelTodo(it))
	}
	return todos, total, nil
}

// GetTodosByCategoryID retrieves todos for a specific category with pagination
func (r *SQLTodoRepository) GetTodosByCategoryID(ctx context.Context, categoryID uint, page, pageSize int) ([]models.Todo, int64, error) {
	if r.queries == nil {
		return nil, 0, sql.ErrConnDone
	}

	// Count total matching records
	total, err := r.queries.CountTodosByCategoryID(ctx, uint64(categoryID))
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []models.Todo{}, total, nil
	}

	// Calculate offset
	offset := int32((page - 1) * pageSize)
	limit := int32(pageSize)

	items, err := r.queries.GetTodosByCategoryID(ctx, db.GetTodosByCategoryIDParams{
		CategoryID: uint64(categoryID),
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, 0, err
	}

	todos := make([]models.Todo, 0, len(items))
	for _, it := range items {
		todos = append(todos, toModelTodo(it))
	}
	return todos, total, nil
}

// GetTodoByID retrieves a single todo by its ID
func (r *SQLTodoRepository) GetTodoByID(ctx context.Context, id uint) (*models.Todo, error) {
	if r.queries == nil {
		return nil, sql.ErrConnDone
	}

	t, err := r.queries.GetTodoByID(ctx, uint64(id))
	if err != nil {
		return nil, err
	}
	todo := toModelTodo(t)
	return &todo, nil
}

// UpdateTodo updates an existing todo
func (r *SQLTodoRepository) UpdateTodo(ctx context.Context, todo *models.Todo) error {
	if r.queries == nil {
		return sql.ErrConnDone
	}

	err := r.queries.UpdateTodo(ctx, db.UpdateTodoParams{
		Title:       todo.Title,
		Description: sql.NullString{String: todo.Description, Valid: todo.Description != ""},
		CategoryID:  uint64(todo.CategoryID),
		Completed:   todo.Completed,
		ID:          uint64(todo.ID),
	})
	if err != nil {
		return err
	}

	// Fetch updated record
	updated, err := r.queries.GetTodoByID(ctx, uint64(todo.ID))
	if err != nil {
		return err
	}
	*todo = toModelTodo(updated)
	return nil
}

// DeleteTodo soft deletes a todo from the database
func (r *SQLTodoRepository) DeleteTodo(ctx context.Context, id uint) error {
	if r.queries == nil {
		return sql.ErrConnDone
	}
	return r.queries.SoftDeleteTodo(ctx, uint64(id))
}
