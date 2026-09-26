CREATE INDEX idx_prompt_risk_events_request_match
  ON prompt_risk_events(request_correlation_id, subject_type, subject_key, event_kind);

CREATE INDEX idx_prompt_risk_events_fingerprint_match
  ON prompt_risk_events(prompt_fingerprint, subject_type, subject_key, created_at, event_kind);
