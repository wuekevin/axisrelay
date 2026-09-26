-- Legacy Gateway singleton settings retained for S0.6 query migration.
CREATE TABLE IF NOT EXISTS system_settings (
  id TINYINT UNSIGNED NOT NULL DEFAULT 1,
  site_name VARCHAR(255) NOT NULL DEFAULT 'AxisRelay',
  site_logo TEXT NULL,
  background_config JSON NULL, grok_config JSON NULL, claude_config JSON NULL, antigravity_oauth_config JSON NULL,
  invite_guide_config JSON NULL, visible_channels_config JSON NULL, channel_test_config JSON NULL, antigravity_config JSON NULL,
  max_concurrency INT UNSIGNED NOT NULL DEFAULT 2, global_rpm BIGINT UNSIGNED NOT NULL DEFAULT 0, test_model VARCHAR(191) NOT NULL DEFAULT 'gpt-5.5',
  test_content TEXT NULL, test_concurrency INT UNSIGNED NOT NULL DEFAULT 50, proxy_url TEXT NULL,
  pg_max_conns INT UNSIGNED NOT NULL DEFAULT 50, redis_pool_size INT UNSIGNED NOT NULL DEFAULT 30,
  auto_clean_unauthorized TINYINT(1) NOT NULL DEFAULT 0, auto_clean_rate_limited TINYINT(1) NOT NULL DEFAULT 0,
  background_refresh_interval_minutes INT UNSIGNED NOT NULL DEFAULT 2, usage_probe_max_age_minutes INT UNSIGNED NOT NULL DEFAULT 10,
  usage_probe_concurrency INT UNSIGNED NOT NULL DEFAULT 16, usage_probe_responses_fallback_enabled TINYINT(1) NOT NULL DEFAULT 1,
  recovery_probe_interval_minutes INT UNSIGNED NOT NULL DEFAULT 30, admin_secret TEXT NULL, auto_clean_full_usage TINYINT(1) NOT NULL DEFAULT 0,
  auto_clean_error TINYINT(1) NOT NULL DEFAULT 0, auto_clean_expired TINYINT(1) NOT NULL DEFAULT 0, lazy_mode TINYINT(1) NOT NULL DEFAULT 0,
  proxy_pool_enabled TINYINT(1) NOT NULL DEFAULT 0, fast_scheduler_enabled TINYINT(1) NOT NULL DEFAULT 0, scheduler_engine VARCHAR(32) NOT NULL DEFAULT '',
  max_retries INT UNSIGNED NOT NULL DEFAULT 2, max_rate_limit_retries INT UNSIGNED NOT NULL DEFAULT 1, reasoning_effort_models JSON NULL,
  allow_remote_migration TINYINT(1) NOT NULL DEFAULT 0, client_compat_mode VARCHAR(32) NOT NULL DEFAULT 'preserve',
  codex_min_cli_version VARCHAR(64) NOT NULL DEFAULT '0.153.3', codex_user_agent_config JSON NULL, codex_images_main_model VARCHAR(191) NOT NULL DEFAULT '',
  usage_log_mode VARCHAR(32) NOT NULL DEFAULT 'full', usage_log_batch_size INT UNSIGNED NOT NULL DEFAULT 200, usage_log_flush_interval_seconds INT UNSIGNED NOT NULL DEFAULT 5,
  stream_flush_policy VARCHAR(32) NOT NULL DEFAULT 'immediate', stream_flush_interval_ms BIGINT UNSIGNED NOT NULL DEFAULT 20,
  first_token_mode VARCHAR(32) NOT NULL DEFAULT 'loose', first_token_timeout_seconds INT UNSIGNED NOT NULL DEFAULT 0, image_storage_config JSON NULL,
  show_full_usage_numbers TINYINT(1) NOT NULL DEFAULT 0, public_key_usage_page_enabled TINYINT(1) NOT NULL DEFAULT 1,
  public_image_studio_page_enabled TINYINT(1) NOT NULL DEFAULT 1, public_account_portal_page_enabled TINYINT(1) NOT NULL DEFAULT 0,
  scheduler_mode VARCHAR(32) NOT NULL DEFAULT 'round_robin', affinity_mode VARCHAR(32) NOT NULL DEFAULT 'bounded', session_affinity_spread INT UNSIGNED NOT NULL DEFAULT 0,
  session_slot_buffer_enabled TINYINT(1) NOT NULL DEFAULT 0, session_slot_buffer_seconds INT UNSIGNED NOT NULL DEFAULT 10,
  models_list_read_max_bytes BIGINT UNSIGNED NOT NULL DEFAULT 8388608, codex_force_websocket TINYINT(1) NOT NULL DEFAULT 0,
  codex_telemetry_enabled TINYINT(1) NOT NULL DEFAULT 0, codex_turn_state_template_cache_enabled TINYINT(1) NOT NULL DEFAULT 0,
  codex_turn_state_account_mode VARCHAR(32) NOT NULL DEFAULT 'auto', codex_telemetry_timing_debug TINYINT(1) NOT NULL DEFAULT 0,
  codex_request_compression TINYINT(1) NOT NULL DEFAULT 1, codex_ws_weak_network_mode TINYINT(1) NOT NULL DEFAULT 0,
  codex_ws_keepalive_enabled TINYINT(1) NOT NULL DEFAULT 0, codex_ws_keepalive_interval_sec INT UNSIGNED NOT NULL DEFAULT 60,
  codex_ws_hide_upstream_errors TINYINT(1) NOT NULL DEFAULT 1, codex_ws_silent_retry_enabled TINYINT(1) NOT NULL DEFAULT 1,
  codex_ws_silent_max_retries INT UNSIGNED NOT NULL DEFAULT 2, codex_ws_size_router_enabled TINYINT(1) NOT NULL DEFAULT 1,
  codex_ws_busy_acquire_max_wait_sec INT UNSIGNED NOT NULL DEFAULT 30, codex_ws_busy_overflow_enabled TINYINT(1) NOT NULL DEFAULT 0,
  codex_ws_busy_patience_sec INT UNSIGNED NOT NULL DEFAULT 2, codex_ws_stateless_slots INT UNSIGNED NOT NULL DEFAULT 8,
  github_token TEXT NULL, github_proxy_url TEXT NULL, codex_overload_pause_enabled TINYINT(1) NOT NULL DEFAULT 0,
  codex_overload_threshold_percent INT UNSIGNED NOT NULL DEFAULT 20, codex_overload_pause_minutes INT UNSIGNED NOT NULL DEFAULT 30,
  codex_overload_window_minutes INT UNSIGNED NOT NULL DEFAULT 5, overflow_auto_compact_enabled TINYINT(1) NOT NULL DEFAULT 0,
  compact_via_responses_enabled TINYINT(1) NOT NULL DEFAULT 0, codex_preflight_sse_passthrough_enabled TINYINT(1) NOT NULL DEFAULT 0,
  first_token_excludes_ws_acquire TINYINT(1) NOT NULL DEFAULT 0, codex_continue_thinking_enabled TINYINT(1) NOT NULL DEFAULT 0,
  codex_continue_max_rounds INT UNSIGNED NOT NULL DEFAULT 8, retry_interval_ms BIGINT UNSIGNED NOT NULL DEFAULT 0,
  transport_retry_policy VARCHAR(32) NOT NULL DEFAULT 'rotate', continuous_retry_policy JSON NULL, codex_synced_cli_version VARCHAR(64) NOT NULL DEFAULT '',
  codex_cli_version_sync_enabled TINYINT(1) NOT NULL DEFAULT 1, codex_cli_version_sync_interval_hours INT UNSIGNED NOT NULL DEFAULT 12,
  claude_synced_cli_version VARCHAR(64) NOT NULL DEFAULT '', model_pricing_overrides JSON NULL, model_pricing_sync_url TEXT NULL,
  ignore_usage_limit_status TINYINT(1) NOT NULL DEFAULT 0, auto_reset_credits_enabled TINYINT(1) NOT NULL DEFAULT 0,
  auto_reset_credits_before_expiry_min INT UNSIGNED NOT NULL DEFAULT 60, auto_activate_5h_window_enabled TINYINT(1) NOT NULL DEFAULT 0,
  utls_shutdown_timeout_minutes INT UNSIGNED NOT NULL DEFAULT 30, codex_fingerprint_default_mode VARCHAR(32) NOT NULL DEFAULT 'off',
  response_cache_local_max_bytes BIGINT UNSIGNED NOT NULL DEFAULT 67108864, response_cache_local_max_entry_bytes BIGINT UNSIGNED NOT NULL DEFAULT 8388608,
  response_cache_reconstruct_max_bytes BIGINT UNSIGNED NOT NULL DEFAULT 67108864, response_cache_write_policy VARCHAR(32) NOT NULL DEFAULT 'always',
  response_cache_config_generation BIGINT UNSIGNED NOT NULL DEFAULT 1, relay_model_cooldown_mode VARCHAR(32) NOT NULL DEFAULT 'off',
  relay_model_cooldown_seconds INT UNSIGNED NOT NULL DEFAULT 2, relay_model_cooldown_backoff_enabled TINYINT(1) NOT NULL DEFAULT 0,
  oauth_model_cooldown_mode VARCHAR(32) NOT NULL DEFAULT 'adaptive', oauth_model_cooldown_seconds INT UNSIGNED NOT NULL DEFAULT 300,
  oauth_model_cooldown_backoff_enabled TINYINT(1) NOT NULL DEFAULT 1, codex_oauth_keepalive_enabled TINYINT(1) NOT NULL DEFAULT 0,
  model_mapping JSON NULL, codex_model_mapping JSON NULL, payload_rules JSON NULL, resin_url TEXT NULL, resin_platform_name VARCHAR(128) NOT NULL DEFAULT '',
  prompt_filter_enabled TINYINT(1) NOT NULL DEFAULT 0, prompt_filter_mode VARCHAR(32) NOT NULL DEFAULT 'monitor', prompt_filter_threshold INT NOT NULL DEFAULT 50,
  prompt_filter_strict_threshold INT NOT NULL DEFAULT 90, prompt_filter_strict_terminal_enabled TINYINT(1) NOT NULL DEFAULT 0,
  prompt_filter_advanced_config JSON NULL, prompt_filter_log_matches TINYINT(1) NOT NULL DEFAULT 1, prompt_filter_max_text_length BIGINT UNSIGNED NOT NULL DEFAULT 81920,
  prompt_filter_sensitive_words LONGTEXT NULL, prompt_filter_custom_patterns JSON NULL, prompt_filter_disabled_patterns JSON NULL,
  prompt_filter_review_enabled TINYINT(1) NOT NULL DEFAULT 0, prompt_filter_review_api_key LONGTEXT NULL,
  prompt_filter_review_base_url TEXT NULL, prompt_filter_review_model VARCHAR(191) NOT NULL DEFAULT 'deepseek-v4-flash',
  prompt_filter_review_timeout_seconds INT UNSIGNED NOT NULL DEFAULT 10, prompt_filter_review_fail_closed TINYINT(1) NOT NULL DEFAULT 1,
  billing_tier_policy VARCHAR(32) NOT NULL DEFAULT 'actual', auto_pause_5h_threshold DECIMAL(8,4) NOT NULL DEFAULT 0,
  auto_pause_7d_threshold DECIMAL(8,4) NOT NULL DEFAULT 0, auto_pause_5h_guard_band_percent DECIMAL(8,4) NOT NULL DEFAULT 5,
  auto_pause_5h_guard_concurrency INT UNSIGNED NOT NULL DEFAULT 1, smart_pacing_enabled TINYINT(1) NOT NULL DEFAULT 0,
  smart_pacing_min_concurrency INT UNSIGNED NOT NULL DEFAULT 1, smart_pacing_windows VARCHAR(255) NOT NULL DEFAULT '5h,7d',
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id), CONSTRAINT chk_system_settings_singleton CHECK (id=1)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS data_migrations (
  version VARCHAR(191) NOT NULL, applied_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), PRIMARY KEY (version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT IGNORE INTO system_settings (id) VALUES (1);
INSERT IGNORE INTO official_pricing_sync_config (singleton_id) VALUES (1);
INSERT IGNORE INTO prompt_log_retention_config (singleton_id) VALUES (1);
INSERT IGNORE INTO api_key_auth_cache_state (id, namespace) VALUES (1, 'axisrelay');
INSERT IGNORE INTO usage_stats_baseline (id) VALUES (1);
INSERT IGNORE INTO usage_stats_rollup_state (id) VALUES (1);
INSERT IGNORE INTO model_registry_sync (id) VALUES (1);
