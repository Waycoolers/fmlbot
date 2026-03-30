alter table user_config
    drop column compliment_count;

alter table user_config
    rename column max_compliment_count to compliment_count;