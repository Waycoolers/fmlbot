alter table user_config
    rename column compliment_count to max_compliment_count;

alter table user_config
    add compliment_count integer default 0 not null;