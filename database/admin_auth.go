package database

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrAdminAuthNotFound = errors.New("admin auth record not found")

type AdminUser struct {
	ID                 int64
	Username           string
	PasswordHash       string
	Status             string
	MustChangePassword bool
}

func NormalizeAdminUsername(value string) string { return strings.ToLower(strings.TrimSpace(value)) }

// ValidAdminUsername keeps account identifiers predictable in logs and UI.
func ValidAdminUsername(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) < 3 || len(value) > 64 {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' {
			continue
		}
		return false
	}
	return true
}

func HashAdminSessionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func newAdminPublicID() (string, error) {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	return "adm_" + base64.RawURLEncoding.EncodeToString(random), nil
}

func (db *DB) AdminUserCount(ctx context.Context) (int64, error) {
	var count int64
	if err := db.conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM admin_users`).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (db *DB) CreateAdminUser(ctx context.Context, username, passwordHash string, mustChange bool) (*AdminUser, error) {
	normalized := NormalizeAdminUsername(username)
	if !ValidAdminUsername(username) || strings.TrimSpace(passwordHash) == "" {
		return nil, errors.New("valid username and password hash are required")
	}
	flag := 0
	if mustChange {
		flag = 1
	}
	query := `INSERT INTO admin_users (username, username_normalized, password_hash, status, must_change_password) VALUES (?, ?, ?, 'active', ?)`
	if db.driver == "postgres" {
		query = `INSERT INTO admin_users (username, username_normalized, password_hash, status, must_change_password) VALUES ($1, $2, $3, 'active', $4) RETURNING id`
		var id int64
		if err := db.conn.QueryRowContext(ctx, query, strings.TrimSpace(username), normalized, passwordHash, mustChange).Scan(&id); err != nil {
			return nil, err
		}
		return &AdminUser{ID: id, Username: strings.TrimSpace(username), PasswordHash: passwordHash, Status: "active", MustChangePassword: mustChange}, nil
	}
	publicID, err := newAdminPublicID()
	if err != nil {
		return nil, err
	}
	query = `INSERT INTO admin_users (public_id, username, username_normalized, email, password_hash, status, must_change_password) VALUES (?, ?, ?, NULL, ?, 'active', ?)`
	result, err := db.conn.ExecContext(ctx, query, publicID, strings.TrimSpace(username), normalized, passwordHash, flag)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &AdminUser{ID: id, Username: strings.TrimSpace(username), PasswordHash: passwordHash, Status: "active", MustChangePassword: mustChange}, nil
}

func (db *DB) DeleteAdminUser(ctx context.Context, userID int64) error {
	query := `DELETE FROM admin_users WHERE id=?`
	if db.driver == "postgres" {
		query = strings.Replace(query, "?", "$1", 1)
	}
	_, err := db.conn.ExecContext(ctx, query, userID)
	return err
}

func (db *DB) FindAdminUserByUsername(ctx context.Context, username string) (*AdminUser, error) {
	query := `SELECT id, username, password_hash, status, must_change_password FROM admin_users WHERE username_normalized = ?`
	args := []any{NormalizeAdminUsername(username)}
	if db.driver == "postgres" {
		query = strings.Replace(query, "?", "$1", 1)
	}
	var user AdminUser
	if err := db.conn.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Status, &user.MustChangePassword); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAdminAuthNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (db *DB) CreateAdminSession(ctx context.Context, userID int64, tokenHash, ipAddress, userAgent string, expiresAt time.Time) error {
	query := `INSERT INTO admin_sessions (admin_user_id, token_hash, ip_address, user_agent, expires_at) VALUES (?, ?, ?, ?, ?)`
	if db.driver == "postgres" {
		query = `INSERT INTO admin_sessions (admin_user_id, token_hash, ip_address, user_agent, expires_at) VALUES ($1, $2, $3, $4, $5)`
	}
	_, err := db.conn.ExecContext(ctx, query, userID, tokenHash, ipAddress, userAgent, expiresAt.UTC())
	return err
}

func (db *DB) AdminUserBySession(ctx context.Context, tokenHash string, now time.Time) (*AdminUser, error) {
	query := `SELECT u.id, u.username, u.password_hash, u.status, u.must_change_password FROM admin_sessions s JOIN admin_users u ON u.id=s.admin_user_id WHERE s.token_hash=? AND s.revoked_at IS NULL AND s.expires_at>? AND u.status='active'`
	args := []any{tokenHash, now.UTC()}
	if db.driver == "postgres" {
		query = strings.Replace(strings.Replace(query, "?", "$1", 1), "?", "$2", 1)
	}
	var user AdminUser
	if err := db.conn.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Status, &user.MustChangePassword); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAdminAuthNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (db *DB) RevokeAdminSession(ctx context.Context, tokenHash string) error {
	query := `UPDATE admin_sessions SET revoked_at=CURRENT_TIMESTAMP WHERE token_hash=? AND revoked_at IS NULL`
	if db.driver == "postgres" {
		query = strings.Replace(query, "?", "$1", 1)
	}
	_, err := db.conn.ExecContext(ctx, query, tokenHash)
	return err
}

func (db *DB) MarkAdminLogin(ctx context.Context, userID int64) error {
	query := `UPDATE admin_users SET last_login_at=CURRENT_TIMESTAMP WHERE id=?`
	if db.driver == "postgres" {
		query = strings.Replace(query, "?", "$1", 1)
	}
	_, err := db.conn.ExecContext(ctx, query, userID)
	return err
}

func (db *DB) ChangeAdminPassword(ctx context.Context, userID int64, passwordHash string) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	updateUser := `UPDATE admin_users SET password_hash=?, must_change_password=0, updated_at=CURRENT_TIMESTAMP WHERE id=?`
	revoke := `UPDATE admin_sessions SET revoked_at=CURRENT_TIMESTAMP WHERE admin_user_id=? AND revoked_at IS NULL`
	if db.driver == "postgres" {
		updateUser = `UPDATE admin_users SET password_hash=$1, must_change_password=FALSE, updated_at=CURRENT_TIMESTAMP WHERE id=$2`
		revoke = `UPDATE admin_sessions SET revoked_at=CURRENT_TIMESTAMP WHERE admin_user_id=$1 AND revoked_at IS NULL`
	}
	if _, err := tx.ExecContext(ctx, updateUser, passwordHash, userID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, revoke, userID); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *DB) EnsureAdminAuthSchema(ctx context.Context) error {
	if db.driver != "postgres" {
		return nil
	}
	statements := []string{
		`CREATE TABLE IF NOT EXISTS admin_users (id BIGSERIAL PRIMARY KEY, username VARCHAR(64) NOT NULL, username_normalized VARCHAR(64) NOT NULL UNIQUE, password_hash VARCHAR(255) NOT NULL, status VARCHAR(32) NOT NULL DEFAULT 'active', must_change_password BOOLEAN NOT NULL DEFAULT FALSE, last_login_at TIMESTAMPTZ NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`,
		`CREATE TABLE IF NOT EXISTS admin_sessions (id BIGSERIAL PRIMARY KEY, admin_user_id BIGINT NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE, token_hash CHAR(64) NOT NULL UNIQUE, ip_address VARCHAR(64) NOT NULL DEFAULT '', user_agent VARCHAR(512) NOT NULL DEFAULT '', expires_at TIMESTAMPTZ NOT NULL, revoked_at TIMESTAMPTZ NULL, last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), created_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`,
		`CREATE INDEX IF NOT EXISTS idx_admin_sessions_user_expiry ON admin_sessions(admin_user_id, expires_at)`,
	}
	for _, statement := range statements {
		if _, err := db.conn.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("admin auth schema: %w", err)
		}
	}
	return nil
}
