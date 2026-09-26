package testmysql

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	drivermysql "github.com/go-sql-driver/mysql"
	platformmysql "github.com/wuekevin/axisrelay/internal/platform/mysql"
	migrationassets "github.com/wuekevin/axisrelay/migrations"
)

const (
	defaultRootDSN = "root@tcp(127.0.0.1:3306)/"
	charset        = "utf8mb4"
	collation      = "utf8mb4_0900_ai_ci"
)

type templateTable struct {
	name string
	ddl  string
}

var (
	mu      sync.Mutex
	schemas = map[string]string{}

	templateOnce   sync.Once
	templateErr    error
	templateSchema string
	templateTables []templateTable
)

// DSN returns an isolated MySQL database for one test/legacy database key.
// Reusing the same legacyKey inside the same test reopens the same database,
// preserving the persistence semantics that file-backed tests relied on.
func DSN(tb testing.TB, legacyKey string) string {
	tb.Helper()

	key := fmt.Sprintf("%d|%s|%s", os.Getpid(), tb.Name(), legacyKey)
	mu.Lock()
	if dsn, ok := schemas[key]; ok {
		mu.Unlock()
		return dsn
	}
	mu.Unlock()

	rootCfg := parseRootConfig(tb)
	ensureTemplate(tb, rootCfg)

	sum := sha256.Sum256([]byte(key))
	schema := fmt.Sprintf("axisrelay_test_%x", sum[:8])
	if err := cloneTemplate(rootCfg, schema); err != nil {
		tb.Fatalf("clone MySQL test database %s: %v", schema, err)
	}

	testCfg := cloneConfig(rootCfg)
	testCfg.DBName = schema
	dsn := testCfg.FormatDSN()

	mu.Lock()
	if existing, ok := schemas[key]; ok {
		mu.Unlock()
		_ = dropSchema(rootCfg, schema)
		return existing
	}
	schemas[key] = dsn
	mu.Unlock()

	tb.Cleanup(func() {
		_ = dropSchema(rootCfg, schema)
		mu.Lock()
		delete(schemas, key)
		mu.Unlock()
	})
	return dsn
}

func ensureTemplate(tb testing.TB, rootCfg *drivermysql.Config) {
	tb.Helper()
	templateOnce.Do(func() {
		sum := sha256.Sum256([]byte(fmt.Sprintf("%d|axisrelay-mysql-test-template", os.Getpid())))
		templateSchema = fmt.Sprintf("axisrelay_tpl_%x", sum[:8])

		if err := dropSchema(rootCfg, templateSchema); err != nil {
			templateErr = err
			return
		}
		rootDB, err := openWithConfig(rootCfg)
		if err != nil {
			templateErr = err
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if _, err := rootDB.ExecContext(ctx, "CREATE DATABASE "+quoteIdent(templateSchema)+" CHARACTER SET "+charset+" COLLATE "+collation); err != nil {
			_ = rootDB.Close()
			templateErr = fmt.Errorf("create MySQL test template: %w", err)
			return
		}
		_ = rootDB.Close()

		templateCfg := cloneConfig(rootCfg)
		templateCfg.DBName = templateSchema
		templateDB, err := openWithConfig(templateCfg)
		if err != nil {
			templateErr = err
			return
		}
		defer templateDB.Close()

		migrations, err := platformmysql.BuildSQLMigrationsFS(migrationassets.FS, ".")
		if err != nil {
			templateErr = err
			return
		}
		migrator, err := platformmysql.NewMigrator(templateDB)
		if err != nil {
			templateErr = err
			return
		}
		if err := migrator.Run(ctx, migrations); err != nil {
			templateErr = fmt.Errorf("migrate MySQL test template: %w", err)
			return
		}

		rows, err := templateDB.QueryContext(ctx, `
			SELECT table_name
			FROM information_schema.tables
			WHERE table_schema = ? AND table_type = 'BASE TABLE'
			ORDER BY table_name
		`, templateSchema)
		if err != nil {
			templateErr = fmt.Errorf("list MySQL test template tables: %w", err)
			return
		}
		var tableNames []string
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				_ = rows.Close()
				templateErr = err
				return
			}
			tableNames = append(tableNames, name)
		}
		if err := rows.Close(); err != nil {
			templateErr = err
			return
		}

		templateTables = make([]templateTable, 0, len(tableNames))
		for _, name := range tableNames {
			var gotName, ddl string
			query := "SHOW CREATE TABLE " + quoteIdent(name)
			if err := templateDB.QueryRowContext(ctx, query).Scan(&gotName, &ddl); err != nil {
				templateErr = fmt.Errorf("show create table %s: %w", name, err)
				return
			}
			templateTables = append(templateTables, templateTable{name: name, ddl: ddl})
		}
	})
	if templateErr != nil {
		tb.Fatalf("initialize MySQL test template: %v", templateErr)
	}
}

func cloneTemplate(rootCfg *drivermysql.Config, schema string) error {
	if len(templateTables) == 0 {
		return fmt.Errorf("MySQL test template has no tables")
	}
	if err := dropSchema(rootCfg, schema); err != nil {
		return err
	}

	rootDB, err := openWithConfig(rootCfg)
	if err != nil {
		return err
	}
	defer rootDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := rootDB.ExecContext(ctx, "CREATE DATABASE "+quoteIdent(schema)+" CHARACTER SET "+charset+" COLLATE "+collation); err != nil {
		return err
	}

	destCfg := cloneConfig(rootCfg)
	destCfg.DBName = schema
	destDB, err := openWithConfig(destCfg)
	if err != nil {
		return err
	}
	defer destDB.Close()
	destConn, err := destDB.Conn(ctx)
	if err != nil {
		return err
	}
	defer destConn.Close()
	if _, err := destConn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS=0"); err != nil {
		return err
	}
	for _, table := range templateTables {
		if _, err := destConn.ExecContext(ctx, table.ddl); err != nil {
			return fmt.Errorf("create cloned table %s: %w", table.name, err)
		}
	}
	if _, err := destConn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS=1"); err != nil {
		return err
	}

	rootConn, err := rootDB.Conn(ctx)
	if err != nil {
		return err
	}
	defer rootConn.Close()
	if _, err := rootConn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS=0"); err != nil {
		return err
	}
	for _, table := range templateTables {
		stmt := "INSERT INTO " + quoteIdent(schema) + "." + quoteIdent(table.name) +
			" SELECT * FROM " + quoteIdent(templateSchema) + "." + quoteIdent(table.name)
		if _, err := rootConn.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("copy template table %s: %w", table.name, err)
		}
	}
	if _, err := rootConn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS=1"); err != nil {
		return err
	}
	return nil
}

func dropSchema(rootCfg *drivermysql.Config, schema string) error {
	if schema == "" {
		return nil
	}
	db, err := openWithConfig(rootCfg)
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = db.ExecContext(ctx, "DROP DATABASE IF EXISTS "+quoteIdent(schema))
	return err
}

func openWithConfig(cfg *drivermysql.Config) (*sql.DB, error) {
	copyCfg := cloneConfig(cfg)
	db, err := sql.Open("mysql", copyCfg.FormatDSN())
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func parseRootConfig(tb testing.TB) *drivermysql.Config {
	tb.Helper()
	raw := os.Getenv("AXISRELAY_TEST_MYSQL_ROOT_DSN")
	if raw == "" {
		raw = defaultRootDSN
	}
	cfg, err := drivermysql.ParseDSN(raw)
	if err != nil {
		tb.Fatalf("parse AXISRELAY_TEST_MYSQL_ROOT_DSN: %v", err)
	}
	cfg.DBName = ""
	cfg.ParseTime = true
	cfg.Loc = time.UTC
	cfg.Collation = collation
	if cfg.Params == nil {
		cfg.Params = map[string]string{}
	}
	cfg.Params["charset"] = charset
	cfg.Params["time_zone"] = "'+00:00'"
	return cfg
}

func cloneConfig(cfg *drivermysql.Config) *drivermysql.Config {
	copyCfg := *cfg
	if cfg.Params != nil {
		copyCfg.Params = make(map[string]string, len(cfg.Params))
		for key, value := range cfg.Params {
			copyCfg.Params[key] = value
		}
	}
	return &copyCfg
}

func quoteIdent(value string) string {
	return "`" + strings.ReplaceAll(value, "`", "``") + "`"
}
