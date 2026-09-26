-- Usage and billing. All monetary values use integer micro-units/minor units.
CREATE TABLE IF NOT EXISTS usage_records (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, request_id VARCHAR(191) CHARACTER SET ascii COLLATE ascii_bin NOT NULL, user_id BIGINT UNSIGNED NULL,
  user_api_key_id BIGINT UNSIGNED NULL, provider_id BIGINT UNSIGNED NULL, provider_account_id BIGINT UNSIGNED NULL, model_id BIGINT UNSIGNED NULL,
  input_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0, output_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0, cached_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0,
  reasoning_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0, image_count BIGINT UNSIGNED NOT NULL DEFAULT 0, status_code INT NOT NULL DEFAULT 0, duration_ms BIGINT UNSIGNED NOT NULL DEFAULT 0,
  client_ip VARCHAR(45) NOT NULL DEFAULT '', metadata_json JSON NULL, created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id), KEY idx_usage_records_request (request_id), KEY idx_usage_records_user_created (user_id, created_at), KEY idx_usage_records_model_created (model_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS billing_records (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, request_id VARCHAR(191) CHARACTER SET ascii COLLATE ascii_bin NOT NULL, user_id BIGINT UNSIGNED NULL,
  usage_record_id BIGINT UNSIGNED NULL, wallet_id BIGINT UNSIGNED NULL, currency CHAR(3) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT 'USD',
  provider_cost BIGINT UNSIGNED NOT NULL DEFAULT 0, user_charge BIGINT UNSIGNED NOT NULL DEFAULT 0, margin BIGINT NOT NULL DEFAULT 0,
  billing_status VARCHAR(32) NOT NULL DEFAULT 'settled', pricing_snapshot_json JSON NULL, created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id), UNIQUE KEY uk_billing_records_request_id (request_id), KEY idx_billing_records_user_created (user_id, created_at),
  CONSTRAINT fk_billing_records_usage FOREIGN KEY (usage_record_id) REFERENCES usage_records(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS usage_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, account_id BIGINT UNSIGNED NOT NULL DEFAULT 0, credential_generation BIGINT UNSIGNED NOT NULL DEFAULT 0,
  client_ip VARCHAR(45) NOT NULL DEFAULT '', client_user_agent VARCHAR(1024) NOT NULL DEFAULT '', upstream_user_agent VARCHAR(1024) NOT NULL DEFAULT '',
  user_agent_overridden TINYINT(1) NOT NULL DEFAULT 0, turn_state_overridden TINYINT(1) NOT NULL DEFAULT 0, turn_state_rewrite_note VARCHAR(255) NOT NULL DEFAULT '',
  internal_reason VARCHAR(255) NOT NULL DEFAULT '', parent_request_id VARCHAR(191) NOT NULL DEFAULT '', endpoint VARCHAR(255) NOT NULL DEFAULT '', model VARCHAR(191) NOT NULL DEFAULT '',
  prompt_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0, completion_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0, total_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0,
  status_code INT NOT NULL DEFAULT 0, duration_ms BIGINT UNSIGNED NOT NULL DEFAULT 0, created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  input_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0, output_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0, reasoning_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0,
  first_token_ms BIGINT UNSIGNED NOT NULL DEFAULT 0, ws_acquire_ms BIGINT UNSIGNED NOT NULL DEFAULT 0, reasoning_effort VARCHAR(100) NOT NULL DEFAULT '',
  effective_model VARCHAR(191) NOT NULL DEFAULT '', upstream_response_model VARCHAR(191) NULL, upstream_model_mismatch TINYINT(1) NULL,
  inbound_endpoint VARCHAR(255) NOT NULL DEFAULT '', upstream_endpoint VARCHAR(255) NOT NULL DEFAULT '', stream TINYINT(1) NOT NULL DEFAULT 0, compact TINYINT(1) NOT NULL DEFAULT 0,
  has_compaction_history TINYINT(1) NOT NULL DEFAULT 0, ultra TINYINT(1) NOT NULL DEFAULT 0, via_websocket TINYINT(1) NOT NULL DEFAULT 0,
  cached_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0, image_input_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0, image_output_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0,
  cached_image_input_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0, cache_write_5m_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0, cache_write_1h_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0,
  service_tier VARCHAR(100) NOT NULL DEFAULT '', requested_service_tier VARCHAR(100) NOT NULL DEFAULT '', actual_service_tier VARCHAR(100) NOT NULL DEFAULT '', billing_service_tier VARCHAR(100) NOT NULL DEFAULT '',
  api_key_id BIGINT UNSIGNED NOT NULL DEFAULT 0, api_key_name VARCHAR(255) NOT NULL DEFAULT '', api_key_masked VARCHAR(64) NOT NULL DEFAULT '',
  image_count BIGINT UNSIGNED NOT NULL DEFAULT 0, image_width INT UNSIGNED NOT NULL DEFAULT 0, image_height INT UNSIGNED NOT NULL DEFAULT 0, image_bytes BIGINT UNSIGNED NOT NULL DEFAULT 0,
  image_format VARCHAR(32) NOT NULL DEFAULT '', image_size VARCHAR(32) NOT NULL DEFAULT '', error_message TEXT NULL, channel VARCHAR(32) NOT NULL DEFAULT '',
  request_id VARCHAR(191) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '', upstream_request_id VARCHAR(191) NOT NULL DEFAULT '', upstream_proxy_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  upstream_proxy_name VARCHAR(255) NOT NULL DEFAULT '', injected_turn_state LONGTEXT NULL, upstream_turn_state LONGTEXT NULL, user_billing_mode VARCHAR(32) NOT NULL DEFAULT '',
  image_unit_price BIGINT UNSIGNED NOT NULL DEFAULT 0, billed_image_count BIGINT UNSIGNED NOT NULL DEFAULT 0, account_billed BIGINT UNSIGNED NOT NULL DEFAULT 0, user_billed BIGINT UNSIGNED NOT NULL DEFAULT 0,
  is_retry_attempt TINYINT(1) NOT NULL DEFAULT 0, attempt_index INT UNSIGNED NOT NULL DEFAULT 0, upstream_error_kind VARCHAR(64) NOT NULL DEFAULT '', prompt_policy_incident_id VARCHAR(64) NULL,
  PRIMARY KEY (id), KEY idx_usage_logs_created_at (created_at), KEY idx_usage_logs_account_created_at (account_id, created_at),
  KEY idx_usage_logs_created_status (created_at, status_code), KEY idx_usage_logs_api_key_created_at (api_key_id, created_at), KEY idx_usage_logs_channel_created_at (channel, created_at),
  KEY idx_usage_logs_request_id (request_id), KEY idx_usage_logs_policy_incident (prompt_policy_incident_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS usage_stats_baseline (
  id TINYINT UNSIGNED NOT NULL DEFAULT 1, total_requests BIGINT UNSIGNED NOT NULL DEFAULT 0, total_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0,
  prompt_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0, completion_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0, cached_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0,
  cache_hit_requests BIGINT UNSIGNED NOT NULL DEFAULT 0, first_token_ms_sum BIGINT UNSIGNED NOT NULL DEFAULT 0, first_token_samples BIGINT UNSIGNED NOT NULL DEFAULT 0,
  account_billed BIGINT UNSIGNED NOT NULL DEFAULT 0, user_billed BIGINT UNSIGNED NOT NULL DEFAULT 0, PRIMARY KEY (id), CONSTRAINT chk_usage_stats_baseline_singleton CHECK (id=1)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS usage_stats_rollup (
  channel VARCHAR(32) NOT NULL, total_requests BIGINT UNSIGNED NOT NULL DEFAULT 0, total_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0,
  prompt_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0, completion_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0, cached_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0,
  cache_hit_requests BIGINT UNSIGNED NOT NULL DEFAULT 0, first_token_ms_sum BIGINT UNSIGNED NOT NULL DEFAULT 0, first_token_samples BIGINT UNSIGNED NOT NULL DEFAULT 0,
  account_billed BIGINT UNSIGNED NOT NULL DEFAULT 0, user_billed BIGINT UNSIGNED NOT NULL DEFAULT 0, PRIMARY KEY (channel)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS usage_stats_rollup_state (
  id TINYINT UNSIGNED NOT NULL DEFAULT 1, initialized TINYINT(1) NOT NULL DEFAULT 0, last_log_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  aggregation_version BIGINT UNSIGNED NOT NULL DEFAULT 1, updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id), CONSTRAINT chk_usage_stats_rollup_state_singleton CHECK (id=1)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS account_daily_usage (
  account_id BIGINT UNSIGNED NOT NULL, day DATE NOT NULL, credits BIGINT UNSIGNED NOT NULL DEFAULT 0, users BIGINT UNSIGNED NOT NULL DEFAULT 0,
  threads BIGINT UNSIGNED NOT NULL DEFAULT 0, turns BIGINT UNSIGNED NOT NULL DEFAULT 0, uncached_input_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0,
  cached_input_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0, output_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0, total_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0,
  settled TINYINT(1) NOT NULL DEFAULT 0, clients_json JSON NOT NULL, models_json JSON NOT NULL, synced_at DATETIME(3) NOT NULL,
  breakdown_percent DECIMAL(8,4) NOT NULL DEFAULT 0, breakdown_json JSON NOT NULL, surfaces_json JSON NOT NULL,
  PRIMARY KEY (account_id, day), KEY idx_account_daily_usage_day (day)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
