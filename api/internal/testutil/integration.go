//go:build integration

package testutil

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func OpenIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open postgres connection: %v", err)
	}

	ApplyMigrations(t, db)
	return db
}

func ApplyMigrations(t *testing.T, db *gorm.DB) {
	t.Helper()

	dir := filepath.Join("..", "database", "migrations")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read migrations directory: %v", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		files = append(files, filepath.Join(dir, entry.Name()))
	}
	sort.Strings(files)

	for _, path := range files {
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read migration %s: %v", path, err)
		}
		if err := db.Exec(string(sqlBytes)).Error; err != nil {
			t.Fatalf("apply migration %s: %v", path, err)
		}
	}
}

func ResetIntegrationDB(t *testing.T, db *gorm.DB) {
	t.Helper()

	statements := []string{
		"TRUNCATE TABLE sandbox_events RESTART IDENTITY CASCADE",
		"TRUNCATE TABLE sandboxes RESTART IDENTITY CASCADE",
		"TRUNCATE TABLE images RESTART IDENTITY CASCADE",
		"TRUNCATE TABLE sessions RESTART IDENTITY CASCADE",
		"TRUNCATE TABLE audit_logs RESTART IDENTITY CASCADE",
		"TRUNCATE TABLE users RESTART IDENTITY CASCADE",
	}

	for _, stmt := range statements {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("reset database with %q: %v", stmt, err)
		}
	}
}
