CREATE TABLE user_devices
(
    user_id    BIGINT REFERENCES users (user_id) ON DELETE CASCADE,
    fcm_token  VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT NOW()
);