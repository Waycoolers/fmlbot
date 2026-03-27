alter table public.important_dates
    alter column telegram_id type bigint using telegram_id::bigint;

alter table public.important_dates
    alter column partner_id type bigint using partner_id::bigint;
