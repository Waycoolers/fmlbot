-- Если на проде уже есть ники, отличающиеся только регистром, индекс не создаётся,
-- чтобы миграция не остановила сервис. Такие ники надо развести вручную и применить индекс отдельно.
DO
$$
    BEGIN
        IF EXISTS (SELECT 1 FROM users GROUP BY LOWER(username) HAVING COUNT(*) > 1) THEN
            RAISE WARNING 'users_username_lower_key was not created: usernames differ only by case';
        ELSE
            CREATE UNIQUE INDEX users_username_lower_key ON users (LOWER(username));
        END IF;
    END
$$;
