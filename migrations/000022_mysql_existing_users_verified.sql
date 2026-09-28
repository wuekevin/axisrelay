-- Existing active accounts predate mandatory email verification. Preserve
-- access during upgrade; new registrations are created in pending state.
UPDATE users
SET email_verified_at = COALESCE(email_verified_at, created_at)
WHERE status = 'active' AND email_verified_at IS NULL;
