-- Link self-service user keys to the gateway keys that authorize /v1 requests.
ALTER TABLE user_api_keys
  ADD COLUMN gateway_api_key_id BIGINT UNSIGNED NULL AFTER user_id,
  ADD UNIQUE KEY uk_user_api_keys_gateway (gateway_api_key_id),
  ADD CONSTRAINT fk_user_api_keys_gateway
    FOREIGN KEY (gateway_api_key_id) REFERENCES api_keys(id) ON DELETE SET NULL;
