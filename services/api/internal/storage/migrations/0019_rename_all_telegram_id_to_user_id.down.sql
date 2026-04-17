alter table users
    rename column user_id to telegram_id;

alter table user_config
    rename column user_id to telegram_id;

alter table user_config
    rename constraint user_config_users_user_id_fk to user_config_users_telegram_id_fk;

alter table user_compliment
    rename column user_id to telegram_id;

alter table user_compliment
    rename constraint user_id___fk to telegram_id___fk;

alter table important_dates
    rename column user_id to telegram_id;

alter table important_dates
    alter column telegram_id type integer using telegram_id::integer;

alter table important_dates
    alter column partner_id type integer using partner_id::integer;

alter table important_dates
    rename constraint important_dates_users_user_id_fk to important_dates_users_telegram_id_fk;

alter table important_dates
    rename constraint important_dates_users_user_id_fk_2 to important_dates_users_telegram_id_fk_2;

