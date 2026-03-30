alter table user_config
    add telegram_id integer default 0 not null
        constraint user_config_users_telegram_id_fk
            references users
            on delete cascade;