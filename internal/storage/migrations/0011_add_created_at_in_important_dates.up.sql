alter table important_dates
    add created_at timestamp default now() not null;