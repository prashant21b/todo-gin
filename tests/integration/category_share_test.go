//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"todo-app/tests/testutil"
)

func TestCategoryShare_ShareGetUpdateUnshare(t *testing.T) {
	testutil.SkipIfNoTestDB(t)
	app, cleanup := testutil.NewTestApp(t, "../../db/schema.sql")
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := testutil.TruncateAll(ctx, app.DB); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	ownerToken := testutil.MustRegister(t, app.Router, "Owner", "owner@share.com", "password123")
	sharedEmail := "shared@share.com"
	testutil.MustRegister(t, app.Router, "Shared User", sharedEmail, "password123")

	// Owner creates a todo (auto-creates category "Work")
	createTodoBody := []byte(`{"title":"Task","description":"","category":"Work"}`)
	w := testutil.Request(app.Router, http.MethodPost, "/api/todos", createTodoBody, ownerToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create todo: expected 201, got %d body=%s", w.Code, w.Body.String())
	}
	var todoResp struct {
		Data struct {
			CategoryID uint `json:"category_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&todoResp); err != nil {
		t.Fatalf("decode todo response: %v", err)
	}
	categoryID := todoResp.Data.CategoryID
	if categoryID == 0 {
		t.Fatal("todo response: expected non-zero category_id")
	}
	categoryIDStr := strconv.FormatUint(uint64(categoryID), 10)

	// Share category with second user (write permission)
	shareBody := []byte(`{"email":"` + sharedEmail + `","permission":"write"}`)
	w = testutil.Request(app.Router, http.MethodPost, "/api/categories/"+categoryIDStr+"/share", shareBody, ownerToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("share category: expected 201, got %d body=%s", w.Code, w.Body.String())
	}

	// Get shares for category
	w = testutil.Request(app.Router, http.MethodGet, "/api/categories/"+categoryIDStr+"/shares", nil, ownerToken)
	if w.Code != http.StatusOK {
		t.Fatalf("get shares: expected 200, got %d", w.Code)
	}
	var sharesResp struct {
		Data []struct {
			SharedWithUserID uint   `json:"shared_with_user_id"`
			Permission       string `json:"permission"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&sharesResp); err != nil {
		t.Fatalf("decode shares: %v", err)
	}
	if len(sharesResp.Data) != 1 {
		t.Fatalf("get shares: expected 1 share, got %d", len(sharesResp.Data))
	}
	if sharesResp.Data[0].Permission != "write" {
		t.Errorf("share permission: expected write, got %s", sharesResp.Data[0].Permission)
	}
	sharedUserID := sharesResp.Data[0].SharedWithUserID
	sharedUserIDStr := strconv.FormatUint(uint64(sharedUserID), 10)

	// Shared user sees category in their list (GET /api/categories)
	sharedToken := testutil.MustLogin(t, app.Router, sharedEmail, "password123")
	w = testutil.Request(app.Router, http.MethodGet, "/api/categories", nil, sharedToken)
	if w.Code != http.StatusOK {
		t.Fatalf("shared user get categories: expected 200, got %d", w.Code)
	}
	var listResp struct {
		Data struct {
			SharedCategories []struct {
				ID         uint   `json:"id"`
				Permission string `json:"permission"`
			} `json:"shared_categories"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&listResp); err != nil {
		t.Fatalf("decode categories list: %v", err)
	}
	if len(listResp.Data.SharedCategories) != 1 || listResp.Data.SharedCategories[0].ID != categoryID {
		t.Errorf("shared user should see 1 shared category; got %d", len(listResp.Data.SharedCategories))
	}

	// Owner updates share permission to read
	updatePermBody := []byte(`{"permission":"read"}`)
	w = testutil.Request(app.Router, http.MethodPut, "/api/categories/"+categoryIDStr+"/shares/"+sharedUserIDStr, updatePermBody, ownerToken)
	if w.Code != http.StatusOK {
		t.Fatalf("update share permission: expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	// Verify permission updated (get shares again)
	w = testutil.Request(app.Router, http.MethodGet, "/api/categories/"+categoryIDStr+"/shares", nil, ownerToken)
	if err := json.NewDecoder(w.Body).Decode(&sharesResp); err != nil {
		t.Fatalf("decode shares again: %v", err)
	}
	if len(sharesResp.Data) != 1 || sharesResp.Data[0].Permission != "read" {
		t.Errorf("after update: expected permission read, got %s", sharesResp.Data[0].Permission)
	}

	// Owner unshares
	w = testutil.Request(app.Router, http.MethodDelete, "/api/categories/"+categoryIDStr+"/shares/"+sharedUserIDStr, nil, ownerToken)
	if w.Code != http.StatusOK {
		t.Fatalf("unshare: expected 200, got %d", w.Code)
	}

	// Shares list should be empty
	w = testutil.Request(app.Router, http.MethodGet, "/api/categories/"+categoryIDStr+"/shares", nil, ownerToken)
	if err := json.NewDecoder(w.Body).Decode(&sharesResp); err != nil {
		t.Fatalf("decode shares after unshare: %v", err)
	}
	if len(sharesResp.Data) != 0 {
		t.Errorf("after unshare: expected 0 shares, got %d", len(sharesResp.Data))
	}
}

func TestCategoryShare_CannotShareWithSelf(t *testing.T) {
	testutil.SkipIfNoTestDB(t)
	app, cleanup := testutil.NewTestApp(t, "../../db/schema.sql")
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := testutil.TruncateAll(ctx, app.DB); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	token := testutil.MustRegister(t, app.Router, "Solo", "solo@share.com", "password123")

	// Create todo (auto-creates category "Personal")
	createTodoBody := []byte(`{"title":"Solo task","description":"","category":"Personal"}`)
	w := testutil.Request(app.Router, http.MethodPost, "/api/todos", createTodoBody, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create todo: expected 201, got %d", w.Code)
	}
	var todoResp struct {
		Data struct {
			CategoryID uint `json:"category_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&todoResp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	idStr := strconv.FormatUint(uint64(todoResp.Data.CategoryID), 10)

	// Try to share with self (same email)
	shareBody := []byte(`{"email":"solo@share.com","permission":"write"}`)
	w = testutil.Request(app.Router, http.MethodPost, "/api/categories/"+idStr+"/share", shareBody, token)
	if w.Code != http.StatusBadRequest {
		t.Errorf("share with self: expected 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCategoryShare_ShareAlreadyExists(t *testing.T) {
	testutil.SkipIfNoTestDB(t)
	app, cleanup := testutil.NewTestApp(t, "../../db/schema.sql")
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := testutil.TruncateAll(ctx, app.DB); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	ownerToken := testutil.MustRegister(t, app.Router, "Owner", "owner2@share.com", "password123")
	sharedEmail := "shared2@share.com"
	testutil.MustRegister(t, app.Router, "Shared", sharedEmail, "password123")

	createTodoBody := []byte(`{"title":"Project task","description":"","category":"Projects"}`)
	w := testutil.Request(app.Router, http.MethodPost, "/api/todos", createTodoBody, ownerToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create todo: expected 201, got %d", w.Code)
	}
	var todoResp struct {
		Data struct {
			CategoryID uint `json:"category_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&todoResp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	idStr := strconv.FormatUint(uint64(todoResp.Data.CategoryID), 10)

	shareBody := []byte(`{"email":"` + sharedEmail + `","permission":"read"}`)
	w = testutil.Request(app.Router, http.MethodPost, "/api/categories/"+idStr+"/share", shareBody, ownerToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("first share: expected 201, got %d", w.Code)
	}

	// Share again with same user -> conflict
	w = testutil.Request(app.Router, http.MethodPost, "/api/categories/"+idStr+"/share", shareBody, ownerToken)
	if w.Code != http.StatusConflict {
		t.Errorf("duplicate share: expected 409, got %d body=%s", w.Code, w.Body.String())
	}
}
