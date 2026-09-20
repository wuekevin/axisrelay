-- User/admin notifications.
CREATE TABLE IF NOT EXISTS notifications (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, public_id CHAR(26) CHARACTER SET ascii COLLATE ascii_bin NOT NULL, user_id BIGINT UNSIGNED NULL,
  audience VARCHAR(32) NOT NULL DEFAULT 'user', notification_type VARCHAR(64) NOT NULL, title VARCHAR(255) NOT NULL, body TEXT NOT NULL,
  data_json JSON NULL, status VARCHAR(32) NOT NULL DEFAULT 'unread', read_at DATETIME(3) NULL, expires_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3), PRIMARY KEY (id), UNIQUE KEY uk_notifications_public_id (public_id),
  KEY idx_notifications_user_status (user_id, status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS notification_deliveries (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, notification_id BIGINT UNSIGNED NOT NULL, channel VARCHAR(32) NOT NULL, destination VARCHAR(512) NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL DEFAULT 'pending', attempt_count INT UNSIGNED NOT NULL DEFAULT 0, provider_message_id VARCHAR(191) NOT NULL DEFAULT '',
  last_error TEXT NULL, next_attempt_at DATETIME(3) NULL, sent_at DATETIME(3) NULL, created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3), PRIMARY KEY (id),
  KEY idx_notification_deliveries_status (status, next_attempt_at, id), KEY idx_notification_deliveries_notification (notification_id),
  CONSTRAINT fk_notification_deliveries_notification FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
