package repository

import (
	"context"
	"database/sql"

	"todo-app/db"
	"todo-app/internal/models"
)

// Ensure SQLCategoryRepository implements CategoryRepository
var _ CategoryRepository = (*SQLCategoryRepository)(nil)

// SQLCategoryRepository implements CategoryRepository using sqlc-generated queries
type SQLCategoryRepository struct {
	queries *db.Queries
}

// NewSQLCategoryRepository creates a new CategoryRepository with the provided queries instance
func NewSQLCategoryRepository(queries *db.Queries) CategoryRepository {
	return &SQLCategoryRepository{queries: queries}
}

// toModelCategory converts db.Category to models.Category
func toModelCategory(c db.Category) models.Category {
	return models.Category{
		ID:        uint(c.ID),
		Name:      c.Name,
		OwnerID:   uint(c.OwnerID),
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// CreateCategory inserts a new category into the database
func (r *SQLCategoryRepository) CreateCategory(ctx context.Context, category *models.Category) error {
	if r.queries == nil {
		return sql.ErrConnDone
	}

	id, err := r.queries.CreateCategory(ctx, db.CreateCategoryParams{
		Name:    category.Name,
		OwnerID: uint64(category.OwnerID),
	})
	if err != nil {
		return err
	}

	// Fetch the created category
	created, err := r.queries.GetCategoryByID(ctx, uint64(id))
	if err != nil {
		return err
	}
	*category = toModelCategory(created)
	return nil
}

// GetCategoryByID retrieves a single category by its ID
func (r *SQLCategoryRepository) GetCategoryByID(ctx context.Context, id uint) (*models.Category, error) {
	if r.queries == nil {
		return nil, sql.ErrConnDone
	}

	c, err := r.queries.GetCategoryByID(ctx, uint64(id))
	if err != nil {
		return nil, err
	}
	category := toModelCategory(c)
	return &category, nil
}

// GetCategoriesByOwnerID retrieves all categories for a specific owner
func (r *SQLCategoryRepository) GetCategoriesByOwnerID(ctx context.Context, ownerID uint) ([]models.Category, error) {
	if r.queries == nil {
		return nil, sql.ErrConnDone
	}

	items, err := r.queries.GetCategoriesByOwnerID(ctx, uint64(ownerID))
	if err != nil {
		return nil, err
	}

	categories := make([]models.Category, 0, len(items))
	for _, item := range items {
		categories = append(categories, toModelCategory(item))
	}
	return categories, nil
}

// GetCategoryByNameAndOwner retrieves a category by name and owner ID
func (r *SQLCategoryRepository) GetCategoryByNameAndOwner(ctx context.Context, ownerID uint, name string) (*models.Category, error) {
	if r.queries == nil {
		return nil, sql.ErrConnDone
	}

	c, err := r.queries.GetCategoryByNameAndOwner(ctx, db.GetCategoryByNameAndOwnerParams{
		OwnerID: uint64(ownerID),
		Name:    name,
	})
	if err != nil {
		return nil, err
	}
	category := toModelCategory(c)
	return &category, nil
}

// UpdateCategory updates an existing category
func (r *SQLCategoryRepository) UpdateCategory(ctx context.Context, category *models.Category) error {
	if r.queries == nil {
		return sql.ErrConnDone
	}

	err := r.queries.UpdateCategory(ctx, db.UpdateCategoryParams{
		Name: category.Name,
		ID:   uint64(category.ID),
	})
	if err != nil {
		return err
	}

	// Fetch updated record
	updated, err := r.queries.GetCategoryByID(ctx, uint64(category.ID))
	if err != nil {
		return err
	}
	*category = toModelCategory(updated)
	return nil
}

// DeleteCategory deletes a category from the database
func (r *SQLCategoryRepository) DeleteCategory(ctx context.Context, id uint) error {
	if r.queries == nil {
		return sql.ErrConnDone
	}
	return r.queries.DeleteCategory(ctx, uint64(id))
}
