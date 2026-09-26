package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type DB struct {
	sql    *sql.DB
	config Config
	health Health
}

func Open(ctx context.Context, cfg Config) (*DB, error) {
	cfg = cfg.Normalize()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	dsn, err := cfg.DSN()
	if err != nil {
		return nil, err
	}

	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	configurePool(sqlDB, cfg)

	pingCtx, pingCancel := context.WithTimeout(ctx, cfg.HealthTimeout)
	if err := sqlDB.PingContext(pingCtx); err != nil {
		pingCancel()
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	pingCancel()

	healthCtx, healthCancel := context.WithTimeout(ctx, cfg.HealthTimeout)
	health, err := CheckHealth(healthCtx, sqlDB)
	healthCancel()
	if err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	return &DB{sql: sqlDB, config: cfg, health: health}, nil
}

func configurePool(db *sql.DB, cfg Config) {
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
}

func (db *DB) SQL() *sql.DB {
	if db == nil {
		return nil
	}
	return db.sql
}

func (db *DB) Config() Config {
	if db == nil {
		return Config{}
	}
	return db.config
}

func (db *DB) Health() Health {
	if db == nil {
		return Health{}
	}
	return db.health
}

func (db *DB) Stats() sql.DBStats {
	if db == nil || db.sql == nil {
		return sql.DBStats{}
	}
	return db.sql.Stats()
}

func (db *DB) Ping(ctx context.Context) error {
	if db == nil || db.sql == nil {
		return fmt.Errorf("mysql database is not initialized")
	}
	timeout := db.config.HealthTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return db.sql.PingContext(probeCtx)
}

func (db *DB) Close() error {
	if db == nil || db.sql == nil {
		return nil
	}
	return db.sql.Close()
}
