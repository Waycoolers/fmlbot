ALTER TABLE important_dates
    ALTER COLUMN created_at TYPE timestamptz USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN last_notification_at TYPE timestamptz USING last_notification_at AT TIME ZONE 'UTC';

CREATE TABLE important_date_deliveries (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    date_id bigint NOT NULL REFERENCES important_dates(id) ON DELETE CASCADE,
    user_id bigint NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    due_on date NOT NULL,
    channel text NOT NULL CHECK (channel IN ('telegram', 'push')),
    message text NOT NULL,
    processed_at timestamptz,
    delivered_at timestamptz,
    UNIQUE (date_id, user_id, due_on, channel)
);

CREATE INDEX important_date_deliveries_pending_idx
    ON important_date_deliveries (id) WHERE processed_at IS NULL;
