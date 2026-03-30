alter table user_config
    add constraint check_compliment_token_bucket
        check ((compliment_token_bucket >= 0) AND (compliment_token_bucket <= 2));