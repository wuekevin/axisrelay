package database

import (
	"context"
	"testing"

	"github.com/wuekevin/axisrelay/internal/testmysql"
)

func newTestDatabase(tb testing.TB, legacyKey string) (*DB, error) {
	tb.Helper()
	return New("mysql", testmysql.DSN(tb, legacyKey))
}

func (db *DB) testTableColumns(ctx context.Context, table string) (map[string]struct{}, error) {
	rows, err := db.conn.QueryContext(ctx, `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = $1
	`, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns := make(map[string]struct{})
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		columns[name] = struct{}{}
	}
	return columns, rows.Err()
}

func (db *DB) testTableIndexes(ctx context.Context, table string) (map[string]bool, error) {
	rows, err := db.conn.QueryContext(ctx, `
		SELECT DISTINCT index_name
		FROM information_schema.statistics
		WHERE table_schema = DATABASE() AND table_name = $1
	`, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	indexes := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		indexes[name] = true
	}
	return indexes, rows.Err()
}
