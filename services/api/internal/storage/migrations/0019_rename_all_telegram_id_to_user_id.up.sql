alter table users
    rename column telegram_id to user_id;

alter table user_config
    rename column telegram_id to user_id;

alter table user_config
    rename constraint user_config_users_telegram_id_fk to user_config_users_user_id_fk;

alter table user_compliment
    rename column telegram_id to user_id;

alter table user_compliment
    rename constraint telegram_id___fk to user_id___fk;

alter table important_dates
    rename column telegram_id to user_id;

alter table important_dates
    rename constraint important_dates_users_telegram_id_fk to important_dates_users_user_id_fk;

alter table important_dates
    rename constraint important_dates_users_telegram_id_fk_2 to important_dates_users_user_id_fk_2;

