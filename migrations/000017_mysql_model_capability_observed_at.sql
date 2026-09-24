ALTER TABLE model_capability_snapshots
  ADD COLUMN observed_at_ns BIGINT UNSIGNED NULL;

UPDATE model_capability_snapshots
SET observed_at_ns = CAST(UNIX_TIMESTAMP(observed_at) * 1000000000 AS UNSIGNED)
WHERE observed_at_ns IS NULL;

ALTER TABLE model_capability_snapshots
  DROP COLUMN observed_at,
  CHANGE COLUMN observed_at_ns observed_at BIGINT UNSIGNED NOT NULL;
