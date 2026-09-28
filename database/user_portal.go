package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrPortalEmailExists = errors.New("portal email already exists")
	ErrPortalAPIKeyLimit = errors.New("portal api key limit reached")
)

type PortalUser struct {
	ID              int64      `json:"-"`
	PublicID        string     `json:"id"`
	Email           string     `json:"email"`
	Status          string     `json:"status"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	DisplayName     string     `json:"display_name"`
	AvatarURL       string     `json:"avatar_url"`
	Phone           string     `json:"phone"`
	Locale          string     `json:"locale"`
	Timezone        string     `json:"timezone"`
}

type PortalSession struct {
	ID         int64      `json:"id"`
	ClientIP   string     `json:"client_ip"`
	UserAgent  string     `json:"user_agent"`
	CreatedAt  time.Time  `json:"created_at"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
	ExpiresAt  time.Time  `json:"expires_at"`
	Current    bool       `json:"current"`
}

type PortalAPIKey struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	Prefix       string     `json:"prefix"`
	Status       string     `json:"status"`
	Enabled      bool       `json:"enabled"`
	QuotaLimit   float64    `json:"quota_limit"`
	QuotaUsed    float64    `json:"quota_used"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	GatewayKeyID int64      `json:"-"`
}

type PortalSubscription struct {
	Code          string     `json:"code"`
	Name          string     `json:"name"`
	Status        string     `json:"status"`
	BillingPeriod string     `json:"billing_period"`
	EndsAt        *time.Time `json:"ends_at,omitempty"`
}

type CreatePortalUserInput struct {
	PublicID        string
	Email           string
	EmailNormalized string
	PasswordHash    string
	DisplayName     string
	Locale          string
	Timezone        string
	SessionHash     string
	ClientIP        string
	UserAgent       string
	SessionExpires  time.Time
}

func nullTimePointer(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	copy := value.Time
	return &copy
}

func scanPortalUser(scanner interface{ Scan(...any) error }) (*PortalUser, error) {
	var user PortalUser
	var verifiedAt, lastLoginAt sql.NullTime
	err := scanner.Scan(
		&user.ID,
		&user.PublicID,
		&user.Email,
		&user.Status,
		&verifiedAt,
		&lastLoginAt,
		&user.CreatedAt,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Phone,
		&user.Locale,
		&user.Timezone,
	)
	if err != nil {
		return nil, err
	}
	user.EmailVerifiedAt = nullTimePointer(verifiedAt)
	user.LastLoginAt = nullTimePointer(lastLoginAt)
	return &user, nil
}

const portalUserSelect = `
	SELECT u.id, u.public_id, u.email, u.status, u.email_verified_at,
	       u.last_login_at, u.created_at, COALESCE(p.display_name, ''),
	       COALESCE(p.avatar_url, ''), COALESCE(p.phone, ''),
	       COALESCE(p.locale, 'zh-CN'), COALESCE(p.timezone, 'UTC')
	FROM users u
	LEFT JOIN profiles p ON p.user_id = u.id`

func (db *DB) CreatePortalUser(ctx context.Context, input CreatePortalUserInput) (*PortalUser, int64, error) {
	if db == nil || db.conn == nil {
		return nil, 0, errors.New("database is unavailable")
	}
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO users (public_id, email, email_normalized, password_hash, status)
		VALUES ($1, $2, $3, $4, 'active')`,
		input.PublicID, input.Email, input.EmailNormalized, input.PasswordHash,
	)
	if err != nil {
		if portalUniqueViolation(err) {
			return nil, 0, ErrPortalEmailExists
		}
		return nil, 0, err
	}
	userID, err := result.LastInsertId()
	if err != nil {
		return nil, 0, err
	}
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO profiles (user_id, display_name, locale, timezone)
		VALUES ($1, $2, $3, $4)`, userID, input.DisplayName, input.Locale, input.Timezone); err != nil {
		return nil, 0, err
	}
	sessionResult, err := tx.ExecContext(ctx, `
		INSERT INTO sessions (user_id, token_hash, client_ip, user_agent, expires_at, last_seen_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		userID, input.SessionHash, input.ClientIP, input.UserAgent, input.SessionExpires, time.Now().UTC(),
	)
	if err != nil {
		return nil, 0, err
	}
	sessionID, err := sessionResult.LastInsertId()
	if err != nil {
		return nil, 0, err
	}
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO login_logs (user_id, email_normalized, success, client_ip, user_agent)
		VALUES ($1, $2, 1, $3, $4)`, userID, input.EmailNormalized, input.ClientIP, input.UserAgent); err != nil {
		return nil, 0, err
	}
	if err = tx.Commit(); err != nil {
		return nil, 0, err
	}
	return db.GetPortalUserByID(ctx, userID), sessionID, nil
}

func (db *DB) GetPortalUserByID(ctx context.Context, userID int64) *PortalUser {
	user, err := scanPortalUser(db.conn.QueryRowContext(ctx, portalUserSelect+` WHERE u.id=$1 AND u.deleted_at IS NULL`, userID))
	if err != nil {
		return nil
	}
	return user
}

func (db *DB) GetPortalUserCredentialsByEmail(ctx context.Context, emailNormalized string) (*PortalUser, string, error) {
	row := db.conn.QueryRowContext(ctx, portalUserSelect+`
		WHERE u.email_normalized=$1 AND u.deleted_at IS NULL`, emailNormalized)
	user, err := scanPortalUser(row)
	if err != nil {
		return nil, "", err
	}
	var passwordHash string
	if err := db.conn.QueryRowContext(ctx, `SELECT password_hash FROM users WHERE id=$1`, user.ID).Scan(&passwordHash); err != nil {
		return nil, "", err
	}
	return user, passwordHash, nil
}

func (db *DB) GetPortalPasswordHash(ctx context.Context, userID int64) (string, error) {
	var passwordHash string
	err := db.conn.QueryRowContext(ctx, `SELECT password_hash FROM users WHERE id=$1 AND deleted_at IS NULL`, userID).Scan(&passwordHash)
	return passwordHash, err
}

func (db *DB) CreatePortalSession(ctx context.Context, userID int64, emailNormalized, tokenHash, clientIP, userAgent string, expiresAt time.Time) (int64, error) {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	now := time.Now().UTC()
	if _, err = tx.ExecContext(ctx, `DELETE FROM sessions WHERE user_id=$1 AND (expires_at <= $2 OR revoked_at IS NOT NULL)`, userID, now); err != nil {
		return 0, err
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO sessions (user_id, token_hash, client_ip, user_agent, expires_at, last_seen_at)
		VALUES ($1, $2, $3, $4, $5, $6)`, userID, tokenHash, clientIP, userAgent, expiresAt, now)
	if err != nil {
		return 0, err
	}
	sessionID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE users SET last_login_at=$1 WHERE id=$2`, now, userID); err != nil {
		return 0, err
	}
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO login_logs (user_id, email_normalized, success, client_ip, user_agent)
		VALUES ($1, $2, 1, $3, $4)`, userID, emailNormalized, clientIP, userAgent); err != nil {
		return 0, err
	}
	return sessionID, tx.Commit()
}

func (db *DB) RecordPortalLoginFailure(ctx context.Context, userID *int64, emailNormalized, reason, clientIP, userAgent string) {
	var userValue any
	if userID != nil && *userID > 0 {
		userValue = *userID
	}
	_, _ = db.conn.ExecContext(ctx, `
		INSERT INTO login_logs (user_id, email_normalized, success, failure_reason, client_ip, user_agent)
		VALUES ($1, $2, 0, $3, $4, $5)`, userValue, emailNormalized, reason, clientIP, userAgent)
}

func (db *DB) GetPortalUserBySession(ctx context.Context, tokenHash string, now time.Time) (*PortalUser, int64, error) {
	var sessionID int64
	err := db.conn.QueryRowContext(ctx, `
		SELECT id FROM sessions
		WHERE token_hash=$1 AND revoked_at IS NULL AND expires_at > $2`, tokenHash, now).Scan(&sessionID)
	if err != nil {
		return nil, 0, err
	}
	row := db.conn.QueryRowContext(ctx, portalUserSelect+`
		JOIN sessions s ON s.user_id=u.id
		WHERE s.id=$1 AND u.status='active' AND u.deleted_at IS NULL`, sessionID)
	user, err := scanPortalUser(row)
	if err != nil {
		return nil, 0, err
	}
	_, _ = db.conn.ExecContext(ctx, `
		UPDATE sessions SET last_seen_at=$1
		WHERE id=$2 AND (last_seen_at IS NULL OR last_seen_at < $3)`, now, sessionID, now.Add(-5*time.Minute))
	return user, sessionID, nil
}

func (db *DB) RevokePortalSessionByHash(ctx context.Context, tokenHash string) error {
	_, err := db.conn.ExecContext(ctx, `UPDATE sessions SET revoked_at=$1 WHERE token_hash=$2 AND revoked_at IS NULL`, time.Now().UTC(), tokenHash)
	return err
}

func (db *DB) ListPortalSessions(ctx context.Context, userID, currentSessionID int64) ([]PortalSession, error) {
	rows, err := db.conn.QueryContext(ctx, `
		SELECT id, client_ip, user_agent, created_at, last_seen_at, expires_at
		FROM sessions
		WHERE user_id=$1 AND revoked_at IS NULL AND expires_at > $2
		ORDER BY COALESCE(last_seen_at, created_at) DESC`, userID, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sessions := make([]PortalSession, 0)
	for rows.Next() {
		var item PortalSession
		var lastSeen sql.NullTime
		if err := rows.Scan(&item.ID, &item.ClientIP, &item.UserAgent, &item.CreatedAt, &lastSeen, &item.ExpiresAt); err != nil {
			return nil, err
		}
		item.LastSeenAt = nullTimePointer(lastSeen)
		item.Current = item.ID == currentSessionID
		sessions = append(sessions, item)
	}
	return sessions, rows.Err()
}

func (db *DB) RevokePortalSession(ctx context.Context, userID, sessionID int64) error {
	result, err := db.conn.ExecContext(ctx, `
		UPDATE sessions SET revoked_at=$1
		WHERE id=$2 AND user_id=$3 AND revoked_at IS NULL`, time.Now().UTC(), sessionID, userID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err == nil && affected == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (db *DB) RevokeOtherPortalSessions(ctx context.Context, userID, currentSessionID int64) error {
	_, err := db.conn.ExecContext(ctx, `
		UPDATE sessions SET revoked_at=$1
		WHERE user_id=$2 AND id<>$3 AND revoked_at IS NULL`, time.Now().UTC(), userID, currentSessionID)
	return err
}

func (db *DB) UpdatePortalProfile(ctx context.Context, userID int64, displayName, phone, locale, timezone string) error {
	result, err := db.conn.ExecContext(ctx, `
		UPDATE profiles SET display_name=$1, phone=$2, locale=$3, timezone=$4
		WHERE user_id=$5`, displayName, phone, locale, timezone, userID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err == nil && affected == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (db *DB) UpdatePortalPassword(ctx context.Context, userID int64, passwordHash string, currentSessionID int64) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE users SET password_hash=$1 WHERE id=$2 AND deleted_at IS NULL`, passwordHash, userID)
	if err != nil {
		return err
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
		return rowsErr
	} else if affected == 0 {
		return sql.ErrNoRows
	}
	if _, err = tx.ExecContext(ctx, `
		UPDATE sessions SET revoked_at=$1 WHERE user_id=$2 AND id<>$3 AND revoked_at IS NULL`,
		time.Now().UTC(), userID, currentSessionID); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *DB) ListPortalAPIKeys(ctx context.Context, userID int64) ([]PortalAPIKey, error) {
	rows, err := db.conn.QueryContext(ctx, `
		SELECT uk.id, uk.name, uk.key_prefix, uk.status, uk.created_at,
		       uk.expires_at, MAX(l.created_at), uk.gateway_api_key_id,
		       COALESCE(g.enabled, 0), COALESCE(g.quota_limit, 0), COALESCE(g.quota_used, 0)
		FROM user_api_keys uk
		LEFT JOIN api_keys g ON g.id=uk.gateway_api_key_id
		LEFT JOIN usage_logs l ON l.api_key_id=uk.gateway_api_key_id
		WHERE uk.user_id=$1 AND uk.revoked_at IS NULL
		GROUP BY uk.id, uk.name, uk.key_prefix, uk.status, uk.created_at, uk.expires_at,
		         uk.gateway_api_key_id, g.enabled, g.quota_limit, g.quota_used
		ORDER BY uk.id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]PortalAPIKey, 0)
	for rows.Next() {
		var item PortalAPIKey
		var expiresAt, lastUsedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.Name, &item.Prefix, &item.Status, &item.CreatedAt,
			&expiresAt, &lastUsedAt, &item.GatewayKeyID, &item.Enabled, &item.QuotaLimit, &item.QuotaUsed); err != nil {
			return nil, err
		}
		item.ExpiresAt = nullTimePointer(expiresAt)
		item.LastUsedAt = nullTimePointer(lastUsedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (db *DB) CreatePortalAPIKey(ctx context.Context, userID int64, name, rawKey, prefix, keyHash string, quotaLimit float64, maxKeys int) (*PortalAPIKey, error) {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	// Serialize key creation per user so concurrent requests cannot bypass the
	// self-service key limit.
	var lockedUserID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM users WHERE id=$1 AND deleted_at IS NULL FOR UPDATE`, userID).Scan(&lockedUserID); err != nil {
		return nil, err
	}
	var activeKeys int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM user_api_keys
		WHERE user_id=$1 AND revoked_at IS NULL`, userID).Scan(&activeKeys); err != nil {
		return nil, err
	}
	if activeKeys >= maxKeys {
		return nil, ErrPortalAPIKeyLimit
	}
	gatewayResult, err := tx.ExecContext(ctx, `
		INSERT INTO api_keys (name, `+"`key`"+`, quota_limit, quota_used, allowed_group_ids, limits, enabled)
		VALUES ($1, $2, $3, 0, '[]', '{}', 1)`, name, rawKey, quotaLimit)
	if err != nil {
		return nil, err
	}
	gatewayID, err := gatewayResult.LastInsertId()
	if err != nil {
		return nil, err
	}
	if err := db.bumpMySQLAPIKeyAuthRevisionTx(ctx, tx, 1); err != nil {
		return nil, err
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO user_api_keys (user_id, gateway_api_key_id, name, key_prefix, key_hash, status)
		VALUES ($1, $2, $3, $4, $5, 'active')`, userID, gatewayID, name, prefix, keyHash)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &PortalAPIKey{ID: id, Name: name, Prefix: prefix, Status: "active", Enabled: true, QuotaLimit: quotaLimit, CreatedAt: time.Now().UTC(), GatewayKeyID: gatewayID}, nil
}

func (db *DB) RevokePortalAPIKey(ctx context.Context, userID, keyID int64) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var gatewayID int64
	if err := tx.QueryRowContext(ctx, `
		SELECT gateway_api_key_id FROM user_api_keys
		WHERE id=$1 AND user_id=$2 AND revoked_at IS NULL FOR UPDATE`, keyID, userID).Scan(&gatewayID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE user_api_keys SET status='revoked', revoked_at=$1 WHERE id=$2`, time.Now().UTC(), keyID); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM api_keys WHERE id=$1`, gatewayID)
	if err != nil {
		return err
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
		return rowsErr
	} else if affected > 0 {
		if err := db.bumpMySQLAPIKeyAuthRevisionTx(ctx, tx, -1); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (db *DB) RenamePortalAPIKey(ctx context.Context, userID, keyID int64, name string) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var gatewayID int64
	if err := tx.QueryRowContext(ctx, `SELECT gateway_api_key_id FROM user_api_keys WHERE id=$1 AND user_id=$2 AND revoked_at IS NULL`, keyID, userID).Scan(&gatewayID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE user_api_keys SET name=$1 WHERE id=$2`, name, keyID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE api_keys SET name=$1 WHERE id=$2`, name, gatewayID); err != nil {
		return err
	}
	if err := db.bumpMySQLAPIKeyAuthRevisionTx(ctx, tx, 0); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *DB) GetPortalAPIKeyUsage(ctx context.Context, userID, keyID int64, start, end time.Time) (*APIKeySelfUsageReport, error) {
	var gatewayID int64
	if err := db.conn.QueryRowContext(ctx, `
		SELECT gateway_api_key_id FROM user_api_keys
		WHERE id=$1 AND user_id=$2 AND revoked_at IS NULL`, keyID, userID).Scan(&gatewayID); err != nil {
		return nil, err
	}
	return db.GetAPIKeySelfUsageReport(ctx, gatewayID, start, end, 1, 10)
}

func (db *DB) GetPortalSubscription(ctx context.Context, userID int64) (*PortalSubscription, error) {
	var item PortalSubscription
	var endsAt sql.NullTime
	err := db.conn.QueryRowContext(ctx, `
		SELECT p.code, p.name, s.status, p.billing_period, s.ends_at
		FROM user_subscriptions s
		JOIN plans p ON p.id=s.plan_id
		WHERE s.user_id=$1 AND s.status='active' AND (s.ends_at IS NULL OR s.ends_at>$2)
		ORDER BY s.id DESC LIMIT 1`, userID, time.Now().UTC()).Scan(
		&item.Code, &item.Name, &item.Status, &item.BillingPeriod, &endsAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return &PortalSubscription{Code: "free", Name: "Free", Status: "active", BillingPeriod: "none"}, nil
	}
	if err != nil {
		return nil, err
	}
	item.EndsAt = nullTimePointer(endsAt)
	return &item, nil
}

func portalUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate entry") || strings.Contains(message, "duplicate key") || strings.Contains(message, "unique constraint") || strings.Contains(message, "23505") || strings.Contains(message, "1062")
}

func (db *DB) ValidatePortalSchema(ctx context.Context) error {
	if db == nil || !db.isMySQL() {
		return fmt.Errorf("user portal requires mysql")
	}
	var count int
	if err := db.conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name IN ('users','profiles','sessions','user_api_keys')`).Scan(&count); err != nil {
		return err
	}
	if count != 4 {
		return fmt.Errorf("user portal schema is incomplete: found %d of 4 required tables", count)
	}
	var gatewayColumnCount int
	if err := db.conn.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema=DATABASE() AND table_name='user_api_keys' AND column_name='gateway_api_key_id'`).Scan(&gatewayColumnCount); err != nil {
		return err
	}
	if gatewayColumnCount != 1 {
		return errors.New("user portal schema is missing user_api_keys.gateway_api_key_id")
	}
	return nil
}
