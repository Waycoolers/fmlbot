alter table compliments
    add created_at timestamp default now() not null;