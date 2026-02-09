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

func TestTodo_CRUD(t *testing.T) {
	testutil.SkipIfNoTestDB(t)
	app, cleanup := testutil.NewTestApp(t, "../../db/schema.sql")
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := testutil.TruncateAll(ctx, app.DB); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	token := testutil.MustRegister(t, app.Router, "Todo User", "todo@example.com", "password123")

	// Create
	createBody := []byte(`{"title":"Integration todo","description":"From test","category":"TestCategory"}`)
	w := testutil.Request(app.Router, http.MethodPost, "/api/todos", createBody, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create todo: expected 201, got %d body=%s", w.Code, w.Body.String())
	}
	var createResp struct {
		Data struct {
			ID uint `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	id := createResp.Data.ID
	if id == 0 {
		t.Fatal("create: expected non-zero id")
	}

	// Get list
	w = testutil.Request(app.Router, http.MethodGet, "/api/todos", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("get todos: expected 200, got %d", w.Code)
	}
	var listResp struct {
		Data []interface{} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&listResp); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listResp.Data) != 1 {
		t.Errorf("get todos: expected 1 item, got %d", len(listResp.Data))
	}

	idStr := strconv.FormatUint(uint64(id), 10)

	// Get by ID
	w = testutil.Request(app.Router, http.MethodGet, "/api/todos/"+idStr, nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("get todo by id: expected 200, got %d", w.Code)
	}

	// Update
	updateBody := []byte(`{"title":"Updated title","completed":true}`)
	w = testutil.Request(app.Router, http.MethodPut, "/api/todos/"+idStr, updateBody, token)
	if w.Code != http.StatusOK {
		t.Fatalf("update todo: expected 200, got %d", w.Code)
	}

	// Delete
	w = testutil.Request(app.Router, http.MethodDelete, "/api/todos/"+idStr, nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("delete todo: expected 200, got %d", w.Code)
	}

	// Get after delete should 404
	w = testutil.Request(app.Router, http.MethodGet, "/api/todos/"+idStr, nil, token)
	if w.Code != http.StatusNotFound {
		t.Errorf("get deleted todo: expected 404, got %d", w.Code)
	}
}
