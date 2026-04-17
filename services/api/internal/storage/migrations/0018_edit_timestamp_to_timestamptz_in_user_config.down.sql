alter table user_config
    alter column last_compliment_at type timestamp using last_compliment_at::timestamp;

alter table user_config
    alter column last_bucket_update type timestamp using last_bucket_update::timestamp;

