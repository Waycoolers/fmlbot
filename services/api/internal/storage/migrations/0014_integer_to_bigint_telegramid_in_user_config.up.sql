alter table user_config
    alter column telegram_id type bigint using telegram_id::bigint;