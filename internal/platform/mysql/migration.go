package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	migrationLockName    = "axisrelay_schema_migrations"
	migrationLockTimeout = 30
)

const migrationTableDDL = `CREATE TABLE IF NOT EXISTS axisrelay_schema_migrations (
	version BIGINT NOT NULL,
	name VARCHAR(191) NOT NULL,
	applied_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
	PRIMARY KEY (version),
	UNIQUE KEY uk_axisrelay_schema_migrations_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci`

// Migration describes one monotonic, idempotent schema step.
// MySQL DDL may commit implicitly, so every Up function must tolerate replay.
type Migration struct {
	Version int64
	Name    string
	Up      func(context.Context, MigrationDB) error
}

// MigrationDB is deliberately narrower than *sql.DB. The runner pins a
// dedicated *sql.Conn while holding GET_LOCK so migrations and lock ownership
// stay on the same MySQL session, including when the pool has MaxOpenConns=1.
type MigrationDB interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type MigrationRecord struct {
	Version   int64
	Name      string
	AppliedAt time.Time
}

type Migrator struct {
	db *sql.DB
}

func NewMigrator(db *sql.DB) (*Migrator, error) {
	if db == nil {
		return nil, fmt.Errorf("mysql migration database is nil")
	}
	return &Migrator{db: db}, nil
}

func (m *Migrator) Run(ctx context.Context, migrations []Migration) (err error) {
	if m == nil || m.db == nil {
		return fmt.Errorf("mysql migrator is not initialized")
	}
	ordered, err := prepareMigrations(migrations)
	if err != nil {
		return err
	}
	if len(ordered) == 0 {
		return nil
	}

	conn, release, err := acquireMigrationLock(ctx, m.db)
	if err != nil {
		return err
	}
	defer func() {
		if releaseErr := release(); err == nil && releaseErr != nil {
			err = releaseErr
		}
	}()

	if _, err := conn.ExecContext(ctx, migrationTableDDL); err != nil {
		return fmt.Errorf("create mysql migration table: %w", err)
	}

	applied, err := appliedVersions(ctx, conn)
	if err != nil {
		return err
	}

	for _, migration := range ordered {
		if name, ok := applied[migration.Version]; ok {
			if name != migration.Name {
				return fmt.Errorf("mysql migration version %d already applied as %q, current name %q", migration.Version, name, migration.Name)
			}
			continue
		}

		if err := migration.Up(ctx, conn); err != nil {
			return fmt.Errorf("apply mysql migration %d %q: %w", migration.Version, migration.Name, err)
		}
		if _, err := conn.ExecContext(ctx, `INSERT INTO axisrelay_schema_migrations(version, name) VALUES (?, ?)`, migration.Version, migration.Name); err != nil {
			return fmt.Errorf("record mysql migration %d %q: %w", migration.Version, migration.Name, err)
		}
	}
	return nil
}

func (m *Migrator) Applied(ctx context.Context) ([]MigrationRecord, error) {
	if m == nil || m.db == nil {
		return nil, fmt.Errorf("mysql migrator is not initialized")
	}
	if _, err := m.db.ExecContext(ctx, migrationTableDDL); err != nil {
		return nil, fmt.Errorf("create mysql migration table: %w", err)
	}

	rows, err := m.db.QueryContext(ctx, `SELECT version, name, applied_at FROM axisrelay_schema_migrations ORDER BY version`)
	if err != nil {
		return nil, fmt.Errorf("list mysql migrations: %w", err)
	}
	defer rows.Close()

	records := make([]MigrationRecord, 0)
	for rows.Next() {
		var record MigrationRecord
		if err := rows.Scan(&record.Version, &record.Name, &record.AppliedAt); err != nil {
			return nil, fmt.Errorf("scan mysql migration: %w", err)
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate mysql migrations: %w", err)
	}
	return records, nil
}

func appliedVersions(ctx context.Context, db MigrationDB) (map[int64]string, error) {
	rows, err := db.QueryContext(ctx, `SELECT version, name FROM axisrelay_schema_migrations ORDER BY version`)
	if err != nil {
		return nil, fmt.Errorf("read mysql migration state: %w", err)
	}
	defer rows.Close()

	applied := make(map[int64]string)
	for rows.Next() {
		var version int64
		var name string
		if err := rows.Scan(&version, &name); err != nil {
			return nil, fmt.Errorf("scan mysql migration state: %w", err)
		}
		applied[version] = name
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate mysql migration state: %w", err)
	}
	return applied, nil
}

func prepareMigrations(migrations []Migration) ([]Migration, error) {
	if len(migrations) == 0 {
		return nil, nil
	}

	ordered := append([]Migration(nil), migrations...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Version < ordered[j].Version })

	versions := make(map[int64]string, len(ordered))
	names := make(map[string]int64, len(ordered))
	for i := range ordered {
		migration := &ordered[i]
		migration.Name = strings.TrimSpace(migration.Name)
		switch {
		case migration.Version <= 0:
			return nil, fmt.Errorf("mysql migration version must be positive: %d", migration.Version)
		case migration.Name == "":
			return nil, fmt.Errorf("mysql migration %d has an empty name", migration.Version)
		case len(migration.Name) > 191:
			return nil, fmt.Errorf("mysql migration %d name exceeds 191 bytes", migration.Version)
		case migration.Up == nil:
			return nil, fmt.Errorf("mysql migration %d %q has no Up function", migration.Version, migration.Name)
		}

		if existing, ok := versions[migration.Version]; ok {
			return nil, fmt.Errorf("duplicate mysql migration version %d: %q and %q", migration.Version, existing, migration.Name)
		}
		if existing, ok := names[migration.Name]; ok {
			return nil, fmt.Errorf("duplicate mysql migration name %q: versions %d and %d", migration.Name, existing, migration.Version)
		}
		versions[migration.Version] = migration.Name
		names[migration.Name] = migration.Version
	}
	return ordered, nil
}

func acquireMigrationLock(ctx context.Context, db *sql.DB) (*sql.Conn, func() error, error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("reserve mysql migration connection: %w", err)
	}

	var acquired sql.NullInt64
	if err := conn.QueryRowContext(ctx, `SELECT GET_LOCK(?, ?)`, migrationLockName, migrationLockTimeout).Scan(&acquired); err != nil {
		_ = conn.Close()
		return nil, nil, fmt.Errorf("acquire mysql migration lock: %w", err)
	}
	if !acquired.Valid || acquired.Int64 != 1 {
		_ = conn.Close()
		return nil, nil, fmt.Errorf("mysql migration lock was not acquired")
	}

	release := func() error {
		defer conn.Close()
		releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var released sql.NullInt64
		if err := conn.QueryRowContext(releaseCtx, `SELECT RELEASE_LOCK(?)`, migrationLockName).Scan(&released); err != nil {
			return fmt.Errorf("release mysql migration lock: %w", err)
		}
		if !released.Valid || released.Int64 != 1 {
			return fmt.Errorf("mysql migration lock was not released")
		}
		return nil
	}
	return conn, release, nil
}
