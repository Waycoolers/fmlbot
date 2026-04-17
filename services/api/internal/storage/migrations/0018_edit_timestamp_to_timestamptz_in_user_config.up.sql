alter table user_config
    alter column last_compliment_at type timestamp with time zone using last_compliment_at::timestamp with time zone;

alter table user_config
    alter column last_bucket_update type timestamp with time zone using last_bucket_update::timestamp with time zone;

