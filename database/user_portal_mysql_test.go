package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

func TestPortalUserAndAPIKeyLifecycleMySQL(t *testing.T) {
	db, err := newTestDatabase(t, "portal-user-lifecycle")
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()
	if err := db.ValidatePortalSchema(ctx); err != nil {
		t.Fatalf("ValidatePortalSchema: %v", err)
	}

	expiresAt := time.Now().UTC().Add(time.Hour)
	user, sessionID, err := db.CreatePortalUser(ctx, CreatePortalUserInput{
		PublicID:        "usr_0123456789abcdefghijkl",
		Email:           "portal@example.com",
		EmailNormalized: "portal@example.com",
		PasswordHash:    "password-hash",
		DisplayName:     "Portal User",
		Locale:          "zh-CN",
		Timezone:        "Asia/Shanghai",
		SessionHash:     "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		ClientIP:        "127.0.0.1",
		UserAgent:       "AxisRelay test",
		SessionExpires:  expiresAt,
	})
	if err != nil {
		t.Fatalf("CreatePortalUser: %v", err)
	}
	if user == nil || user.Email != "portal@example.com" || sessionID <= 0 {
		t.Fatalf("unexpected portal user: user=%#v session=%d", user, sessionID)
	}

	resolved, resolvedSessionID, err := db.GetPortalUserBySession(ctx,
		"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", time.Now().UTC())
	if err != nil || resolved == nil || resolved.ID != user.ID || resolvedSessionID != sessionID {
		t.Fatalf("GetPortalUserBySession: user=%#v session=%d err=%v", resolved, resolvedSessionID, err)
	}

	rawKey := "sk-axis-portal-integration-key"
	key, err := db.CreatePortalAPIKey(ctx, user.ID, "Production", rawKey, "sk-axis-portal-int", "key-hash", 3, 1)
	if err != nil {
		t.Fatalf("CreatePortalAPIKey: %v", err)
	}
	if key == nil || key.ID <= 0 || key.GatewayKeyID <= 0 || key.QuotaLimit != 3 {
		t.Fatalf("unexpected portal key: %#v", key)
	}
	if _, err := db.CreatePortalAPIKey(ctx, user.ID, "Overflow", rawKey+"-2", "sk-axis-overflow", "key-hash-2", 3, 1); !errors.Is(err, ErrPortalAPIKeyLimit) {
		t.Fatalf("second CreatePortalAPIKey error = %v, want ErrPortalAPIKeyLimit", err)
	}
	gatewayKey, err := db.GetAPIKeyByValue(ctx, rawKey)
	if err != nil || gatewayKey.ID != key.GatewayKeyID || gatewayKey.QuotaLimit != 3 {
		t.Fatalf("GetAPIKeyByValue: key=%#v err=%v", gatewayKey, err)
	}

	if err := db.RenamePortalAPIKey(ctx, user.ID, key.ID, "Renamed"); err != nil {
		t.Fatalf("RenamePortalAPIKey: %v", err)
	}
	listed, err := db.ListPortalAPIKeys(ctx, user.ID)
	if err != nil || len(listed) != 1 || listed[0].Name != "Renamed" {
		t.Fatalf("ListPortalAPIKeys: keys=%#v err=%v", listed, err)
	}

	if err := db.RevokePortalAPIKey(ctx, user.ID, key.ID); err != nil {
		t.Fatalf("RevokePortalAPIKey: %v", err)
	}
	if _, err := db.GetAPIKeyByValue(ctx, rawKey); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("revoked gateway key lookup error = %v, want sql.ErrNoRows", err)
	}
	var status string
	var gatewayID sql.NullInt64
	if err := db.conn.QueryRowContext(ctx, `
		SELECT status, gateway_api_key_id FROM user_api_keys WHERE id=$1`, key.ID).Scan(&status, &gatewayID); err != nil {
		t.Fatalf("read revoked portal key audit row: %v", err)
	}
	if status != "revoked" || gatewayID.Valid {
		t.Fatalf("revoked portal key audit row = status %q gateway %#v", status, gatewayID)
	}
}
