CREATE INDEX idx_usage_logs_account_generation_created_at
  ON usage_logs(account_id, credential_generation, created_at);
