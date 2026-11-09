-- Legacy Telegram registrations used their chat ID as the account ID.
-- Mobile registrations use Unix microseconds (the same cutoff as cmd/migration).
UPDATE users
SET telegram_id = user_id
WHERE telegram_id IS NULL AND user_id > 0 AND user_id < 1000000000000000
  AND NOT EXISTS (SELECT 1 FROM users linked WHERE linked.telegram_id = users.user_id);

CREATE TABLE telegram_link_codes (
    user_id bigint PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
    code_hash bytea NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL
);
