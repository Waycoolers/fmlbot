alter table user_config
    alter column telegram_id type integer using telegram_id::integer;