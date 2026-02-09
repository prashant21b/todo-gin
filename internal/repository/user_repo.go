package repository

import (
	"context"
	"database/sql"

	"todo-app/db"
	"todo-app/internal/models"
)

// Ensure SQLUserRepository implements UserRepository
var _ UserRepository = (*SQLUserRepository)(nil)

// SQLUserRepository implements UserRepository using sqlc-generated queries
type SQLUserRepository struct {
	queries *db.Queries
}

// NewSQLUserRepository creates a new UserRepository with the provided queries instance
func NewSQLUserRepository(queries *db.Queries) UserRepository {
	return &SQLUserRepository{queries: queries}
}

// toModelUser converts db.User to models.User
func toModelUser(u db.User) models.User {
	return models.User{
		ID:        uint(u.ID),
		Name:      u.Name,
		Email:     u.Email,
		Password:  u.Password,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// CreateUser inserts a new user into the database
func (r *SQLUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	if r.queries == nil {
		return sql.ErrConnDone
	}

	// Insert and get the new ID atomically (no race condition)
	id, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Name:     user.Name,
		Email:    user.Email,
		Password: user.Password,
	})
	if err != nil {
		return err
	}

	// Fetch by exact ID (safe, no race condition)
	u, err := r.queries.GetUserByID(ctx, uint64(id))
	if err != nil {
		return err
	}
	*user = toModelUser(u)
	return nil
}

// GetUserByEmail retrieves a user by their email address
func (r *SQLUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if r.queries == nil {
		return nil, sql.ErrConnDone
	}

	u, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	user := toModelUser(u)
	return &user, nil
}

// GetUserByID retrieves a user by their ID
func (r *SQLUserRepository) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	if r.queries == nil {
		return nil, sql.ErrConnDone
	}

	u, err := r.queries.GetUserByID(ctx, uint64(id))
	if err != nil {
		return nil, err
	}
	user := toModelUser(u)
	return &user, nil
}
