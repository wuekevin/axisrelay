package database

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	platformmysql "github.com/wuekevin/axisrelay/internal/platform/mysql"
)

func openS07MySQLIntegrationDB(t *testing.T) *DB {
	t.Helper()

	dsn := os.Getenv("AXISRELAY_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("AXISRELAY_TEST_MYSQL_DSN is not set")
	}

	conn, err := sql.Open(mysqlCompatDriverName, dsn)
	if err != nil {
		t.Fatalf("open MySQL integration database: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := conn.PingContext(ctx); err != nil {
		t.Fatalf("ping MySQL integration database: %v", err)
	}
	if _, err := platformmysql.CheckHealth(ctx, conn); err != nil {
		t.Fatalf("MySQL integration health check: %v", err)
	}

	migrations, err := platformmysql.BuildSQLMigrations(filepath.Join("..", "migrations"))
	if err != nil {
		t.Fatalf("load MySQL migrations: %v", err)
	}
	migrator, err := platformmysql.NewMigrator(conn)
	if err != nil {
		t.Fatalf("create MySQL migrator: %v", err)
	}
	if err := migrator.Run(ctx, migrations); err != nil {
		t.Fatalf("run MySQL migrations: %v", err)
	}

	return &DB{
		conn:   conn,
		driver: "mysql",
	}
}

func resetS07MySQLIntegrationState(t *testing.T, db *DB) {
	t.Helper()
	ctx := context.Background()

	if _, err := db.conn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS=0"); err != nil {
		t.Fatalf("disable foreign key checks: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.conn.ExecContext(context.Background(), "SET FOREIGN_KEY_CHECKS=1")
	})

	for _, table := range []string{
		"prompt_filter_logs",
		"prompt_risk_events",
		"prompt_risk_event_sources",
		"scheduler_outbox",
		"usage_logs",
		"usage_stats_rollup",
		"usage_stats_baseline",
		"usage_stats_rollup_state",
		"proxies",
		"api_keys",
		"accounts",
		"system_settings",
	} {
		if _, err := db.conn.ExecContext(ctx, "TRUNCATE TABLE "+table); err != nil {
			t.Fatalf("truncate %s: %v", table, err)
		}
	}

	for _, statement := range []string{
		"INSERT INTO system_settings(id) VALUES (1)",
		"INSERT INTO usage_stats_baseline(id) VALUES (1)",
		"INSERT INTO usage_stats_rollup_state(id) VALUES (1)",
	} {
		if _, err := db.conn.ExecContext(ctx, statement); err != nil {
			t.Fatalf("restore MySQL singleton row: %v", err)
		}
	}
}

func TestS07MySQLQueryIntegration(t *testing.T) {
	db := openS07MySQLIntegrationDB(t)
	ctx := context.Background()

	t.Run("accounts", func(t *testing.T) {
		resetS07MySQLIntegrationState(t, db)

		id, err := db.InsertAccount(ctx, "s07-account", "refresh-token", "")
		if err != nil {
			t.Fatalf("InsertAccount: %v", err)
		}
		if id <= 0 {
			t.Fatalf("InsertAccount id = %d, want > 0", id)
		}
		count, err := db.CountAll(ctx)
		if err != nil {
			t.Fatalf("CountAll: %v", err)
		}
		if count != 1 {
			t.Fatalf("CountAll = %d, want 1", count)
		}
	})

	t.Run("api_keys", func(t *testing.T) {
		resetS07MySQLIntegrationState(t, db)

		id, err := db.InsertAPIKey(ctx, "s07-key", "sk-s07-mysql")
		if err != nil {
			t.Fatalf("InsertAPIKey: %v", err)
		}
		if id <= 0 {
			t.Fatalf("InsertAPIKey id = %d, want > 0", id)
		}
		keys, err := db.ListAPIKeys(ctx)
		if err != nil {
			t.Fatalf("ListAPIKeys: %v", err)
		}
		if len(keys) != 1 || keys[0].Key != "sk-s07-mysql" {
			t.Fatalf("ListAPIKeys = %#v", keys)
		}
	})

	t.Run("system_settings", func(t *testing.T) {
		resetS07MySQLIntegrationState(t, db)

		if err := db.UpdateCodexSyncedCLIVersion(ctx, "9.9.9-s07"); err != nil {
			t.Fatalf("UpdateCodexSyncedCLIVersion: %v", err)
		}
		settings, err := db.GetSystemSettings(ctx)
		if err != nil {
			t.Fatalf("GetSystemSettings: %v", err)
		}
		if settings.CodexSyncedCLIVersion != "9.9.9-s07" {
			t.Fatalf("CodexSyncedCLIVersion = %q", settings.CodexSyncedCLIVersion)
		}
	})

	t.Run("usage", func(t *testing.T) {
		resetS07MySQLIntegrationState(t, db)

		if _, err := db.conn.ExecContext(ctx, `
			INSERT INTO usage_logs (
				account_id, endpoint, model, total_tokens, prompt_tokens, completion_tokens,
				input_tokens, output_tokens, status_code, duration_ms, created_at, channel, internal_reason
			) VALUES (0, '/v1/responses', 'gpt-5.5', 10, 6, 4, 6, 4, 200, 12, CURRENT_TIMESTAMP(3), 'codex', '')
		`); err != nil {
			t.Fatalf("seed usage_logs: %v", err)
		}
		stats, err := db.GetUsageStatsSummary(ctx, time.Now().Add(-time.Hour), time.Time{}, "codex")
		if err != nil {
			t.Fatalf("GetUsageStatsSummary: %v", err)
		}
		if stats.TodayRequests != 1 || stats.TodayTokens != 10 {
			t.Fatalf("usage stats requests=%d tokens=%d", stats.TodayRequests, stats.TodayTokens)
		}
	})

	t.Run("scheduler_persistence", func(t *testing.T) {
		resetS07MySQLIntegrationState(t, db)

		if err := db.InsertSchedulerOutboxEvent(ctx, SchedulerEntityAccount, 42, "update"); err != nil {
			t.Fatalf("InsertSchedulerOutboxEvent: %v", err)
		}
		events, err := db.ListSchedulerOutboxEventsAfter(ctx, 0, 10)
		if err != nil {
			t.Fatalf("ListSchedulerOutboxEventsAfter: %v", err)
		}
		if len(events) != 1 || events[0].EntityID != 42 || events[0].EventType != "update" {
			t.Fatalf("scheduler events = %#v", events)
		}
	})

	t.Run("model_pricing", func(t *testing.T) {
		resetS07MySQLIntegrationState(t, db)

		if err := db.UpdateModelPricingSettings(ctx, "{}", "https://pricing.s07.test/models.json"); err != nil {
			t.Fatalf("UpdateModelPricingSettings: %v", err)
		}
		settings, err := db.GetSystemSettings(ctx)
		if err != nil {
			t.Fatalf("GetSystemSettings after pricing update: %v", err)
		}
		if settings.ModelPricingSyncURL != "https://pricing.s07.test/models.json" {
			t.Fatalf("ModelPricingSyncURL = %q", settings.ModelPricingSyncURL)
		}
	})

	t.Run("proxy", func(t *testing.T) {
		resetS07MySQLIntegrationState(t, db)

		id, err := db.InsertProxy(ctx, "http://127.0.0.1:18080", "s07-proxy")
		if err != nil {
			t.Fatalf("InsertProxy: %v", err)
		}
		proxy, err := db.GetProxy(ctx, id)
		if err != nil {
			t.Fatalf("GetProxy: %v", err)
		}
		if proxy.URL != "http://127.0.0.1:18080" || proxy.Label != "s07-proxy" {
			t.Fatalf("proxy = %#v", proxy)
		}
	})

	t.Run("prompt_filter", func(t *testing.T) {
		resetS07MySQLIntegrationState(t, db)

		if _, err := db.conn.ExecContext(ctx, `
			INSERT INTO prompt_filter_logs (
				source, endpoint, action, matched_patterns, created_at
			) VALUES ('s07', '/v1/responses', 'block', JSON_ARRAY('s07-rule'), CURRENT_TIMESTAMP(3))
		`); err != nil {
			t.Fatalf("seed prompt_filter_logs: %v", err)
		}
		logs, err := db.ListPromptFilterLogs(ctx, 10)
		if err != nil {
			t.Fatalf("ListPromptFilterLogs: %v", err)
		}
		if len(logs) != 1 || logs[0].Source != "s07" || logs[0].Action != "block" {
			t.Fatalf("prompt filter logs = %#v", logs)
		}
	})
}
