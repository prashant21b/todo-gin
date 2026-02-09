package repository

import (
	"context"
	"database/sql"

	"todo-app/db"
	"todo-app/internal/models"
)

// Ensure SQLCategoryShareRepository implements CategoryShareRepository
var _ CategoryShareRepository = (*SQLCategoryShareRepository)(nil)

// SQLCategoryShareRepository implements CategoryShareRepository using sqlc-generated queries
type SQLCategoryShareRepository struct {
	queries *db.Queries
}

// NewSQLCategoryShareRepository creates a new CategoryShareRepository with the provided queries instance
func NewSQLCategoryShareRepository(queries *db.Queries) CategoryShareRepository {
	return &SQLCategoryShareRepository{queries: queries}
}

// toModelCategoryShare converts db.CategoryShare to models.CategoryShare
func toModelCategoryShare(cs db.CategoryShare) models.CategoryShare {
	return models.CategoryShare{
		ID:               uint(cs.ID),
		CategoryID:       uint(cs.CategoryID),
		SharedWithUserID: uint(cs.SharedWithUserID),
		Permission:       models.Permission(cs.Permission),
		CreatedAt:        cs.CreatedAt,
	}
}

// CreateCategoryShare inserts a new category share into the database
func (r *SQLCategoryShareRepository) CreateCategoryShare(ctx context.Context, share *models.CategoryShare) error {
	if r.queries == nil {
		return sql.ErrConnDone
	}

	id, err := r.queries.CreateCategoryShare(ctx, db.CreateCategoryShareParams{
		CategoryID:       uint64(share.CategoryID),
		SharedWithUserID: uint64(share.SharedWithUserID),
		Permission:       db.CategorySharesPermission(share.Permission),
	})
	if err != nil {
		return err
	}

	// Fetch the created share
	created, err := r.queries.GetCategoryShareByID(ctx, uint64(id))
	if err != nil {
		return err
	}
	*share = toModelCategoryShare(created)
	return nil
}

// GetCategoryShareByID retrieves a single category share by its ID
func (r *SQLCategoryShareRepository) GetCategoryShareByID(ctx context.Context, id uint) (*models.CategoryShare, error) {
	if r.queries == nil {
		return nil, sql.ErrConnDone
	}

	cs, err := r.queries.GetCategoryShareByID(ctx, uint64(id))
	if err != nil {
		return nil, err
	}
	share := toModelCategoryShare(cs)
	return &share, nil
}

// GetCategoryShareByCategoryAndUser retrieves a share by category and user
func (r *SQLCategoryShareRepository) GetCategoryShareByCategoryAndUser(ctx context.Context, categoryID, userID uint) (*models.CategoryShare, error) {
	if r.queries == nil {
		return nil, sql.ErrConnDone
	}

	cs, err := r.queries.GetCategoryShareByCategoryAndUser(ctx, db.GetCategoryShareByCategoryAndUserParams{
		CategoryID:       uint64(categoryID),
		SharedWithUserID: uint64(userID),
	})
	if err != nil {
		return nil, err
	}
	share := toModelCategoryShare(cs)
	return &share, nil
}

// GetSharesForCategory retrieves all shares for a category with user details
func (r *SQLCategoryShareRepository) GetSharesForCategory(ctx context.Context, categoryID uint) ([]models.CategoryShareWithUser, error) {
	if r.queries == nil {
		return nil, sql.ErrConnDone
	}

	items, err := r.queries.GetSharesForCategory(ctx, uint64(categoryID))
	if err != nil {
		return nil, err
	}

	shares := make([]models.CategoryShareWithUser, 0, len(items))
	for _, item := range items {
		shares = append(shares, models.CategoryShareWithUser{
			ID:                  uint(item.ID),
			CategoryID:          uint(item.CategoryID),
			SharedWithUserID:    uint(item.SharedWithUserID),
			Permission:          models.Permission(item.Permission),
			CreatedAt:           item.CreatedAt,
			SharedWithUserName:  item.SharedWithUserName,
			SharedWithUserEmail: item.SharedWithUserEmail,
		})
	}
	return shares, nil
}

// GetSharedCategoriesForUser retrieves all categories shared with a user
func (r *SQLCategoryShareRepository) GetSharedCategoriesForUser(ctx context.Context, userID uint) ([]models.SharedCategoryWithOwner, error) {
	if r.queries == nil {
		return nil, sql.ErrConnDone
	}

	items, err := r.queries.GetSharedCategoriesForUser(ctx, uint64(userID))
	if err != nil {
		return nil, err
	}

	categories := make([]models.SharedCategoryWithOwner, 0, len(items))
	for _, item := range items {
		categories = append(categories, models.SharedCategoryWithOwner{
			ID:         uint(item.ID),
			Name:       item.Name,
			OwnerID:    uint(item.OwnerID),
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  item.UpdatedAt,
			Permission: models.Permission(item.Permission),
			OwnerName:  item.OwnerName,
			OwnerEmail: item.OwnerEmail,
		})
	}
	return categories, nil
}

// UpdateCategorySharePermission updates the permission for a share
func (r *SQLCategoryShareRepository) UpdateCategorySharePermission(ctx context.Context, id uint, permission models.Permission) error {
	if r.queries == nil {
		return sql.ErrConnDone
	}

	return r.queries.UpdateCategorySharePermission(ctx, db.UpdateCategorySharePermissionParams{
		ID:         uint64(id),
		Permission: db.CategorySharesPermission(permission),
	})
}

// DeleteCategoryShare deletes a category share by ID
func (r *SQLCategoryShareRepository) DeleteCategoryShare(ctx context.Context, id uint) error {
	if r.queries == nil {
		return sql.ErrConnDone
	}
	return r.queries.DeleteCategoryShare(ctx, uint64(id))
}

// DeleteCategoryShareByUserAndCategory deletes a share by category and user
func (r *SQLCategoryShareRepository) DeleteCategoryShareByUserAndCategory(ctx context.Context, categoryID, userID uint) error {
	if r.queries == nil {
		return sql.ErrConnDone
	}
	return r.queries.DeleteCategoryShareByUserAndCategory(ctx, db.DeleteCategoryShareByUserAndCategoryParams{
		CategoryID:       uint64(categoryID),
		SharedWithUserID: uint64(userID),
	})
}

// GetUserPermissionForCategory gets the user's permission for a category
func (r *SQLCategoryShareRepository) GetUserPermissionForCategory(ctx context.Context, userID, categoryID uint) (string, error) {
	if r.queries == nil {
		return "", sql.ErrConnDone
	}

	result, err := r.queries.GetUserPermissionForCategory(ctx, db.GetUserPermissionForCategoryParams{
		OwnerID:          uint64(userID),
		SharedWithUserID: uint64(userID),
		ID:               uint64(categoryID),
	})
	if err != nil {
		return "", err
	}

	// Convert interface{} to string
	if permission, ok := result.(string); ok {
		return permission, nil
	}
	if permission, ok := result.([]byte); ok {
		return string(permission), nil
	}
	return "none", nil
}

// GetTodosGroupedByCategory retrieves all todos grouped by categories accessible to the user
func (r *SQLCategoryShareRepository) GetTodosGroupedByCategory(ctx context.Context, userID uint) ([]models.CategoryWithTodosRow, error) {
	if r.queries == nil {
		return nil, sql.ErrConnDone
	}

	items, err := r.queries.GetTodosGroupedByCategory(ctx, db.GetTodosGroupedByCategoryParams{
		OwnerID:            uint64(userID),
		SharedWithUserID:   uint64(userID),
		OwnerID_2:          uint64(userID),
		SharedWithUserID_2: uint64(userID),
	})
	if err != nil {
		return nil, err
	}

	rows := make([]models.CategoryWithTodosRow, 0, len(items))
	for _, item := range items {
		ownerName := ""
		if item.CategoryOwnerName.Valid {
			ownerName = item.CategoryOwnerName.String
		}

		permission := ""
		if item.UserPermission != nil {
			if p, ok := item.UserPermission.(string); ok {
				permission = p
			} else if p, ok := item.UserPermission.([]byte); ok {
				permission = string(p)
			}
		}

		var createdAt, updatedAt *string
		if item.TodoCreatedAt.Valid {
			t := item.TodoCreatedAt.Time.Format("2006-01-02T15:04:05Z")
			createdAt = &t
		}
		if item.TodoUpdatedAt.Valid {
			t := item.TodoUpdatedAt.Time.Format("2006-01-02T15:04:05Z")
			updatedAt = &t
		}

		rows = append(rows, models.CategoryWithTodosRow{
			CategoryID:        uint(item.CategoryID),
			CategoryName:      item.CategoryName,
			CategoryOwnerID:   uint(item.CategoryOwnerID),
			CategoryOwnerName: ownerName,
			UserPermission:    permission,
			TodoID:            uint(item.TodoID),
			TodoTitle:         item.TodoTitle,
			TodoDescription:   item.TodoDescription,
			TodoCompleted:     item.TodoCompleted,
			TodoCreatedBy:     uint(item.TodoCreatedBy),
			TodoCreatorName:   item.TodoCreatorName,
			TodoCreatedAt:     createdAt,
			TodoUpdatedAt:     updatedAt,
		})
	}
	return rows, nil
}
