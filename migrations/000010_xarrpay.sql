-- XArrPay integration. Secrets remain encrypted/opaque at the application layer.
CREATE TABLE IF NOT EXISTS xarrpay_configs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, merchant_code VARCHAR(128) NOT NULL, enabled TINYINT(1) NOT NULL DEFAULT 0,
  api_base_url VARCHAR(2048) NOT NULL DEFAULT '', merchant_id VARCHAR(191) NOT NULL DEFAULT '', credential_ciphertext LONGTEXT NULL,
  notify_url VARCHAR(2048) NOT NULL DEFAULT '', return_url VARCHAR(2048) NOT NULL DEFAULT '', config_json JSON NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id), UNIQUE KEY uk_xarrpay_configs_merchant_code (merchant_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS xarrpay_events (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, event_id VARCHAR(191) CHARACTER SET ascii COLLATE ascii_bin NOT NULL, merchant_code VARCHAR(128) NOT NULL,
  order_no VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '', payment_no VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '',
  event_type VARCHAR(64) NOT NULL DEFAULT '', signature_hash CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '', payload_json JSON NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'received', error_message TEXT NULL, received_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), processed_at DATETIME(3) NULL,
  PRIMARY KEY (id), UNIQUE KEY uk_xarrpay_events_event_id (event_id), KEY idx_xarrpay_events_order (order_no), KEY idx_xarrpay_events_status_received (status, received_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
