package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// WithTx executes fn inside a transaction and guarantees rollback on every
// non-commit path. A rollback failure is joined with the original callback
// error so callers do not lose the primary failure.
func WithTx(ctx context.Context, db *sql.DB, options *sql.TxOptions, fn func(*sql.Tx) error) (err error) {
	if db == nil {
		return fmt.Errorf("mysql database is nil")
	}
	if fn == nil {
		return fmt.Errorf("mysql transaction callback is nil")
	}

	tx, err := db.BeginTx(ctx, options)
	if err != nil {
		return fmt.Errorf("begin mysql transaction: %w", err)
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			_ = tx.Rollback()
			panic(recovered)
		}
		if err == nil {
			return
		}
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			err = errors.Join(err, fmt.Errorf("rollback mysql transaction: %w", rollbackErr))
		}
	}()

	if err = fn(tx); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit mysql transaction: %w", err)
	}
	return nil
}

func (db *DB) WithTx(ctx context.Context, options *sql.TxOptions, fn func(*sql.Tx) error) error {
	if db == nil {
		return fmt.Errorf("mysql database is nil")
	}
	return WithTx(ctx, db.sql, options, fn)
}
