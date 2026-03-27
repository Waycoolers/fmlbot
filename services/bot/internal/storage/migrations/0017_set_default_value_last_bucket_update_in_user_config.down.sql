alter table public.user_config
    alter column last_bucket_update drop not null;

alter table public.user_config
    alter column last_bucket_update drop default;