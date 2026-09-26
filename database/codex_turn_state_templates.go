package database

import (
	"context"
	"database/sql"
	"time"
)

const CodexTurnStateRenewalMaxAttempts = 10

// CodexTurnStateTemplate is an upstream-minted opaque value. Never expose Value in admin JSON.
type CodexTurnStateTemplate struct {
	AccountID int64
	Model     string
	Value     string `json:"-"`
	IssuedAt  int64
	Strikes   int
	UpdatedAt int64
}

func (db *DB) ensureCodexTurnStateTemplateSchema(ctx context.Context) error {
	if db.isMySQL() {
		return nil
	}
	_, err := db.conn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS codex_turn_state_templates (
 account_id BIGINT NOT NULL, model TEXT NOT NULL, value TEXT NOT NULL,
 issued_at BIGINT NOT NULL, strikes INTEGER NOT NULL DEFAULT 0,
 updated_at BIGINT NOT NULL, PRIMARY KEY(account_id,model))`)
	if err != nil {
		return err
	}
	// Keep attempts across restarts without changing existing template rows.
	_, err = db.conn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS codex_turn_state_renewals (
 account_id BIGINT NOT NULL, model TEXT NOT NULL, issued_at BIGINT NOT NULL,
 attempts INTEGER NOT NULL, next_attempt_at BIGINT NOT NULL,
 PRIMARY KEY(account_id,model))`)
	if err != nil {
		return err
	}
	// Store only URL hashes: retries avoid previously attempted exits without
	// copying proxy credentials into renewal history.
	_, err = db.conn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS codex_turn_state_renewal_routes (
 account_id BIGINT NOT NULL, model TEXT NOT NULL, issued_at BIGINT NOT NULL,
 attempt INTEGER NOT NULL, proxy_id BIGINT NOT NULL, proxy_hash TEXT NOT NULL,
 PRIMARY KEY(account_id,model,issued_at,attempt))`)
	if err != nil {
		return err
	}
	return db.ensureCodexTurnStateHistorySchema(ctx)
}

func (db *DB) turnStateMillisArg(value int64) interface{} {
	if !db.isMySQL() {
		return value
	}
	return time.UnixMilli(value).UTC()
}

func mysqlMillisExpr(column string) string {
	return "TIMESTAMPDIFF(MICROSECOND,'1970-01-01 00:00:00'," + column + ") DIV 1000"
}

func (db *DB) SaveCodexTurnStateTemplate(ctx context.Context, row CodexTurnStateTemplate) error {
	if db.isMySQL() {
		_, err := db.conn.ExecContext(ctx, `INSERT INTO codex_turn_state_templates
 (account_id,model,value,issued_at,strikes,updated_at) VALUES ($1,$2,$3,$4,$5,$6)
 ON DUPLICATE KEY UPDATE
 value=IF(VALUES(issued_at)>=issued_at,VALUES(value),value),
 strikes=IF(VALUES(issued_at)>=issued_at,VALUES(strikes),strikes),
 updated_at=IF(VALUES(issued_at)>=issued_at,VALUES(updated_at),updated_at),
 issued_at=GREATEST(issued_at,VALUES(issued_at))`,
			row.AccountID, row.Model, row.Value, db.turnStateMillisArg(row.IssuedAt), row.Strikes, db.turnStateMillisArg(row.UpdatedAt))
		return err
	}
	_, err := db.conn.ExecContext(ctx, `INSERT INTO codex_turn_state_templates
 (account_id,model,value,issued_at,strikes,updated_at) VALUES ($1,$2,$3,$4,$5,$6)
 ON CONFLICT(account_id,model) DO UPDATE SET value=excluded.value,issued_at=excluded.issued_at,
 strikes=excluded.strikes,updated_at=excluded.updated_at
 WHERE excluded.issued_at>=codex_turn_state_templates.issued_at`, row.AccountID, row.Model, row.Value, row.IssuedAt, row.Strikes, row.UpdatedAt)
	return err
}

func (db *DB) GetCodexTurnStateTemplate(ctx context.Context, accountID int64, model string) (CodexTurnStateTemplate, bool, error) {
	var row CodexTurnStateTemplate
	query := `SELECT account_id,model,value,issued_at,strikes,updated_at FROM codex_turn_state_templates WHERE account_id=$1 AND model=$2`
	if db.isMySQL() {
		query = `SELECT account_id,model,value,` + mysqlMillisExpr("issued_at") + `,strikes,` + mysqlMillisExpr("updated_at") + `
			FROM codex_turn_state_templates WHERE account_id=$1 AND model=$2`
	}
	err := db.conn.QueryRowContext(ctx, query, accountID, model).Scan(&row.AccountID, &row.Model, &row.Value, &row.IssuedAt, &row.Strikes, &row.UpdatedAt)
	if err == sql.ErrNoRows {
		return row, false, nil
	}
	return row, err == nil, err
}

func (db *DB) ListCodexTurnStateTemplates(ctx context.Context, accountID int64) ([]CodexTurnStateTemplate, error) {
	query := `SELECT account_id,model,value,issued_at,strikes,updated_at FROM codex_turn_state_templates WHERE account_id=$1 ORDER BY model`
	if db.isMySQL() {
		query = `SELECT account_id,model,value,` + mysqlMillisExpr("issued_at") + `,strikes,` + mysqlMillisExpr("updated_at") + `
			FROM codex_turn_state_templates WHERE account_id=$1 ORDER BY model`
	}
	return db.listCodexTurnStateTemplates(ctx, query, accountID)
}

func (db *DB) ListCodexTurnStateRenewalCandidates(ctx context.Context, now int64) ([]CodexTurnStateTemplate, error) {
	query := `SELECT t.account_id,t.model,t.value,t.issued_at,t.strikes,t.updated_at
 FROM codex_turn_state_templates t LEFT JOIN codex_turn_state_renewals r ON t.account_id=r.account_id AND t.model=r.model
 WHERE r.account_id IS NULL OR t.issued_at>r.issued_at OR
 (t.issued_at=r.issued_at AND r.attempts<$2 AND r.next_attempt_at<=$1)
 ORDER BY t.issued_at,t.account_id,t.model`
	args := []any{now, CodexTurnStateRenewalMaxAttempts}
	if db.isMySQL() {
		query = `SELECT t.account_id,t.model,t.value,` + mysqlMillisExpr("t.issued_at") + `,t.strikes,` + mysqlMillisExpr("t.updated_at") + `
 FROM codex_turn_state_templates t LEFT JOIN codex_turn_state_renewals r ON t.account_id=r.account_id AND t.model=r.model
 WHERE r.account_id IS NULL OR t.issued_at>r.issued_at OR
 (t.issued_at=r.issued_at AND r.attempts<$2 AND r.next_attempt_at<=$1)
 ORDER BY t.issued_at,t.account_id,t.model`
		args[0] = db.turnStateMillisArg(now)
	}
	return db.listCodexTurnStateTemplates(ctx, query, args...)
}

func (db *DB) listCodexTurnStateTemplates(ctx context.Context, query string, args ...any) ([]CodexTurnStateTemplate, error) {
	rows, err := db.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []CodexTurnStateTemplate{}
	for rows.Next() {
		var row CodexTurnStateTemplate
		if err := rows.Scan(&row.AccountID, &row.Model, &row.Value, &row.IssuedAt, &row.Strikes, &row.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// ClaimCodexTurnStateRenewal atomically reserves one of ten attempts for this
// upstream issuance. Ordinary re-saves of the same token do not reset the budget.
func (db *DB) ClaimCodexTurnStateRenewal(ctx context.Context, row CodexTurnStateTemplate, now, leaseUntil int64) (int, error) {
	if db.isMySQL() {
		attempt := 0
		err := db.withWriteTx(ctx, func(tx *sql.Tx) error {
			var templateIssued int64
			err := tx.QueryRowContext(ctx,
				`SELECT `+mysqlMillisExpr("issued_at")+` FROM codex_turn_state_templates WHERE account_id=$1 AND model=$2 FOR UPDATE`,
				row.AccountID, row.Model,
			).Scan(&templateIssued)
			if err == sql.ErrNoRows || templateIssued != row.IssuedAt {
				return nil
			}
			if err != nil {
				return err
			}

			var existingIssued, nextAttempt int64
			var attempts int
			err = tx.QueryRowContext(ctx,
				`SELECT `+mysqlMillisExpr("issued_at")+`,attempts,`+mysqlMillisExpr("next_attempt_at")+`
				 FROM codex_turn_state_renewals WHERE account_id=$1 AND model=$2 FOR UPDATE`,
				row.AccountID, row.Model,
			).Scan(&existingIssued, &attempts, &nextAttempt)
			if err == sql.ErrNoRows {
				_, err = tx.ExecContext(ctx, `INSERT INTO codex_turn_state_renewals
					(account_id,model,issued_at,attempts,next_attempt_at) VALUES ($1,$2,$3,1,$4)`,
					row.AccountID, row.Model, db.turnStateMillisArg(row.IssuedAt), db.turnStateMillisArg(leaseUntil))
				if err == nil {
					attempt = 1
				}
				return err
			}
			if err != nil {
				return err
			}

			switch {
			case row.IssuedAt > existingIssued:
				attempt = 1
			case row.IssuedAt == existingIssued && attempts < CodexTurnStateRenewalMaxAttempts && nextAttempt <= now:
				attempt = attempts + 1
			default:
				return nil
			}
			_, err = tx.ExecContext(ctx, `UPDATE codex_turn_state_renewals
				SET issued_at=$3,attempts=$4,next_attempt_at=$5
				WHERE account_id=$1 AND model=$2`,
				row.AccountID, row.Model, db.turnStateMillisArg(row.IssuedAt), attempt, db.turnStateMillisArg(leaseUntil))
			return err
		})
		return attempt, err
	}

	var attempt int
	err := db.conn.QueryRowContext(ctx, `INSERT INTO codex_turn_state_renewals
 (account_id,model,issued_at,attempts,next_attempt_at)
 SELECT $1,$2,$3,1,$5 WHERE EXISTS (SELECT 1 FROM codex_turn_state_templates
 WHERE account_id=$1 AND model=$2 AND issued_at=$3)
 ON CONFLICT(account_id,model) DO UPDATE SET issued_at=excluded.issued_at,
 attempts=CASE WHEN excluded.issued_at>codex_turn_state_renewals.issued_at THEN 1 ELSE codex_turn_state_renewals.attempts+1 END,
 next_attempt_at=excluded.next_attempt_at
 WHERE excluded.issued_at>codex_turn_state_renewals.issued_at OR
 (excluded.issued_at=codex_turn_state_renewals.issued_at AND codex_turn_state_renewals.attempts<$6 AND codex_turn_state_renewals.next_attempt_at<=$4)
 RETURNING attempts`, row.AccountID, row.Model, row.IssuedAt, now, leaseUntil, CodexTurnStateRenewalMaxAttempts).Scan(&attempt)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return attempt, err
}

func (db *DB) FinishCodexTurnStateRenewal(ctx context.Context, row CodexTurnStateTemplate, attempt int, nextAttempt int64) error {
	args := []any{row.AccountID, row.Model, row.IssuedAt, attempt, nextAttempt}
	if db.isMySQL() {
		args[2] = db.turnStateMillisArg(row.IssuedAt)
		args[4] = db.turnStateMillisArg(nextAttempt)
	}
	_, err := db.conn.ExecContext(ctx, `UPDATE codex_turn_state_renewals SET next_attempt_at=$5
 WHERE account_id=$1 AND model=$2 AND issued_at=$3 AND attempts=$4`, args...)
	return err
}

func (db *DB) PruneCodexTurnStateRenewals(ctx context.Context) error {
	_, err := db.conn.ExecContext(ctx, `DELETE FROM codex_turn_state_renewals WHERE NOT EXISTS
 (SELECT 1 FROM codex_turn_state_templates t WHERE t.account_id=codex_turn_state_renewals.account_id AND t.model=codex_turn_state_renewals.model)`)
	if err != nil {
		return err
	}
	_, err = db.conn.ExecContext(ctx, `DELETE FROM codex_turn_state_renewal_routes WHERE NOT EXISTS
 (SELECT 1 FROM codex_turn_state_renewals r WHERE r.account_id=codex_turn_state_renewal_routes.account_id
 AND r.model=codex_turn_state_renewal_routes.model AND r.issued_at=codex_turn_state_renewal_routes.issued_at)`)
	return err
}

func (db *DB) DeleteCodexTurnStateTemplate(ctx context.Context, accountID int64, model string) error {
	_, err := db.conn.ExecContext(ctx, `DELETE FROM codex_turn_state_templates WHERE account_id=$1 AND ($2='' OR model=$2)`, accountID, model)
	return err
}

func (db *DB) PruneCodexTurnStateTemplates(ctx context.Context, oldestIssued int64, maxEntries int) error {
	oldestArg := interface{}(oldestIssued)
	if db.isMySQL() {
		oldestArg = db.turnStateMillisArg(oldestIssued)
	}
	_, err := db.conn.ExecContext(ctx, `DELETE FROM codex_turn_state_templates WHERE issued_at<=$1`, oldestArg)
	if err != nil {
		return err
	}
	if maxEntries < 0 {
		return nil
	}
	if db.isMySQL() {
		_, err = db.conn.ExecContext(ctx, `DELETE t FROM codex_turn_state_templates t
			JOIN (
				SELECT account_id,model FROM codex_turn_state_templates
				ORDER BY issued_at DESC,account_id,model
				LIMIT 18446744073709551615 OFFSET $1
			) stale ON stale.account_id=t.account_id AND stale.model=t.model`, maxEntries)
		return err
	}
	limit := "ALL"
	if db.isSQLite() {
		limit = "-1"
	}
	_, err = db.conn.ExecContext(ctx, `DELETE FROM codex_turn_state_templates WHERE (account_id,model) IN
 (SELECT account_id,model FROM codex_turn_state_templates ORDER BY issued_at DESC,account_id,model LIMIT `+limit+` OFFSET $1)`, maxEntries)
	return err
}

// CodexTurnStateRenewalProxyHashes returns exits already tried for this issuance.
func (db *DB) CodexTurnStateRenewalProxyHashes(ctx context.Context, row CodexTurnStateTemplate) (map[string]bool, error) {
	issuedAt := interface{}(row.IssuedAt)
	if db.isMySQL() {
		issuedAt = db.turnStateMillisArg(row.IssuedAt)
	}
	rows, err := db.conn.QueryContext(ctx, `SELECT proxy_hash FROM codex_turn_state_renewal_routes
 WHERE account_id=$1 AND model=$2 AND issued_at=$3 ORDER BY attempt`, row.AccountID, row.Model, issuedAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	used := make(map[string]bool)
	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			return nil, err
		}
		used[hash] = true
	}
	return used, rows.Err()
}

func (db *DB) RecordCodexTurnStateRenewalProxy(ctx context.Context, row CodexTurnStateTemplate, attempt int, proxyID int64, hash string) error {
	issuedAt := interface{}(row.IssuedAt)
	query := `INSERT INTO codex_turn_state_renewal_routes
 (account_id,model,issued_at,attempt,proxy_id,proxy_hash) VALUES ($1,$2,$3,$4,$5,$6)
 ON CONFLICT(account_id,model,issued_at,attempt) DO NOTHING`
	if db.isMySQL() {
		issuedAt = db.turnStateMillisArg(row.IssuedAt)
		query = `INSERT IGNORE INTO codex_turn_state_renewal_routes
			(account_id,model,issued_at,attempt,proxy_id,proxy_hash) VALUES ($1,$2,$3,$4,$5,$6)`
	}
	_, err := db.conn.ExecContext(ctx, query, row.AccountID, row.Model, issuedAt, attempt, proxyID, hash)
	return err
}
