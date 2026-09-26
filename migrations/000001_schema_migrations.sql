-- AxisRelay S0.5 / migration metadata
CREATE TABLE IF NOT EXISTS schema_migrations (
  version BIGINT UNSIGNED NOT NULL,
  name VARCHAR(255) NOT NULL,
  checksum CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  applied_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (version),
  UNIQUE KEY uk_schema_migrations_name (name)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
