alter table public.important_dates
    alter column telegram_id type integer using telegram_id::integer;

alter table public.important_dates
    alter column partner_id type integer using partner_id::integer;
