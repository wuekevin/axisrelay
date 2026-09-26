package auth

import (
	"database/sql"
	"testing"

	"github.com/wuekevin/axisrelay/database"
	"github.com/wuekevin/axisrelay/internal/testmysql"
)

func newTestDatabase(tb testing.TB, legacyKey string) (*database.DB, error) {
	tb.Helper()
	return database.New("mysql", testmysql.DSN(tb, legacyKey))
}

func openRawTestDatabase(tb testing.TB, legacyKey string) (*sql.DB, error) {
	tb.Helper()
	return sql.Open("mysql", testmysql.DSN(tb, legacyKey))
}
