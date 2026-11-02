-- admin_users already exists from 000003_admin_rbac.sql. Extend it in place so
-- installations with existing RBAC rows keep their IDs, roles and passwords.
ALTER TABLE admin_users
    ADD COLUMN username_normalized VARCHAR(64) NULL AFTER username,
    ADD COLUMN must_change_password TINYINT(1) NOT NULL DEFAULT 0 AFTER status;

UPDATE admin_users
SET username_normalized = LOWER(TRIM(username))
WHERE username_normalized IS NULL OR username_normalized = '';

ALTER TABLE admin_users
    MODIFY COLUMN username_normalized VARCHAR(64) NOT NULL,
    MODIFY COLUMN email VARCHAR(320) NULL,
    ADD UNIQUE KEY uk_admin_users_username_normalized (username_normalized);

CREATE TABLE IF NOT EXISTS admin_sessions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    admin_user_id BIGINT UNSIGNED NOT NULL,
    token_hash CHAR(64) NOT NULL,
    ip_address VARCHAR(64) NOT NULL DEFAULT '',
    user_agent VARCHAR(512) NOT NULL DEFAULT '',
    expires_at DATETIME(6) NOT NULL,
    revoked_at DATETIME(6) NULL,
    last_seen_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uk_admin_sessions_token_hash (token_hash),
    KEY idx_admin_sessions_user_expiry (admin_user_id, expires_at),
    CONSTRAINT fk_admin_sessions_user FOREIGN KEY (admin_user_id) REFERENCES admin_users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
