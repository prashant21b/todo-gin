package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"todo-app/internal/dto"
	"todo-app/internal/models"
	"todo-app/internal/repository/mocks"
)

func createTestCategoryService(
	categoryRepo *mocks.MockCategoryRepository,
	categoryShareRepo *mocks.MockCategoryShareRepository,
	userRepo *mocks.MockUserRepository,
) CategoryService {
	if categoryRepo == nil {
		categoryRepo = &mocks.MockCategoryRepository{}
	}
	if categoryShareRepo == nil {
		categoryShareRepo = &mocks.MockCategoryShareRepository{}
	}
	if userRepo == nil {
		userRepo = &mocks.MockUserRepository{}
	}
	// Provide a default mock todo repo so service can fetch todos for categories
	todoRepo := &mocks.MockTodoRepository{}
	return NewCategoryService(categoryRepo, categoryShareRepo, userRepo, todoRepo)
}

func TestCategoryService_CreateCategory(t *testing.T) {
	tests := []struct {
		name       string
		req        dto.CreateCategoryRequest
		existsErr  error
		createErr  error
		wantErr    bool
		expectedID uint
	}{
		{
			name:       "successful creation",
			req:        dto.CreateCategoryRequest{Name: "Work", OwnerID: 1},
			existsErr:  sql.ErrNoRows,
			createErr:  nil,
			wantErr:    false,
			expectedID: 1,
		},
		{
			name:      "category name already exists",
			req:       dto.CreateCategoryRequest{Name: "Work", OwnerID: 1},
			existsErr: nil, // no error means found
			wantErr:   true,
		},
		{
			name:      "database error on create",
			req:       dto.CreateCategoryRequest{Name: "Work", OwnerID: 1},
			existsErr: sql.ErrNoRows,
			createErr: errors.New("database error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			categoryRepo := &mocks.MockCategoryRepository{
				GetCategoryByNameAndOwnerFunc: func(ctx context.Context, ownerID uint, name string) (*models.Category, error) {
					if tt.existsErr != nil {
						return nil, tt.existsErr
					}
					return &models.Category{ID: 1, Name: name, OwnerID: ownerID}, nil
				},
				CreateCategoryFunc: func(ctx context.Context, category *models.Category) error {
					if tt.createErr != nil {
						return tt.createErr
					}
					category.ID = 1
					return nil
				},
			}

			service := createTestCategoryService(categoryRepo, nil, nil)
			cat, err := service.CreateCategory(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCategory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && cat.ID != tt.expectedID {
				t.Errorf("CreateCategory() ID = %v, want %v", cat.ID, tt.expectedID)
			}
		})
	}
}

func TestCategoryService_GetCategoryByID(t *testing.T) {
	tests := []struct {
		name       string
		categoryID uint
		userID     uint
		category   *models.Category
		getErr     error
		permission string
		wantErr    bool
	}{
		{
			name:       "owner can access",
			categoryID: 1,
			userID:     1,
			category:   &models.Category{ID: 1, Name: "Work", OwnerID: 1},
			wantErr:    false,
		},
		{
			name:       "shared user can access",
			categoryID: 1,
			userID:     2,
			category:   &models.Category{ID: 1, Name: "Work", OwnerID: 1},
			permission: "read",
			wantErr:    false,
		},
		{
			name:       "non-shared user cannot access",
			categoryID: 1,
			userID:     3,
			category:   &models.Category{ID: 1, Name: "Work", OwnerID: 1},
			permission: "none",
			wantErr:    true,
		},
		{
			name:       "category not found",
			categoryID: 999,
			userID:     1,
			getErr:     sql.ErrNoRows,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			categoryRepo := &mocks.MockCategoryRepository{
				GetCategoryByIDFunc: func(ctx context.Context, id uint) (*models.Category, error) {
					if tt.getErr != nil {
						return nil, tt.getErr
					}
					return tt.category, nil
				},
			}

			categoryShareRepo := &mocks.MockCategoryShareRepository{
				GetUserPermissionForCategoryFunc: func(ctx context.Context, userID, categoryID uint) (string, error) {
					return tt.permission, nil
				},
			}

			service := createTestCategoryService(categoryRepo, categoryShareRepo, nil)
			cat, err := service.GetCategoryByID(context.Background(), tt.categoryID, tt.userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetCategoryByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && cat.ID != tt.categoryID {
				t.Errorf("GetCategoryByID() ID = %v, want %v", cat.ID, tt.categoryID)
			}
		})
	}
}

func TestCategoryService_UpdateCategory(t *testing.T) {
	tests := []struct {
		name      string
		req       dto.UpdateCategoryRequest
		existing  *models.Category
		getErr    error
		updateErr error
		wantErr   bool
	}{
		{
			name:     "successful update",
			req:      dto.UpdateCategoryRequest{ID: 1, UserID: 1, Name: "Updated"},
			existing: &models.Category{ID: 1, Name: "Original", OwnerID: 1},
			wantErr:  false,
		},
		{
			name:     "not owner - forbidden",
			req:      dto.UpdateCategoryRequest{ID: 1, UserID: 2, Name: "Updated"},
			existing: &models.Category{ID: 1, Name: "Original", OwnerID: 1},
			wantErr:  true,
		},
		{
			name:    "category not found",
			req:     dto.UpdateCategoryRequest{ID: 999, UserID: 1, Name: "Updated"},
			getErr:  sql.ErrNoRows,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			categoryRepo := &mocks.MockCategoryRepository{
				GetCategoryByIDFunc: func(ctx context.Context, id uint) (*models.Category, error) {
					if tt.getErr != nil {
						return nil, tt.getErr
					}
					return tt.existing, nil
				},
				GetCategoryByNameAndOwnerFunc: func(ctx context.Context, ownerID uint, name string) (*models.Category, error) {
					return nil, sql.ErrNoRows
				},
				UpdateCategoryFunc: func(ctx context.Context, category *models.Category) error {
					return tt.updateErr
				},
			}

			service := createTestCategoryService(categoryRepo, nil, nil)
			cat, err := service.UpdateCategory(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateCategory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && cat.Name != tt.req.Name {
				t.Errorf("UpdateCategory() Name = %v, want %v", cat.Name, tt.req.Name)
			}
		})
	}
}

func TestCategoryService_DeleteCategory(t *testing.T) {
	tests := []struct {
		name       string
		categoryID uint
		userID     uint
		existing   *models.Category
		getErr     error
		deleteErr  error
		wantErr    bool
	}{
		{
			name:       "successful delete",
			categoryID: 1,
			userID:     1,
			existing:   &models.Category{ID: 1, Name: "Work", OwnerID: 1},
			wantErr:    false,
		},
		{
			name:       "not owner - forbidden",
			categoryID: 1,
			userID:     2,
			existing:   &models.Category{ID: 1, Name: "Work", OwnerID: 1},
			wantErr:    true,
		},
		{
			name:       "category not found",
			categoryID: 999,
			userID:     1,
			getErr:     sql.ErrNoRows,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			categoryRepo := &mocks.MockCategoryRepository{
				GetCategoryByIDFunc: func(ctx context.Context, id uint) (*models.Category, error) {
					if tt.getErr != nil {
						return nil, tt.getErr
					}
					return tt.existing, nil
				},
				DeleteCategoryFunc: func(ctx context.Context, id uint) error {
					return tt.deleteErr
				},
			}

			service := createTestCategoryService(categoryRepo, nil, nil)
			err := service.DeleteCategory(context.Background(), tt.categoryID, tt.userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteCategory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCategoryService_ShareCategory(t *testing.T) {
	tests := []struct {
		name            string
		req             dto.ShareCategoryRequest
		category        *models.Category
		shareWithUser   *models.User
		existingShare   *models.CategoryShare
		getCategoryErr  error
		getUserErr      error
		getShareErr     error
		createShareErr  error
		wantErr         bool
		expectedErrType error
	}{
		{
			name:          "successful share",
			req:           dto.ShareCategoryRequest{CategoryID: 1, OwnerID: 1, ShareWithEmail: "user2@test.com", Permission: "read"},
			category:      &models.Category{ID: 1, Name: "Work", OwnerID: 1},
			shareWithUser: &models.User{ID: 2, Email: "user2@test.com"},
			getShareErr:   sql.ErrNoRows,
			wantErr:       false,
		},
		{
			name:           "category not found",
			req:            dto.ShareCategoryRequest{CategoryID: 999, OwnerID: 1, ShareWithEmail: "user2@test.com", Permission: "read"},
			getCategoryErr: sql.ErrNoRows,
			wantErr:        true,
		},
		{
			name:       "user to share with not found",
			req:        dto.ShareCategoryRequest{CategoryID: 1, OwnerID: 1, ShareWithEmail: "unknown@test.com", Permission: "read"},
			category:   &models.Category{ID: 1, Name: "Work", OwnerID: 1},
			getUserErr: sql.ErrNoRows,
			wantErr:    true,
		},
		{
			name:          "cannot share with self",
			req:           dto.ShareCategoryRequest{CategoryID: 1, OwnerID: 1, ShareWithEmail: "owner@test.com", Permission: "read"},
			category:      &models.Category{ID: 1, Name: "Work", OwnerID: 1},
			shareWithUser: &models.User{ID: 1, Email: "owner@test.com"}, // same as owner
			wantErr:       true,
		},
		{
			name:          "share already exists",
			req:           dto.ShareCategoryRequest{CategoryID: 1, OwnerID: 1, ShareWithEmail: "user2@test.com", Permission: "read"},
			category:      &models.Category{ID: 1, Name: "Work", OwnerID: 1},
			shareWithUser: &models.User{ID: 2, Email: "user2@test.com"},
			existingShare: &models.CategoryShare{ID: 1, CategoryID: 1, SharedWithUserID: 2},
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			categoryRepo := &mocks.MockCategoryRepository{
				GetCategoryByIDFunc: func(ctx context.Context, id uint) (*models.Category, error) {
					if tt.getCategoryErr != nil {
						return nil, tt.getCategoryErr
					}
					return tt.category, nil
				},
			}

			userRepo := &mocks.MockUserRepository{
				GetUserByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
					if tt.getUserErr != nil {
						return nil, tt.getUserErr
					}
					return tt.shareWithUser, nil
				},
			}

			categoryShareRepo := &mocks.MockCategoryShareRepository{
				GetCategoryShareByCategoryAndUserFunc: func(ctx context.Context, categoryID, userID uint) (*models.CategoryShare, error) {
					if tt.getShareErr != nil {
						return nil, tt.getShareErr
					}
					return tt.existingShare, nil
				},
				CreateCategoryShareFunc: func(ctx context.Context, share *models.CategoryShare) error {
					share.ID = 1
					return tt.createShareErr
				},
			}

			service := createTestCategoryService(categoryRepo, categoryShareRepo, userRepo)
			share, err := service.ShareCategory(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ShareCategory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && share == nil {
				t.Error("ShareCategory() returned nil share")
			}
		})
	}
}

func TestCategoryService_UnshareCategory(t *testing.T) {
	tests := []struct {
		name        string
		req         dto.UnshareCategoryRequest
		category    *models.Category
		share       *models.CategoryShare
		getCatErr   error
		getShareErr error
		deleteErr   error
		wantErr     bool
	}{
		{
			name:     "successful unshare",
			req:      dto.UnshareCategoryRequest{CategoryID: 1, OwnerID: 1, SharedWithUserID: 2},
			category: &models.Category{ID: 1, Name: "Work", OwnerID: 1},
			share:    &models.CategoryShare{ID: 1, CategoryID: 1, SharedWithUserID: 2},
			wantErr:  false,
		},
		{
			name:      "category not found",
			req:       dto.UnshareCategoryRequest{CategoryID: 999, OwnerID: 1, SharedWithUserID: 2},
			getCatErr: sql.ErrNoRows,
			wantErr:   true,
		},
		{
			name:        "share not found",
			req:         dto.UnshareCategoryRequest{CategoryID: 1, OwnerID: 1, SharedWithUserID: 99},
			category:    &models.Category{ID: 1, Name: "Work", OwnerID: 1},
			getShareErr: sql.ErrNoRows,
			wantErr:     true,
		},
		{
			name:     "not owner - forbidden",
			req:      dto.UnshareCategoryRequest{CategoryID: 1, OwnerID: 2, SharedWithUserID: 3},
			category: &models.Category{ID: 1, Name: "Work", OwnerID: 1}, // different owner
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			categoryRepo := &mocks.MockCategoryRepository{
				GetCategoryByIDFunc: func(ctx context.Context, id uint) (*models.Category, error) {
					if tt.getCatErr != nil {
						return nil, tt.getCatErr
					}
					return tt.category, nil
				},
			}

			categoryShareRepo := &mocks.MockCategoryShareRepository{
				GetCategoryShareByCategoryAndUserFunc: func(ctx context.Context, categoryID, userID uint) (*models.CategoryShare, error) {
					if tt.getShareErr != nil {
						return nil, tt.getShareErr
					}
					return tt.share, nil
				},
				DeleteCategoryShareByUserAndCategoryFunc: func(ctx context.Context, categoryID, userID uint) error {
					return tt.deleteErr
				},
			}

			service := createTestCategoryService(categoryRepo, categoryShareRepo, nil)
			err := service.UnshareCategory(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("UnshareCategory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCategoryService_GetCategories(t *testing.T) {
	t.Run("returns user categories", func(t *testing.T) {
		categoryRepo := &mocks.MockCategoryRepository{
			GetCategoriesByOwnerIDFunc: func(ctx context.Context, ownerID uint) ([]models.Category, error) {
				return []models.Category{
					{ID: 1, Name: "Work", OwnerID: ownerID},
					{ID: 2, Name: "Personal", OwnerID: ownerID},
				}, nil
			},
		}

		service := createTestCategoryService(categoryRepo, nil, nil)
		categories, err := service.GetCategories(context.Background(), 1)

		if err != nil {
			t.Errorf("GetCategories() error = %v", err)
		}
		if len(categories) != 2 {
			t.Errorf("GetCategories() returned %d categories, want 2", len(categories))
		}
	})

	t.Run("returns empty list for new user", func(t *testing.T) {
		categoryRepo := &mocks.MockCategoryRepository{
			GetCategoriesByOwnerIDFunc: func(ctx context.Context, ownerID uint) ([]models.Category, error) {
				return []models.Category{}, nil
			},
		}

		service := createTestCategoryService(categoryRepo, nil, nil)
		categories, err := service.GetCategories(context.Background(), 1)

		if err != nil {
			t.Errorf("GetCategories() error = %v", err)
		}
		if len(categories) != 0 {
			t.Errorf("GetCategories() returned %d categories, want 0", len(categories))
		}
	})
}

func TestCategoryService_GetSharesForCategory(t *testing.T) {
	t.Run("owner can get shares", func(t *testing.T) {
		categoryRepo := &mocks.MockCategoryRepository{
			GetCategoryByIDFunc: func(ctx context.Context, id uint) (*models.Category, error) {
				return &models.Category{ID: 1, Name: "Work", OwnerID: 1}, nil
			},
		}

		categoryShareRepo := &mocks.MockCategoryShareRepository{
			GetSharesForCategoryFunc: func(ctx context.Context, categoryID uint) ([]models.CategoryShareWithUser, error) {
				return []models.CategoryShareWithUser{
					{ID: 1, CategoryID: 1, SharedWithUserID: 2, SharedWithUserEmail: "user2@test.com"},
				}, nil
			},
		}

		service := createTestCategoryService(categoryRepo, categoryShareRepo, nil)
		shares, err := service.GetSharesForCategory(context.Background(), 1, 1)

		if err != nil {
			t.Errorf("GetSharesForCategory() error = %v", err)
		}
		if len(shares) != 1 {
			t.Errorf("GetSharesForCategory() returned %d shares, want 1", len(shares))
		}
	})

	t.Run("non-owner cannot get shares", func(t *testing.T) {
		categoryRepo := &mocks.MockCategoryRepository{
			GetCategoryByIDFunc: func(ctx context.Context, id uint) (*models.Category, error) {
				return &models.Category{ID: 1, Name: "Work", OwnerID: 1}, nil
			},
		}

		service := createTestCategoryService(categoryRepo, nil, nil)
		_, err := service.GetSharesForCategory(context.Background(), 1, 2) // userID 2 is not owner

		if err == nil {
			t.Error("GetSharesForCategory() expected error for non-owner")
		}
	})
}
