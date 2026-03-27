alter table user_config
    drop constraint user_config_users_telegram_id_fk;

alter table user_config
    drop column telegram_id;