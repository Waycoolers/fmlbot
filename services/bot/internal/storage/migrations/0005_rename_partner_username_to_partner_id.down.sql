alter table users
    rename column partner_id to partner_username;

alter table users
    alter column partner_username type varchar(255) using partner_username::varchar(255);

alter table users
    alter column partner_username set default '';