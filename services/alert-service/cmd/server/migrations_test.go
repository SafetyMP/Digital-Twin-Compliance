package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadMigrationSQL(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "001_alerts.sql")
	const sql = "-- test migration"
	if err := os.WriteFile(path, []byte(sql), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, err := readMigrationSQL([]string{filepath.Join(dir, "missing.sql"), path})
	if err != nil {
		t.Fatalf("readMigrationSQL: %v", err)
	}
	if got != sql {
		t.Fatalf("sql = %q", got)
	}
}

func TestReadMigrationSQLReturnsNonNotExistError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	_, err := readMigrationSQL([]string{dir})
	if err == nil {
		t.Fatal("expected error when path is a directory")
	}
	if os.IsNotExist(err) {
		t.Fatalf("expected non-NotExist error, got %v", err)
	}
}

func TestMigrationSearchPaths(t *testing.T) {
	t.Parallel()

	paths := migrationSearchPaths()
	if len(paths) != 2 || paths[0] != "migrations/001_alerts.sql" {
		t.Fatalf("paths = %v", paths)
	}
}
