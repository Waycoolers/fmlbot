package storage

import (
	"embed"
	"errors"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func (s *Storage) Migrate() error {
	slog.Info("Running migrations...")
	files, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return err
	}
	driver, err := pgx.WithInstance(s.db.DB, &pgx.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs", files, "pgx", driver)
	if err != nil {
		return err
	}

	err = m.Up()

	if err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
		slog.Info("Migration hasn't changed anything")
	}

	slog.Info("Migration completed successfully")
	return nil
}
