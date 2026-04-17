module github.com/Waycoolers/fmlbot/services/api

go 1.25.4

require (
	github.com/Waycoolers/fmlbot/common v0.0.0-00010101000000-000000000000
	github.com/golang-migrate/migrate/v4 v4.19.1
	github.com/jmoiron/sqlx v1.4.0
	github.com/joho/godotenv v1.5.1
	github.com/robfig/cron/v3 v3.0.1
)

require (
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.3 // indirect
	github.com/jackc/pgerrcode v0.0.0-20220416144525-469b46aa5efa // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgtype v1.14.0 // indirect
	github.com/jackc/pgx/v4 v4.18.2 // indirect
	github.com/lib/pq v1.10.9 // indirect
	golang.org/x/crypto v0.49.0 // indirect
	golang.org/x/text v0.35.0 // indirect
)

replace github.com/Waycoolers/fmlbot/services/auth => ../auth

replace github.com/Waycoolers/fmlbot/common => ../../common
