UPDATE public.user_config
SET last_bucket_update = NOW()
WHERE last_bucket_update IS NULL;

ALTER TABLE public.user_config
    ALTER COLUMN last_bucket_update SET NOT NULL,
    ALTER COLUMN last_bucket_update SET DEFAULT NOW();