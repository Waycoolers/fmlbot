ALTER TABLE users
    RENAME COLUMN partner_username TO partner_id;

ALTER TABLE users
    ALTER COLUMN partner_id DROP DEFAULT;

ALTER TABLE users
    ALTER COLUMN partner_id TYPE bigint USING partner_id::bigint;

ALTER TABLE users
    ALTER COLUMN partner_id SET DEFAULT 0;