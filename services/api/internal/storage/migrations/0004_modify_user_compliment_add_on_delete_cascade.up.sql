alter table user_compliment
    drop constraint telegram_id___fk;

alter table user_compliment
    add constraint telegram_id___fk
        foreign key (telegram_id) references users
            on delete cascade;

alter table user_compliment
    drop constraint compliment_id___fk;

alter table user_compliment
    add constraint compliment_id___fk
        foreign key (compliment_id) references compliments
            on delete cascade;