package mysql

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestMigrationMetadataDDLContract(t *testing.T) {
	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS schema_migrations",
		"version BIGINT UNSIGNED",
		"name VARCHAR(255)",
		"checksum CHAR(64) CHARACTER SET ascii COLLATE ascii_bin",
		"applied_at DATETIME(3)",
	} {
		if !strings.Contains(migrationTableDDL, fragment) {
			t.Fatalf("migration metadata DDL missing %q", fragment)
		}
	}
}

func TestSplitSQLStatementsIgnoresQuotedAndCommentSemicolons(t *testing.T) {
	input := "-- comment; still comment\nCREATE TABLE t (v VARCHAR(32));\n" +
		"INSERT INTO t(v) VALUES ('a;b'); # hash; comment\n" +
		"/* block; comment */ UPDATE t SET v=\"c;d\";"
	statements := splitSQLStatements(input)
	if len(statements) != 3 {
		t.Fatalf("statements = %d, want 3: %#v", len(statements), statements)
	}
	for _, statement := range statements {
		if strings.TrimSpace(statement) == "" {
			t.Fatal("empty statement returned")
		}
	}
}

func TestProjectSQLMigrationsContract(t *testing.T) {
	dir := filepath.Join("..", "..", "..", "migrations")
	files, err := LoadSQLMigrationFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 20 {
		t.Fatalf("migration files = %d, want 20", len(files))
	}
	for i, file := range files {
		wantVersion := uint64(i + 1)
		if file.Version != wantVersion {
			t.Fatalf("migration[%d].Version = %d, want %d", i, file.Version, wantVersion)
		}
		if !validMigrationChecksum(file.Checksum) {
			t.Fatalf("migration %s checksum invalid: %q", filepath.Base(file.Path), file.Checksum)
		}
	}

	migrations, err := BuildSQLMigrations(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(migrations) != len(files) {
		t.Fatalf("built migrations = %d, want %d", len(migrations), len(files))
	}

	var all strings.Builder
	for _, file := range files {
		all.WriteString(file.SQL)
		all.WriteByte('\n')
		for _, statement := range splitSQLStatements(file.SQL) {
			upper := strings.ToUpper(statement)
			if !strings.Contains(upper, "CREATE TABLE") {
				continue
			}
			if !strings.Contains(upper, "ENGINE=INNODB") {
				t.Fatalf("%s has CREATE TABLE without InnoDB", filepath.Base(file.Path))
			}
			if !strings.Contains(strings.ToLower(statement), "charset=utf8mb4") {
				t.Fatalf("%s has CREATE TABLE without utf8mb4", filepath.Base(file.Path))
			}
		}
	}
	sqlText := all.String()

	legacyTables := []string{
		"account_daily_usage", "account_events", "account_group_members", "account_groups",
		"account_model_cooldowns", "accounts", "api_key_auth_cache_state",
		"api_key_model_request_counters", "api_key_model_request_ledger", "api_key_scope_counters",
		"api_keys", "codex_invite_recipients", "codex_invite_snapshots",
		"codex_oauth_refresh_attempts", "codex_turn_state_renewal_history",
		"codex_turn_state_renewal_routes", "codex_turn_state_renewals",
		"codex_turn_state_templates", "data_migrations", "grok_account_fact_snapshots",
		"grok_credential_identity_claims", "grok_model_capabilities", "grok_model_catalog_items",
		"grok_model_catalog_snapshots", "grok_state_migration_progress", "image_assets",
		"image_generation_jobs", "image_prompt_templates", "maintenance_jobs",
		"model_capability_snapshots", "model_registry", "model_registry_sync",
		"official_pricing_sync_config", "prompt_conversation_locks", "prompt_filter_logs",
		"prompt_filter_newapi_bindings", "prompt_log_retention_config", "prompt_policy_incidents",
		"prompt_review_profiles", "prompt_risk_event_sources", "prompt_risk_events",
		"prompt_risk_identities", "prompt_risk_trust_events", "prompt_risk_trust_policies",
		"prompt_rule_candidate_evidence", "prompt_rule_candidates", "proxies",
		"proxy_risk_score_snapshots", "proxy_risk_scoring_profiles", "quality_test_jobs",
		"quality_test_prompts", "scheduler_outbox", "system_settings", "usage_logs",
		"usage_stats_baseline", "usage_stats_rollup", "usage_stats_rollup_state",
	}
	commercialTables := []string{
		"users", "profiles", "sessions", "login_logs", "email_verifications",
		"admin_users", "roles", "permissions", "admin_user_roles", "role_permissions",
		"wallets", "wallet_transactions", "orders", "payments", "payment_webhooks",
		"payment_provider_configs", "plans", "plan_model_permissions", "user_subscriptions",
		"providers", "provider_accounts", "models", "provider_models", "model_prices",
		"provider_costs", "user_api_keys", "usage_records", "billing_records",
		"xarrpay_configs", "notifications", "audit_logs",
	}
	for _, table := range append(legacyTables, commercialTables...) {
		needle := "CREATE TABLE IF NOT EXISTS " + table
		if !strings.Contains(sqlText, needle) {
			t.Fatalf("migration schema missing table %s", table)
		}
	}

	prohibited := regexp.MustCompile(`(?i)\b(postgres|postgresql|jsonb|timestamptz|bigserial|returning)\b|on\s+conflict|\$[0-9]+`)
	if match := prohibited.FindString(sqlText); match != "" {
		t.Fatalf("migration SQL contains PostgreSQL syntax %q", match)
	}

	moneyBinaryFloat := regexp.MustCompile(`(?i)(amount|balance|quota_limit|quota_used|total_used|used_cost|account_billed|user_billed|image_unit_price|credits|price_amount|provider_cost|user_charge|margin)[^,\n]*(float|double)`)
	if match := moneyBinaryFloat.FindString(sqlText); match != "" {
		t.Fatalf("money column uses binary floating SQL type: %q", match)
	}
	for _, precisionInvariant := range []string{
		"MODIFY COLUMN account_billed DECIMAL(20,8)",
		"MODIFY COLUMN user_billed DECIMAL(20,8)",
		"MODIFY COLUMN quota_limit DECIMAL(20,8)",
		"MODIFY COLUMN quota_used DECIMAL(20,8)",
		"MODIFY COLUMN total_used DECIMAL(20,8)",
		"MODIFY COLUMN used_cost DECIMAL(20,8)",
	} {
		if !strings.Contains(sqlText, precisionInvariant) {
			t.Fatalf("billing precision migration missing %q", precisionInvariant)
		}
	}

	for _, invariant := range []string{
		"UNIQUE KEY uk_users_public_id (public_id)",
		"UNIQUE KEY uk_users_email (email)",
		"UNIQUE KEY uk_users_email_normalized (email_normalized)",
		"UNIQUE KEY uk_sessions_token_hash (token_hash)",
		"UNIQUE KEY uk_wallets_user_id (user_id)",
		"UNIQUE KEY uk_orders_order_no (order_no)",
		"UNIQUE KEY uk_payments_payment_no (payment_no)",
		"UNIQUE KEY uk_billing_records_request_id (request_id)",
	} {
		if !strings.Contains(sqlText, invariant) {
			t.Fatalf("migration invariant missing %q", invariant)
		}
	}
}
