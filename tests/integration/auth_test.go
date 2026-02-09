//go:build integration

package integration

import (
	"context"
	"net/http"
	"testing"
	"time"

	"todo-app/tests/testutil"
)

func TestAuth_RegisterAndLogin(t *testing.T) {
	testutil.SkipIfNoTestDB(t)
	app, cleanup := testutil.NewTestApp(t, "../../db/schema.sql")
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := testutil.TruncateAll(ctx, app.DB); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	token := testutil.MustRegister(t, app.Router, "Test User", "test@example.com", "password123")
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	loginToken := testutil.MustLogin(t, app.Router, "test@example.com", "password123")
	if loginToken != token {
		t.Log("login token may differ from register token (new JWT each time); at least both non-empty")
	}
}

func TestAuth_RegisterDuplicateEmail(t *testing.T) {
	testutil.SkipIfNoTestDB(t)
	app, cleanup := testutil.NewTestApp(t, "../../db/schema.sql")
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := testutil.TruncateAll(ctx, app.DB); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	testutil.MustRegister(t, app.Router, "First", "dup@example.com", "password123")
	_, status, _ := testutil.Register(app.Router, "Second", "dup@example.com", "password456")
	if status != http.StatusConflict {
		t.Errorf("duplicate register: expected 409, got %d", status)
	}
}

func TestAuth_LoginWrongPassword(t *testing.T) {
	testutil.SkipIfNoTestDB(t)
	app, cleanup := testutil.NewTestApp(t, "../../db/schema.sql")
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := testutil.TruncateAll(ctx, app.DB); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	testutil.MustRegister(t, app.Router, "User", "wrong@example.com", "correct")
	_, status, _ := testutil.Login(app.Router, "wrong@example.com", "wrong")
	if status != http.StatusUnauthorized {
		t.Errorf("wrong password: expected 401, got %d", status)
	}
}

func TestAuth_ProtectedRouteWithoutToken(t *testing.T) {
	testutil.SkipIfNoTestDB(t)
	app, cleanup := testutil.NewTestApp(t, "../../db/schema.sql")
	defer cleanup()

	w := testutil.Request(app.Router, http.MethodGet, "/api/todos", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("GET /api/todos without token: expected 401, got %d", w.Code)
	}
}
