package testutil

import (
	"context"
	"os"
	"strings"
	"time"

	"todo-app/db"
)

// TruncateAll deletes all data from test tables in dependency order.
// Call between tests to get a clean state. Uses the same DB connection from TestApp.
// No-op if SKIP_TRUNCATE is set to "true" or "1" (e.g. when demoing so DB data is preserved).
func TruncateAll(ctx context.Context, database *db.DB) error {
	if strings.TrimSpace(strings.ToLower(os.Getenv("SKIP_TRUNCATE"))) == "true" ||
		os.Getenv("SKIP_TRUNCATE") == "1" {
		return nil
	}
	if database == nil || database.SQL == nil {
		return nil
	}
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tables := []string{"todos", "category_shares", "categories", "users"}
	for _, table := range tables {
		if _, err := database.SQL.ExecContext(timeout, "DELETE FROM "+table); err != nil {
			return err
		}
	}
	return nil
}
