package mysql

import (
	"context"
	"strings"
	"testing"
	"time"

	drivermysql "github.com/go-sql-driver/mysql"
)

const testMigrationChecksum = "0000000000000000000000000000000000000000000000000000000000000000"

func TestConfigDSNUsesCommercialMySQLContract(t *testing.T) {
	cfg := Config{
		Host:     "db.internal",
		User:     "axisrelay",
		Password: "secret",
		Database: "axisrelay",
	}
	dsn, err := cfg.DSN()
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := drivermysql.ParseDSN(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Net != "tcp" || parsed.Addr != "db.internal:3306" {
		t.Fatalf("address = %s/%s, want tcp/db.internal:3306", parsed.Net, parsed.Addr)
	}
	if !parsed.ParseTime {
		t.Fatal("parseTime must be true")
	}
	if parsed.Loc != time.UTC {
		t.Fatalf("loc = %v, want UTC", parsed.Loc)
	}
	if !strings.Contains(dsn, "charset=utf8mb4") {
		t.Fatalf("dsn %q does not force charset=utf8mb4", dsn)
	}
}

func TestConfigRejectsNonUTF8MB4(t *testing.T) {
	cfg := Config{Host: "db", User: "axisrelay", Database: "axisrelay", Charset: "utf8"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("non-utf8mb4 charset was accepted")
	}
}

func TestParseMySQLVersion(t *testing.T) {
	for _, tc := range []struct {
		version string
		major   int
		minor   int
		patch   int
		valid   bool
	}{
		{version: "8.0.42", major: 8, minor: 0, patch: 42, valid: true},
		{version: "8.4.6-commercial", major: 8, minor: 4, patch: 6, valid: true},
		{version: "9.1.0", major: 9, minor: 1, patch: 0, valid: true},
		{version: "8.0", major: 8, minor: 0, patch: 0, valid: true},
		{version: "10.11.8-MariaDB", valid: false},
		{version: "", valid: false},
	} {
		t.Run(tc.version, func(t *testing.T) {
			major, minor, patch, err := parseMySQLVersion(tc.version)
			if (err == nil) != tc.valid {
				t.Fatalf("parseMySQLVersion(%q) err=%v, valid=%v", tc.version, err, tc.valid)
			}
			if err == nil && (major != tc.major || minor != tc.minor || patch != tc.patch) {
				t.Fatalf("parseMySQLVersion(%q) = %d.%d.%d, want %d.%d.%d", tc.version, major, minor, patch, tc.major, tc.minor, tc.patch)
			}
		})
	}
}

func TestValidateHealthRequiresMySQL8AndUTF8MB4(t *testing.T) {
	tests := []struct {
		name   string
		health Health
		valid  bool
	}{
		{name: "mysql 8 utf8mb4", health: Health{Version: "8.0.42", CharacterSet: "utf8mb4", Collation: "utf8mb4_0900_ai_ci"}, valid: true},
		{name: "mysql 5.7 rejected", health: Health{Version: "5.7.44", CharacterSet: "utf8mb4", Collation: "utf8mb4_general_ci"}},
		{name: "mariadb rejected", health: Health{Version: "10.11.8-MariaDB", CharacterSet: "utf8mb4", Collation: "utf8mb4_general_ci"}},
		{name: "utf8 rejected", health: Health{Version: "8.0.42", CharacterSet: "utf8mb3", Collation: "utf8mb3_general_ci"}},
		{name: "wrong collation rejected", health: Health{Version: "8.0.42", CharacterSet: "utf8mb4", Collation: "latin1_swedish_ci"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := validateHealth(tc.health)
			if (err == nil) != tc.valid {
				t.Fatalf("validateHealth(%+v) err=%v, valid=%v", tc.health, err, tc.valid)
			}
			if err == nil && got.Major < 8 {
				t.Fatalf("validated major = %d, want >= 8", got.Major)
			}
		})
	}
}

func TestPrepareMigrationsSortsWithoutMutatingInput(t *testing.T) {
	noop := func(context.Context, MigrationDB) error { return nil }
	input := []Migration{
		{Version: 30, Name: " third ", Checksum: testMigrationChecksum, Up: noop},
		{Version: 10, Name: "first", Checksum: testMigrationChecksum, Up: noop},
		{Version: 20, Name: "second", Checksum: testMigrationChecksum, Up: noop},
	}

	ordered, err := prepareMigrations(input)
	if err != nil {
		t.Fatal(err)
	}
	if got := []uint64{ordered[0].Version, ordered[1].Version, ordered[2].Version}; got[0] != 10 || got[1] != 20 || got[2] != 30 {
		t.Fatalf("order = %v", got)
	}
	if ordered[2].Name != "third" {
		t.Fatalf("trimmed name = %q, want third", ordered[2].Name)
	}
	if input[0].Name != " third " {
		t.Fatalf("input mutated: %q", input[0].Name)
	}
}

func TestPrepareMigrationsRejectsInvalidDefinitions(t *testing.T) {
	noop := func(context.Context, MigrationDB) error { return nil }
	longName := strings.Repeat("x", 256)
	tests := []struct {
		name       string
		migrations []Migration
	}{
		{name: "non-positive version", migrations: []Migration{{Version: 0, Name: "bad", Checksum: testMigrationChecksum, Up: noop}}},
		{name: "empty name", migrations: []Migration{{Version: 1, Name: " ", Checksum: testMigrationChecksum, Up: noop}}},
		{name: "long name", migrations: []Migration{{Version: 1, Name: longName, Checksum: testMigrationChecksum, Up: noop}}},
		{name: "invalid checksum", migrations: []Migration{{Version: 1, Name: "bad", Checksum: "xyz", Up: noop}}},
		{name: "nil up", migrations: []Migration{{Version: 1, Name: "bad", Checksum: testMigrationChecksum}}},
		{name: "duplicate version", migrations: []Migration{{Version: 1, Name: "a", Checksum: testMigrationChecksum, Up: noop}, {Version: 1, Name: "b", Checksum: testMigrationChecksum, Up: noop}}},
		{name: "duplicate name", migrations: []Migration{{Version: 1, Name: "same", Checksum: testMigrationChecksum, Up: noop}, {Version: 2, Name: "same", Checksum: testMigrationChecksum, Up: noop}}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := prepareMigrations(tc.migrations); err == nil {
				t.Fatal("invalid migration definition was accepted")
			}
		})
	}
}

func TestMigrationTableUsesInnoDBUTF8MB4(t *testing.T) {
	upper := strings.ToUpper(migrationTableDDL)
	if !strings.Contains(upper, "ENGINE=INNODB") {
		t.Fatal("migration table must use InnoDB")
	}
	if !strings.Contains(strings.ToLower(migrationTableDDL), "charset=utf8mb4") {
		t.Fatal("migration table must use utf8mb4")
	}
}
