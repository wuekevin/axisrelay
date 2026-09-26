-- User-facing API keys plus legacy Gateway API-key state.
CREATE TABLE IF NOT EXISTS user_api_keys (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, user_id BIGINT UNSIGNED NOT NULL, name VARCHAR(128) NOT NULL DEFAULT '',
  key_prefix VARCHAR(32) CHARACTER SET ascii COLLATE ascii_bin NOT NULL, key_hash CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active', quota_limit BIGINT UNSIGNED NOT NULL DEFAULT 0, quota_used BIGINT UNSIGNED NOT NULL DEFAULT 0,
  limits_json JSON NULL, last_used_at DATETIME(3) NULL, expires_at DATETIME(3) NULL, revoked_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id), UNIQUE KEY uk_user_api_keys_hash (key_hash), KEY idx_user_api_keys_user_status (user_id, status, id),
  CONSTRAINT fk_user_api_keys_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS api_keys (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, name VARCHAR(255) NOT NULL DEFAULT '', `key` VARCHAR(255) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  quota_limit BIGINT UNSIGNED NOT NULL DEFAULT 0, quota_used BIGINT UNSIGNED NOT NULL DEFAULT 0, total_used BIGINT UNSIGNED NOT NULL DEFAULT 0,
  reset_count BIGINT UNSIGNED NOT NULL DEFAULT 0, last_reset_at DATETIME(3) NULL, allowed_group_ids JSON NULL, expires_at DATETIME(3) NULL,
  enabled TINYINT(1) NOT NULL DEFAULT 1, limits JSON NULL, created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id), UNIQUE KEY uk_api_keys_key (`key`), KEY idx_api_keys_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS api_key_model_request_counters (
  api_key_id BIGINT UNSIGNED NOT NULL, rule_id VARCHAR(128) NOT NULL, window_start DATETIME(3) NOT NULL, reset_at DATETIME(3) NOT NULL,
  used_requests BIGINT UNSIGNED NOT NULL DEFAULT 0, PRIMARY KEY (api_key_id, rule_id, window_start), KEY idx_api_key_model_reset (reset_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS api_key_model_request_ledger (
  api_key_id BIGINT UNSIGNED NOT NULL, rule_id VARCHAR(128) NOT NULL, request_id VARCHAR(191) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  window_start DATETIME(3) NOT NULL, created_at DATETIME(3) NOT NULL, PRIMARY KEY (api_key_id, rule_id, request_id), KEY idx_api_key_model_ledger_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS api_key_scope_counters (
  api_key_id BIGINT UNSIGNED NOT NULL, scope_type VARCHAR(32) NOT NULL, scope_id BIGINT UNSIGNED NOT NULL,
  used_cost BIGINT UNSIGNED NOT NULL DEFAULT 0, used_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0, used_requests BIGINT UNSIGNED NOT NULL DEFAULT 0,
  reset_count BIGINT UNSIGNED NOT NULL DEFAULT 0, last_reset_at DATETIME(3) NULL, updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (api_key_id, scope_type, scope_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS api_key_auth_cache_state (
  id TINYINT UNSIGNED NOT NULL DEFAULT 1, namespace VARCHAR(191) NOT NULL, generation BIGINT UNSIGNED NOT NULL DEFAULT 1, key_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (id), CONSTRAINT chk_api_key_auth_cache_singleton CHECK (id=1)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS prompt_filter_newapi_bindings (
  api_key_id BIGINT UNSIGNED NOT NULL, platform_code VARCHAR(100) NOT NULL, platform_name VARCHAR(255) NOT NULL DEFAULT '', secret TEXT NOT NULL,
  enabled TINYINT(1) NOT NULL DEFAULT 1, require_signed_identity TINYINT(1) NOT NULL DEFAULT 0, prompt_filter_scope VARCHAR(32) NOT NULL DEFAULT 'inherit',
  policy_mode VARCHAR(32) NOT NULL DEFAULT 'inherit', policy_profile VARCHAR(64) NOT NULL DEFAULT 'inherit', previous_secret TEXT NULL, previous_secret_expires_at DATETIME(3) NULL,
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3), PRIMARY KEY (api_key_id), UNIQUE KEY uk_prompt_filter_binding_platform (platform_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
