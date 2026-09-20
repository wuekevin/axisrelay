-- AxisRelay commercial provider/model schema plus legacy Gateway runtime tables.
CREATE TABLE IF NOT EXISTS providers (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  code VARCHAR(64) NOT NULL,
  name VARCHAR(128) NOT NULL,
  provider_type VARCHAR(32) NOT NULL DEFAULT 'relay',
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  config_json JSON NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id), UNIQUE KEY uk_providers_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS provider_accounts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  provider_id BIGINT UNSIGNED NOT NULL,
  legacy_account_id BIGINT UNSIGNED NULL,
  name VARCHAR(255) NOT NULL DEFAULT '',
  auth_type VARCHAR(32) NOT NULL DEFAULT 'oauth',
  credentials_json JSON NOT NULL,
  proxy_url TEXT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  credential_generation BIGINT UNSIGNED NOT NULL DEFAULT 1,
  metadata_json JSON NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_provider_accounts_provider_status (provider_id, status, id),
  UNIQUE KEY uk_provider_accounts_legacy (legacy_account_id),
  CONSTRAINT fk_provider_accounts_provider FOREIGN KEY (provider_id) REFERENCES providers(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS models (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  model_key VARCHAR(191) NOT NULL,
  display_name VARCHAR(255) NOT NULL DEFAULT '',
  category VARCHAR(64) NOT NULL DEFAULT 'chat',
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  pro_only TINYINT(1) NOT NULL DEFAULT 0,
  capabilities_json JSON NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id), UNIQUE KEY uk_models_model_key (model_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS provider_models (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  provider_id BIGINT UNSIGNED NOT NULL,
  model_id BIGINT UNSIGNED NOT NULL,
  upstream_model VARCHAR(191) NOT NULL,
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  priority INT NOT NULL DEFAULT 0,
  config_json JSON NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_provider_models_provider_upstream (provider_id, upstream_model),
  UNIQUE KEY uk_provider_models_provider_model (provider_id, model_id),
  CONSTRAINT fk_provider_models_provider FOREIGN KEY (provider_id) REFERENCES providers(id) ON DELETE CASCADE,
  CONSTRAINT fk_provider_models_model FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS model_prices (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  model_id BIGINT UNSIGNED NOT NULL,
  currency CHAR(3) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT 'USD',
  input_microunits_per_million BIGINT UNSIGNED NOT NULL DEFAULT 0,
  output_microunits_per_million BIGINT UNSIGNED NOT NULL DEFAULT 0,
  cached_input_microunits_per_million BIGINT UNSIGNED NOT NULL DEFAULT 0,
  image_input_microunits_per_million BIGINT UNSIGNED NOT NULL DEFAULT 0,
  image_output_microunits_per_million BIGINT UNSIGNED NOT NULL DEFAULT 0,
  effective_at DATETIME(3) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id), KEY idx_model_prices_model_effective (model_id, effective_at),
  CONSTRAINT fk_model_prices_model FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS provider_costs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  provider_id BIGINT UNSIGNED NOT NULL,
  model_id BIGINT UNSIGNED NOT NULL,
  currency CHAR(3) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT 'USD',
  input_microunits_per_million BIGINT UNSIGNED NOT NULL DEFAULT 0,
  output_microunits_per_million BIGINT UNSIGNED NOT NULL DEFAULT 0,
  cached_input_microunits_per_million BIGINT UNSIGNED NOT NULL DEFAULT 0,
  effective_at DATETIME(3) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id), KEY idx_provider_costs_provider_model (provider_id, model_id, effective_at),
  CONSTRAINT fk_provider_costs_provider FOREIGN KEY (provider_id) REFERENCES providers(id) ON DELETE CASCADE,
  CONSTRAINT fk_provider_costs_model FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Legacy Gateway account and routing tables retained for S0.6 query migration.
CREATE TABLE IF NOT EXISTS accounts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(255) NOT NULL DEFAULT '', platform VARCHAR(32) NOT NULL DEFAULT 'openai', type VARCHAR(32) NOT NULL DEFAULT 'oauth',
  credentials JSON NOT NULL, proxy_url TEXT NULL, status VARCHAR(32) NOT NULL DEFAULT 'active', cooldown_reason VARCHAR(128) NOT NULL DEFAULT '',
  cooldown_until DATETIME(3) NULL, score_bias_override BIGINT NULL, base_concurrency_override BIGINT NULL, skip_warm_tier TINYINT(1) NOT NULL DEFAULT 0,
  note TEXT NULL, error_message TEXT NULL, tags JSON NULL, credential_generation BIGINT UNSIGNED NOT NULL DEFAULT 1,
  enabled TINYINT(1) NOT NULL DEFAULT 1, locked TINYINT(1) NOT NULL DEFAULT 0, credit_enabled TINYINT(1) NOT NULL DEFAULT 0, credit_skip_usage_window TINYINT(1) NOT NULL DEFAULT 0,
  image_quota_remaining BIGINT NULL, image_quota_total BIGINT NULL, today_used_count BIGINT UNSIGNED NOT NULL DEFAULT 0, image_quota_reset_at DATETIME(3) NULL,
  credential_family_id VARCHAR(191) NOT NULL DEFAULT '', deleted_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id), KEY idx_accounts_status (status), KEY idx_accounts_platform (platform), KEY idx_accounts_cooldown_until (cooldown_until),
  KEY idx_accounts_created_id (created_at, id), KEY idx_accounts_updated_id (updated_at, id), KEY idx_accounts_credential_family (credential_family_id, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS account_groups (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, name VARCHAR(255) NOT NULL, description TEXT NULL, color VARCHAR(32) NOT NULL DEFAULT '',
  sort_order INT NOT NULL DEFAULT 0, base_concurrency_override BIGINT NULL, proxy_urls JSON NULL, channel VARCHAR(32) NOT NULL DEFAULT 'codex',
  auto_pause_5h_threshold DECIMAL(8,4) NOT NULL DEFAULT 0, auto_pause_7d_threshold DECIMAL(8,4) NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id), UNIQUE KEY uk_account_groups_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS account_group_members (
  account_id BIGINT UNSIGNED NOT NULL, group_id BIGINT UNSIGNED NOT NULL, created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (account_id, group_id), KEY idx_account_group_members_group (group_id),
  CONSTRAINT fk_account_group_members_account FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
  CONSTRAINT fk_account_group_members_group FOREIGN KEY (group_id) REFERENCES account_groups(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS account_model_cooldowns (
  account_id BIGINT UNSIGNED NOT NULL, model VARCHAR(191) NOT NULL, reason VARCHAR(255) NOT NULL DEFAULT '',
  reset_at DATETIME(3) NOT NULL, updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (account_id, model), KEY idx_account_model_cooldowns_reset_at (reset_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS account_events (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, account_id BIGINT UNSIGNED NOT NULL DEFAULT 0, event_type VARCHAR(64) NOT NULL, source VARCHAR(64) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), PRIMARY KEY (id), KEY idx_account_events_created (created_at), KEY idx_account_events_type_created (event_type, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS proxies (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, url TEXT NOT NULL, url_hash CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL, label VARCHAR(255) NOT NULL DEFAULT '',
  enabled TINYINT(1) NOT NULL DEFAULT 1, test_ip VARCHAR(45) NOT NULL DEFAULT '', test_location VARCHAR(255) NOT NULL DEFAULT '', test_latency_ms BIGINT UNSIGNED NOT NULL DEFAULT 0,
  test_status VARCHAR(32) NOT NULL DEFAULT 'untested', created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id), UNIQUE KEY uk_proxies_url_hash (url_hash)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS model_registry (
  id VARCHAR(191) NOT NULL, enabled TINYINT(1) NOT NULL DEFAULT 1, category VARCHAR(64) NOT NULL DEFAULT 'codex', source VARCHAR(64) NOT NULL DEFAULT 'manual',
  pro_only TINYINT(1) NOT NULL DEFAULT 0, api_key_auth_available TINYINT(1) NOT NULL DEFAULT 1, last_seen_at DATETIME(3) NULL,
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3), PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS model_registry_sync (
  id TINYINT UNSIGNED NOT NULL DEFAULT 1, source_url TEXT NULL, last_synced_at DATETIME(3) NULL, PRIMARY KEY (id), CONSTRAINT chk_model_registry_sync_singleton CHECK (id=1)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS model_capability_snapshots (
  account_id BIGINT UNSIGNED NOT NULL, credential_generation BIGINT UNSIGNED NOT NULL, observed_at DATETIME(3) NOT NULL, models_json JSON NOT NULL, PRIMARY KEY (account_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS scheduler_outbox (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, entity_type VARCHAR(64) NOT NULL, entity_id BIGINT UNSIGNED NOT NULL DEFAULT 0, event_type VARCHAR(64) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), PRIMARY KEY (id), KEY idx_scheduler_outbox_created (created_at, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS maintenance_jobs (
  entity_id BIGINT UNSIGNED NOT NULL, job_kind VARCHAR(64) NOT NULL, due_at DATETIME(3) NOT NULL, lease_owner VARCHAR(191) NOT NULL DEFAULT '',
  lease_until DATETIME(3) NULL, attempts INT UNSIGNED NOT NULL DEFAULT 0, last_error TEXT NULL, updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (entity_id, job_kind), KEY idx_maintenance_jobs_due (job_kind, due_at, entity_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS codex_oauth_refresh_attempts (
  rt_fingerprint CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL, attempt_id VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  started_at DATETIME(3) NOT NULL, PRIMARY KEY (rt_fingerprint), UNIQUE KEY uk_codex_refresh_attempt_id (attempt_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS codex_turn_state_templates (
  account_id BIGINT UNSIGNED NOT NULL, model VARCHAR(191) NOT NULL, value LONGTEXT NOT NULL, issued_at DATETIME(3) NOT NULL, strikes INT UNSIGNED NOT NULL DEFAULT 0,
  updated_at DATETIME(3) NOT NULL, PRIMARY KEY (account_id, model)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS codex_turn_state_renewals (
  account_id BIGINT UNSIGNED NOT NULL, model VARCHAR(191) NOT NULL, issued_at DATETIME(3) NOT NULL, attempts INT UNSIGNED NOT NULL DEFAULT 0,
  next_attempt_at DATETIME(3) NOT NULL, PRIMARY KEY (account_id, model)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS codex_turn_state_renewal_routes (
  account_id BIGINT UNSIGNED NOT NULL, model VARCHAR(191) NOT NULL, issued_at DATETIME(3) NOT NULL, attempt INT UNSIGNED NOT NULL,
  proxy_id BIGINT UNSIGNED NOT NULL DEFAULT 0, proxy_hash CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL, PRIMARY KEY (account_id, model, issued_at, attempt)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS codex_turn_state_renewal_history (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, account_id BIGINT UNSIGNED NOT NULL, account_name VARCHAR(255) NOT NULL, plan_type VARCHAR(64) NOT NULL,
  model VARCHAR(191) NOT NULL, attempt INT UNSIGNED NOT NULL, max_attempts INT UNSIGNED NOT NULL, proxy_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  proxy_name VARCHAR(255) NOT NULL DEFAULT '', proxy_url TEXT NULL, proxy_ip VARCHAR(45) NOT NULL DEFAULT '', route VARCHAR(64) NOT NULL DEFAULT '',
  started_at DATETIME(3) NOT NULL, finished_at DATETIME(3) NULL, duration_ms BIGINT UNSIGNED NOT NULL DEFAULT 0, status VARCHAR(32) NOT NULL DEFAULT 'running',
  reason TEXT NULL, expires_before DATETIME(3) NOT NULL, expires_after DATETIME(3) NULL, PRIMARY KEY (id),
  KEY idx_turn_state_history_account (account_id, id), KEY idx_turn_state_history_status (status, started_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS codex_invite_snapshots (
  account_id BIGINT UNSIGNED NOT NULL, snapshot_kind VARCHAR(64) NOT NULL, scope VARCHAR(191) NOT NULL DEFAULT '', credential_generation BIGINT UNSIGNED NOT NULL DEFAULT 1,
  http_status INT NOT NULL DEFAULT 0, payload_json JSON NOT NULL, observed_at DATETIME(3) NOT NULL, expires_at DATETIME(3) NOT NULL,
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3), PRIMARY KEY (account_id, snapshot_kind, scope), KEY idx_codex_invite_snapshots_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS codex_invite_recipients (
  email_key VARCHAR(320) NOT NULL, email VARCHAR(320) NOT NULL, sender_account_id BIGINT UNSIGNED NOT NULL DEFAULT 0, program_id VARCHAR(128) NOT NULL DEFAULT '',
  entrypoint VARCHAR(128) NOT NULL DEFAULT '', state VARCHAR(32) NOT NULL, reservation_id VARCHAR(128) NOT NULL DEFAULT '', request_id VARCHAR(191) NOT NULL DEFAULT '',
  referral_id VARCHAR(191) NOT NULL DEFAULT '', invite_url TEXT NULL, upstream_status INT NOT NULL DEFAULT 0, upstream_recipient_status VARCHAR(64) NOT NULL DEFAULT '',
  invited_at DATETIME(3) NULL, created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (email_key), KEY idx_codex_invite_recipients_state (state, updated_at), KEY idx_codex_invite_recipients_reservation (reservation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS grok_account_fact_snapshots (
  account_id BIGINT UNSIGNED NOT NULL, fact_kind VARCHAR(64) NOT NULL, credential_generation BIGINT UNSIGNED NOT NULL, status VARCHAR(32) NOT NULL DEFAULT 'unknown',
  http_status INT NOT NULL DEFAULT 0, source VARCHAR(64) NOT NULL DEFAULT '', payload_json JSON NOT NULL, field_presence_json JSON NOT NULL,
  observed_at DATETIME(3) NOT NULL, expires_at DATETIME(3) NOT NULL, updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (account_id, fact_kind), KEY idx_grok_facts_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS grok_model_catalog_snapshots (
  account_id BIGINT UNSIGNED NOT NULL, origin VARCHAR(64) NOT NULL, credential_generation BIGINT UNSIGNED NOT NULL, auth_kind VARCHAR(32) NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL DEFAULT 'unknown', http_etag VARCHAR(255) NOT NULL DEFAULT '', etag_hint VARCHAR(255) NOT NULL DEFAULT '', etag_hint_observed_at DATETIME(3) NULL,
  observed_at DATETIME(3) NOT NULL, expires_at DATETIME(3) NOT NULL, updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (account_id, origin)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS grok_model_catalog_items (
  account_id BIGINT UNSIGNED NOT NULL, origin VARCHAR(64) NOT NULL, model_id VARCHAR(191) NOT NULL, credential_generation BIGINT UNSIGNED NOT NULL,
  display_name VARCHAR(255) NOT NULL DEFAULT '', description TEXT NULL, base_url TEXT NULL, api_base_url TEXT NULL, api_backend VARCHAR(64) NOT NULL DEFAULT '',
  context_window BIGINT UNSIGNED NOT NULL DEFAULT 0, max_output_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0, reasoning_effort VARCHAR(64) NOT NULL DEFAULT '',
  reasoning_efforts_json JSON NOT NULL, supports_reasoning_effort TINYINT(1) NOT NULL DEFAULT 0, supports_backend_search TINYINT(1) NOT NULL DEFAULT 0,
  stream_tool_calls TINYINT(1) NOT NULL DEFAULT 0, supported_in_api TINYINT(1) NOT NULL DEFAULT 1, hidden TINYINT(1) NOT NULL DEFAULT 0,
  extra_headers_json JSON NOT NULL, field_presence_json JSON NOT NULL, first_seen_at DATETIME(3) NOT NULL, PRIMARY KEY (account_id, origin, model_id),
  KEY idx_grok_catalog_items_model (model_id, account_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS grok_model_capabilities (
  account_id BIGINT UNSIGNED NOT NULL, model_id VARCHAR(191) NOT NULL, origin VARCHAR(64) NOT NULL, protocol VARCHAR(64) NOT NULL, credential_generation BIGINT UNSIGNED NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'untested', http_status INT NOT NULL DEFAULT 0, provider_code VARCHAR(128) NOT NULL DEFAULT '', source VARCHAR(64) NOT NULL DEFAULT '',
  retry_after_seconds BIGINT UNSIGNED NOT NULL DEFAULT 0, observed_at DATETIME(3) NOT NULL, expires_at DATETIME(3) NOT NULL, updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (account_id, model_id, origin, protocol), KEY idx_grok_capabilities_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS grok_credential_identity_claims (
  identity_key VARCHAR(255) NOT NULL, account_id BIGINT UNSIGNED NOT NULL, created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (identity_key), KEY idx_grok_identity_claims_account (account_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS grok_state_migration_progress (
  version VARCHAR(64) NOT NULL, phase VARCHAR(64) NOT NULL DEFAULT 'families', last_account_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  completed_at DATETIME(3) NULL, updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3), PRIMARY KEY (version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS image_prompt_templates (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, name VARCHAR(255) NOT NULL DEFAULT '', prompt LONGTEXT NOT NULL, model VARCHAR(191) NOT NULL DEFAULT '',
  size VARCHAR(32) NOT NULL DEFAULT '', quality VARCHAR(32) NOT NULL DEFAULT '', output_format VARCHAR(32) NOT NULL DEFAULT '', background VARCHAR(64) NOT NULL DEFAULT '',
  style VARCHAR(64) NOT NULL DEFAULT '', tags JSON NULL, favorite TINYINT(1) NOT NULL DEFAULT 0, usage_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
  last_used_at DATETIME(3) NULL, created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id), KEY idx_image_prompt_templates_updated (updated_at), KEY idx_image_prompt_templates_favorite (favorite, updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS image_generation_jobs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, status VARCHAR(32) NOT NULL DEFAULT 'queued', prompt LONGTEXT NOT NULL, params_json JSON NOT NULL,
  api_key_id BIGINT UNSIGNED NOT NULL DEFAULT 0, api_key_name VARCHAR(255) NOT NULL DEFAULT '', api_key_masked VARCHAR(64) NOT NULL DEFAULT '', error_message TEXT NULL,
  duration_ms BIGINT UNSIGNED NOT NULL DEFAULT 0, created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), started_at DATETIME(3) NULL, completed_at DATETIME(3) NULL,
  PRIMARY KEY (id), KEY idx_image_generation_jobs_created (created_at), KEY idx_image_generation_jobs_status (status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS image_assets (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, job_id BIGINT UNSIGNED NOT NULL DEFAULT 0, template_id BIGINT UNSIGNED NOT NULL DEFAULT 0, filename VARCHAR(255) NOT NULL DEFAULT '',
  storage_path TEXT NOT NULL, mime_type VARCHAR(128) NOT NULL DEFAULT '', bytes BIGINT UNSIGNED NOT NULL DEFAULT 0, width INT UNSIGNED NOT NULL DEFAULT 0, height INT UNSIGNED NOT NULL DEFAULT 0,
  model VARCHAR(191) NOT NULL DEFAULT '', requested_size VARCHAR(32) NOT NULL DEFAULT '', actual_size VARCHAR(32) NOT NULL DEFAULT '', quality VARCHAR(32) NOT NULL DEFAULT '',
  output_format VARCHAR(32) NOT NULL DEFAULT '', revised_prompt LONGTEXT NULL, created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id), KEY idx_image_assets_created (created_at), KEY idx_image_assets_job_id (job_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS quality_test_jobs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, slot TINYINT UNSIGNED NULL, account_id BIGINT UNSIGNED NOT NULL, account_name VARCHAR(255) NOT NULL, plan_type VARCHAR(64) NOT NULL,
  channel VARCHAR(32) NOT NULL, model VARCHAR(191) NOT NULL, reasoning_effort VARCHAR(64) NOT NULL DEFAULT '', prompt LONGTEXT NOT NULL, status VARCHAR(32) NOT NULL DEFAULT 'running',
  output LONGTEXT NULL, preset_kind VARCHAR(64) NOT NULL DEFAULT '', preset_ref VARCHAR(191) NOT NULL DEFAULT '', preset_name VARCHAR(255) NOT NULL DEFAULT '',
  metrics_json JSON NULL, error TEXT NULL, created_at DATETIME(3) NOT NULL, updated_at DATETIME(3) NOT NULL, deadline_at DATETIME(3) NOT NULL, completed_at DATETIME(3) NULL,
  PRIMARY KEY (id), UNIQUE KEY uk_quality_test_jobs_slot (slot), KEY idx_quality_test_jobs_account_status (account_id, status), KEY idx_quality_test_created (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS quality_test_prompts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, name VARCHAR(255) NOT NULL, prompt LONGTEXT NOT NULL, usage_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
  last_used_at DATETIME(3) NULL, created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS official_pricing_sync_config (
  singleton_id TINYINT UNSIGNED NOT NULL DEFAULT 1, enabled TINYINT(1) NOT NULL DEFAULT 0, interval_minutes INT UNSIGNED NOT NULL DEFAULT 1440,
  include_openai TINYINT(1) NOT NULL DEFAULT 1, include_grok TINYINT(1) NOT NULL DEFAULT 1, include_claude TINYINT(1) NOT NULL DEFAULT 1,
  last_attempt_at DATETIME(3) NULL, last_success_at DATETIME(3) NULL, last_error TEXT NULL, last_warning TEXT NULL,
  PRIMARY KEY (singleton_id), CONSTRAINT chk_official_pricing_sync_singleton CHECK (singleton_id=1)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
