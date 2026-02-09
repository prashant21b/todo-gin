//go:build integration

package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"todo-app/tests/testutil"
)

func TestHealth(t *testing.T) {
	testutil.SkipIfNoTestDB(t)
	app, cleanup := testutil.NewTestApp(t, "../../db/schema.sql")
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /api/health: expected 200, got %d", w.Code)
	}
}
