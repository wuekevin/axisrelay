-- Audit, prompt safety, identity-risk and proxy-risk schema.
CREATE TABLE IF NOT EXISTS audit_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, actor_type VARCHAR(32) NOT NULL, actor_id BIGINT UNSIGNED NULL, action VARCHAR(128) NOT NULL,
  resource_type VARCHAR(64) NOT NULL DEFAULT '', resource_id VARCHAR(191) NOT NULL DEFAULT '', client_ip VARCHAR(45) NOT NULL DEFAULT '',
  request_id VARCHAR(191) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '', before_json JSON NULL, after_json JSON NULL, metadata_json JSON NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), PRIMARY KEY (id), KEY idx_audit_logs_actor_created (actor_type, actor_id, created_at),
  KEY idx_audit_logs_resource_created (resource_type, resource_id, created_at), KEY idx_audit_logs_request (request_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS prompt_conversation_locks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, lock_key CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL, status VARCHAR(24) NOT NULL DEFAULT 'active',
  identity_kind VARCHAR(24) NOT NULL DEFAULT 'newapi', platform VARCHAR(100) NOT NULL DEFAULT '', newapi_user_id VARCHAR(255) NOT NULL DEFAULT '',
  session_fingerprint VARCHAR(32) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '', session_hash CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '',
  incident_id VARCHAR(64) NOT NULL DEFAULT '', decision_id VARCHAR(128) NOT NULL DEFAULT '', request_id VARCHAR(255) NOT NULL DEFAULT '', reason_code VARCHAR(100) NOT NULL DEFAULT '',
  endpoint VARCHAR(255) NOT NULL DEFAULT '', model VARCHAR(128) NOT NULL DEFAULT '', trigger_count BIGINT UNSIGNED NOT NULL DEFAULT 1, unlock_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
  locked_at DATETIME(3) NOT NULL, unlocked_at DATETIME(3) NULL, unlock_reason TEXT NULL, created_at DATETIME(3) NOT NULL, updated_at DATETIME(3) NOT NULL,
  PRIMARY KEY (id), UNIQUE KEY uk_prompt_conversation_locks_key (lock_key), KEY idx_prompt_conversation_locks_status (status, updated_at),
  KEY idx_prompt_conversation_locks_session (session_hash, status), KEY idx_prompt_conversation_locks_user_cooldown (platform, newapi_user_id, status, locked_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS prompt_filter_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), source VARCHAR(64) NOT NULL DEFAULT '',
  endpoint VARCHAR(255) NOT NULL DEFAULT '', request_protocol VARCHAR(64) NOT NULL DEFAULT '', request_provider VARCHAR(64) NOT NULL DEFAULT '', model VARCHAR(191) NOT NULL DEFAULT '',
  action VARCHAR(32) NOT NULL DEFAULT '', mode VARCHAR(32) NOT NULL DEFAULT '', score INT NOT NULL DEFAULT 0, audit_score INT NOT NULL DEFAULT 0, threshold_value INT NOT NULL DEFAULT 0,
  policy_profile VARCHAR(64) NOT NULL DEFAULT '', reason_code VARCHAR(100) NOT NULL DEFAULT '', primary_origin VARCHAR(64) NOT NULL DEFAULT '', strike_eligible TINYINT(1) NOT NULL DEFAULT 0,
  matched_patterns JSON NULL, text_preview TEXT NULL, match_context TEXT NULL, api_key_id BIGINT UNSIGNED NOT NULL DEFAULT 0, api_key_name VARCHAR(255) NOT NULL DEFAULT '',
  api_key_masked VARCHAR(64) NOT NULL DEFAULT '', client_ip VARCHAR(45) NOT NULL DEFAULT '', error_code VARCHAR(128) NOT NULL DEFAULT '', review_model VARCHAR(191) NOT NULL DEFAULT '',
  review_flagged TINYINT(1) NOT NULL DEFAULT 0, review_error TEXT NULL, reviewed TINYINT(1) NOT NULL DEFAULT 0, review_confidence DECIMAL(8,6) NULL, review_threshold DECIMAL(8,6) NULL,
  review_reason TEXT NULL, review_endpoint VARCHAR(255) NOT NULL DEFAULT '', review_request_mode VARCHAR(64) NOT NULL DEFAULT '', review_latency_ms BIGINT UNSIGNED NULL, full_text LONGTEXT NULL,
  request_correlation_id VARCHAR(191) NOT NULL DEFAULT '', newapi_policy_status VARCHAR(32) NOT NULL DEFAULT '', newapi_platform VARCHAR(100) NOT NULL DEFAULT '',
  newapi_user_id VARCHAR(255) NOT NULL DEFAULT '', newapi_request_id VARCHAR(191) NOT NULL DEFAULT '', newapi_decision_id VARCHAR(128) NOT NULL DEFAULT '',
  session_hash CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '', PRIMARY KEY (id),
  KEY idx_prompt_filter_logs_created_at (created_at), KEY idx_prompt_filter_logs_action_created_at (action, created_at), KEY idx_prompt_filter_logs_source_id (source, id),
  KEY idx_prompt_filter_logs_reviewed_id (reviewed, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS prompt_log_retention_config (
  singleton_id TINYINT UNSIGNED NOT NULL DEFAULT 1, retention_days INT UNSIGNED NOT NULL DEFAULT 7, last_run_at DATETIME(3) NULL,
  last_deleted_logs BIGINT UNSIGNED NOT NULL DEFAULT 0, last_deleted_events BIGINT UNSIGNED NOT NULL DEFAULT 0, last_deleted_sources BIGINT UNSIGNED NOT NULL DEFAULT 0,
  last_duration_ms BIGINT UNSIGNED NOT NULL DEFAULT 0, last_error TEXT NULL, PRIMARY KEY (singleton_id), CONSTRAINT chk_prompt_log_retention_singleton CHECK (singleton_id=1)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS prompt_policy_incidents (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, incident_id VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL, request_correlation_id VARCHAR(191) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), attempt_index INT UNSIGNED NOT NULL DEFAULT 0, transport VARCHAR(32) NOT NULL DEFAULT '', endpoint VARCHAR(255) NOT NULL DEFAULT '',
  request_protocol VARCHAR(64) NOT NULL DEFAULT '', request_provider VARCHAR(64) NOT NULL DEFAULT '', model VARCHAR(191) NOT NULL DEFAULT '', status_code INT NOT NULL DEFAULT 0,
  account_id BIGINT UNSIGNED NOT NULL DEFAULT 0, account_name VARCHAR(255) NOT NULL DEFAULT '', account_platform VARCHAR(64) NOT NULL DEFAULT '',
  account_group_ids JSON NULL, account_group_names JSON NULL, api_key_id BIGINT UNSIGNED NOT NULL DEFAULT 0, api_key_name VARCHAR(255) NOT NULL DEFAULT '',
  api_key_masked VARCHAR(64) NOT NULL DEFAULT '', api_key_allowed_group_ids JSON NULL, api_key_allowed_group_names JSON NULL, platform VARCHAR(100) NOT NULL DEFAULT '',
  newapi_policy_status VARCHAR(32) NOT NULL DEFAULT '', newapi_platform VARCHAR(100) NOT NULL DEFAULT '', newapi_user_id VARCHAR(255) NOT NULL DEFAULT '',
  newapi_request_id VARCHAR(191) NOT NULL DEFAULT '', session_hash CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '',
  client_ip_hash CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '', source_ref VARCHAR(255) NOT NULL DEFAULT '', upstream_error_code VARCHAR(128) NOT NULL DEFAULT '',
  upstream_error TEXT NULL, local_evaluation_state VARCHAR(32) NOT NULL DEFAULT '', local_outcome VARCHAR(32) NOT NULL DEFAULT '', local_action VARCHAR(32) NOT NULL DEFAULT '',
  local_score INT NULL, local_raw_score INT NULL, local_audit_score INT NULL, local_audit_raw_score INT NULL, local_threshold INT NOT NULL DEFAULT 0,
  local_mode VARCHAR(32) NOT NULL DEFAULT '', local_policy_profile VARCHAR(64) NOT NULL DEFAULT '', local_reason_code VARCHAR(100) NOT NULL DEFAULT '', local_reason TEXT NULL,
  local_primary_origin VARCHAR(64) NOT NULL DEFAULT '', local_strike_eligible TINYINT(1) NOT NULL DEFAULT 0, local_review_model VARCHAR(191) NOT NULL DEFAULT '',
  local_review_flagged TINYINT(1) NOT NULL DEFAULT 0, local_review_error TEXT NULL, local_matched_patterns JSON NULL,
  prompt_fingerprint CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '', prompt_preview TEXT NULL, prompt_text LONGTEXT NULL, prompt_available TINYINT(1) NOT NULL DEFAULT 0,
  local_comparison VARCHAR(32) NOT NULL DEFAULT '', candidate_id BIGINT UNSIGNED NOT NULL DEFAULT 0, candidate_evidence_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (id), UNIQUE KEY uk_prompt_policy_incidents_incident_id (incident_id), KEY idx_prompt_policy_incidents_request (request_correlation_id, created_at),
  KEY idx_prompt_policy_incidents_created (created_at), KEY idx_prompt_policy_incidents_api_key (api_key_id, created_at), KEY idx_prompt_policy_incidents_account (account_id, created_at),
  KEY idx_prompt_policy_incidents_endpoint (endpoint, created_at), KEY idx_prompt_policy_incidents_outcome (local_outcome, created_at), KEY idx_prompt_policy_incidents_comparison (local_comparison, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS prompt_review_profiles (
  id VARCHAR(191) NOT NULL, name VARCHAR(255) NOT NULL, base_url TEXT NULL, model VARCHAR(191) NOT NULL DEFAULT '', request_mode VARCHAR(64) NOT NULL DEFAULT 'moderations',
  adapter_json JSON NOT NULL, api_keys LONGTEXT NULL, timeout_seconds INT UNSIGNED NOT NULL DEFAULT 10, active TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3), PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS prompt_risk_event_sources (
  source_type VARCHAR(32) NOT NULL, source_id VARCHAR(96) NOT NULL, processed_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (source_type, source_id), KEY idx_prompt_risk_sources_processed (processed_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS prompt_risk_events (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), source_type VARCHAR(32) NOT NULL, source_id VARCHAR(96) NOT NULL,
  incident_id VARCHAR(64) NOT NULL DEFAULT '', prompt_filter_log_id BIGINT UNSIGNED NOT NULL DEFAULT 0, request_correlation_id VARCHAR(64) NOT NULL DEFAULT '',
  subject_type VARCHAR(32) NOT NULL, subject_key VARCHAR(160) NOT NULL, subject_display VARCHAR(255) NOT NULL DEFAULT '', platform VARCHAR(100) NOT NULL DEFAULT '',
  is_person TINYINT(1) NOT NULL DEFAULT 0, identity_confidence INT NOT NULL DEFAULT 0, event_kind VARCHAR(64) NOT NULL, request_risk_score INT NOT NULL DEFAULT 0,
  evidence_confidence INT NOT NULL DEFAULT 0, reason_code VARCHAR(100) NOT NULL DEFAULT '', action VARCHAR(32) NOT NULL DEFAULT '', local_outcome VARCHAR(32) NOT NULL DEFAULT '',
  local_comparison VARCHAR(32) NOT NULL DEFAULT '', endpoint VARCHAR(256) NOT NULL DEFAULT '', model VARCHAR(100) NOT NULL DEFAULT '',
  prompt_fingerprint CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '', prompt_preview TEXT NULL, api_key_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  api_key_name VARCHAR(255) NOT NULL DEFAULT '', api_key_masked VARCHAR(64) NOT NULL DEFAULT '', account_id BIGINT UNSIGNED NOT NULL DEFAULT 0, account_name VARCHAR(255) NOT NULL DEFAULT '',
  PRIMARY KEY (id), UNIQUE KEY uk_prompt_risk_events_source_subject (source_type, source_id, subject_type, subject_key),
  KEY idx_prompt_risk_events_subject (subject_type, subject_key, created_at), KEY idx_prompt_risk_events_created (created_at), KEY idx_prompt_risk_events_kind (event_kind, created_at),
  KEY idx_prompt_risk_events_incident (incident_id), KEY idx_prompt_risk_events_api_key (api_key_id, created_at), KEY idx_prompt_risk_events_account (account_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS prompt_risk_identities (
  subject_type VARCHAR(32) NOT NULL, subject_key VARCHAR(160) NOT NULL, platform VARCHAR(100) NOT NULL DEFAULT '', external_user_id VARCHAR(255) NOT NULL DEFAULT '',
  user_name VARCHAR(128) NOT NULL DEFAULT '', user_email VARCHAR(320) NOT NULL DEFAULT '', user_group VARCHAR(100) NOT NULL DEFAULT '', source VARCHAR(32) NOT NULL DEFAULT '',
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3), PRIMARY KEY (subject_type, subject_key),
  KEY idx_prompt_risk_identities_external (platform, external_user_id), KEY idx_prompt_risk_identities_updated (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS prompt_risk_trust_policies (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, subject_type VARCHAR(40) NOT NULL, subject_key VARCHAR(128) NOT NULL, status VARCHAR(24) NOT NULL DEFAULT 'active',
  source VARCHAR(24) NOT NULL DEFAULT 'manual', reason TEXT NULL, risk_threshold INT NOT NULL DEFAULT 35, valid_until DATETIME(3) NOT NULL,
  last_evaluated_at DATETIME(3) NULL, last_risk_score INT NOT NULL DEFAULT 0, last_risk_level VARCHAR(24) NOT NULL DEFAULT 'low', bypass_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
  last_bypass_at DATETIME(3) NULL, model_review_count BIGINT UNSIGNED NOT NULL DEFAULT 0, last_model_review_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id), UNIQUE KEY uk_prompt_risk_trust_subject_key (subject_key), KEY idx_prompt_risk_trust_status_until (status, valid_until)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS prompt_risk_trust_events (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, policy_id BIGINT UNSIGNED NOT NULL, subject_type VARCHAR(40) NOT NULL, subject_key VARCHAR(128) NOT NULL,
  event_type VARCHAR(40) NOT NULL, reason TEXT NULL, risk_score INT NOT NULL DEFAULT 0, risk_level VARCHAR(24) NOT NULL DEFAULT '',
  request_id_hash CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '', created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id), KEY idx_prompt_risk_trust_events_policy (policy_id, created_at), KEY idx_prompt_risk_trust_events_subject (subject_type, subject_key, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS prompt_rule_candidates (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, fingerprint CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL, kind VARCHAR(32) NOT NULL DEFAULT 'pattern',
  status VARCHAR(32) NOT NULL DEFAULT 'pending', last_source VARCHAR(64) NOT NULL DEFAULT '', name VARCHAR(255) NOT NULL DEFAULT '', category VARCHAR(64) NOT NULL DEFAULT '',
  rule_json JSON NOT NULL, rationale TEXT NULL, source_url TEXT NULL, evidence_count BIGINT UNSIGNED NOT NULL DEFAULT 0, sample_preview TEXT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  last_seen_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), published_at DATETIME(3) NULL, dismissed_at DATETIME(3) NULL,
  PRIMARY KEY (id), UNIQUE KEY uk_prompt_rule_candidates_fingerprint (fingerprint), KEY idx_prompt_rule_candidates_status_seen (status, last_seen_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS prompt_rule_candidate_evidence (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, candidate_id BIGINT UNSIGNED NOT NULL, source_kind VARCHAR(64) NOT NULL DEFAULT '', source_ref TEXT NULL,
  source_ref_hash CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL, sample_preview TEXT NULL, metadata_json JSON NOT NULL, request_protocol VARCHAR(64) NOT NULL DEFAULT '',
  request_provider VARCHAR(64) NOT NULL DEFAULT '', model VARCHAR(191) NOT NULL DEFAULT '', api_key_id BIGINT UNSIGNED NOT NULL DEFAULT 0, api_key_name VARCHAR(255) NOT NULL DEFAULT '',
  prompt_policy_incident_id VARCHAR(64) NULL, observed_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id), UNIQUE KEY uk_prompt_rule_evidence_source (candidate_id, source_kind, source_ref_hash), KEY idx_prompt_rule_evidence_incident (prompt_policy_incident_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS proxy_risk_scoring_profiles (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, name VARCHAR(255) NOT NULL, provider VARCHAR(64) NOT NULL DEFAULT 'scamalytics', enabled TINYINT(1) NOT NULL DEFAULT 0,
  priority INT NOT NULL DEFAULT 0, base_url TEXT NULL, access_token LONGTEXT NULL, scamalytics_host VARCHAR(255) NOT NULL DEFAULT '', scamalytics_user VARCHAR(255) NOT NULL DEFAULT '',
  scamalytics_key LONGTEXT NULL, timeout_seconds INT UNSIGNED NOT NULL DEFAULT 8, concurrency INT UNSIGNED NOT NULL DEFAULT 3, request_delay_ms BIGINT UNSIGNED NOT NULL DEFAULT 0,
  cache_ttl_seconds BIGINT UNSIGNED NOT NULL DEFAULT 3600, max_checks_per_job BIGINT UNSIGNED NOT NULL DEFAULT 0, daily_check_limit BIGINT UNSIGNED NOT NULL DEFAULT 0,
  credit_reserve BIGINT UNSIGNED NOT NULL DEFAULT 0, allow_force_refresh TINYINT(1) NOT NULL DEFAULT 0, resolve_hostnames TINYINT(1) NOT NULL DEFAULT 0,
  allow_private_targets TINYINT(1) NOT NULL DEFAULT 0, docs_url TEXT NULL, tutorial_url TEXT NULL, daily_used_date DATE NULL, daily_used_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
  credits_remaining BIGINT NULL, credits_used BIGINT NULL, credit_reset_at DATETIME(3) NULL, last_quota_checked_at DATETIME(3) NULL, last_error TEXT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id), UNIQUE KEY uk_proxy_risk_profiles_name (name), KEY idx_proxy_risk_profile_enabled (enabled, priority, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS proxy_risk_score_snapshots (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, proxy_id BIGINT UNSIGNED NOT NULL, profile_id BIGINT UNSIGNED NOT NULL, provider VARCHAR(64) NOT NULL DEFAULT '',
  resolved_ip VARCHAR(45) NOT NULL DEFAULT '', score INT NULL, risk_level VARCHAR(32) NOT NULL DEFAULT '', recommendation TEXT NULL, proxy_type VARCHAR(64) NOT NULL DEFAULT '',
  is_vpn TINYINT(1) NOT NULL DEFAULT 0, is_tor TINYINT(1) NOT NULL DEFAULT 0, is_datacenter TINYINT(1) NOT NULL DEFAULT 0, is_blacklisted TINYINT(1) NOT NULL DEFAULT 0,
  blacklist_sources JSON NULL, isp VARCHAR(255) NOT NULL DEFAULT '', country VARCHAR(128) NOT NULL DEFAULT '', latency_ms BIGINT UNSIGNED NOT NULL DEFAULT 0,
  status VARCHAR(32) NOT NULL DEFAULT 'error', error TEXT NULL, features_json JSON NULL, raw_response_json JSON NULL, checked_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), expires_at DATETIME(3) NULL,
  PRIMARY KEY (id), KEY idx_proxy_risk_snapshot_latest (proxy_id, profile_id, checked_at, id), KEY idx_proxy_risk_snapshot_profile_time (profile_id, checked_at, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
