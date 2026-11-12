package database

import (
	"context"
	"os"
	"testing"
)

func TestDatabaseMigrationAndTables(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://devpulse:devpulse_secret@localhost:5433/devpulse?sslmode=disable"
	}

	ctx := context.Background()
	db, err := Connect(ctx, dbURL)
	if err != nil {
		t.Skipf("Postgres DB not accessible at %s, skipping integration test: %v", dbURL, err)
		return
	}
	defer db.Close()

	// Verify required tables exist
	tables := []string{"accounts", "monitored_repos", "work_logs", "todos", "blockers", "ideas"}
	for _, tbl := range tables {
		var exists bool
		query := `SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' AND table_name = $1
		)`
		err := db.Pool.QueryRow(ctx, query, tbl).Scan(&exists)
		if err != nil {
			t.Fatalf("Failed to check table %s existence: %v", tbl, err)
		}
		if !exists {
			t.Errorf("Expected table %s to exist in database", tbl)
		}
	}
}
