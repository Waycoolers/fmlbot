CREATE TABLE cron_tasks
(
    name        VARCHAR(64) PRIMARY KEY,
    last_run_at TIMESTAMPTZ NOT NULL
);
