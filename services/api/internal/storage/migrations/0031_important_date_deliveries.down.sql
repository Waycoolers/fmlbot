DROP TABLE important_date_deliveries;

ALTER TABLE important_dates
    ALTER COLUMN created_at TYPE timestamp USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN last_notification_at TYPE timestamp USING last_notification_at AT TIME ZONE 'UTC';
