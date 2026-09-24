package database

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestNewMySQLProductionPath(t *testing.T) {
	dsn := os.Getenv("AXISRELAY_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("AXISRELAY_TEST_MYSQL_DSN is not set")
	}

	db, err := New("mysql", dsn)
	if err != nil {
		t.Fatalf("New(mysql) error: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if got := db.Driver(); got != "mysql" {
		t.Fatalf("Driver() = %q, want mysql", got)
	}
	if got := db.Label(); got != "MySQL" {
		t.Fatalf("Label() = %q, want MySQL", got)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var migrationCount int
	if err := db.conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations").Scan(&migrationCount); err != nil {
		t.Fatalf("count schema_migrations: %v", err)
	}
	if migrationCount != 20 {
		t.Fatalf("schema_migrations count = %d, want 20", migrationCount)
	}

	if _, err := db.GetSystemSettings(ctx); err != nil {
		t.Fatalf("GetSystemSettings() after MySQL migration: %v", err)
	}
}
