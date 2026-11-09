package storage

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"github.com/jmoiron/sqlx"
)

func TestPostgresRefreshRotationRollback(t *testing.T) {
	dsn := os.Getenv("FMLBOT_TEST_AUTH_POSTGRES")
	if dsn == "" {
		t.Skip("set FMLBOT_TEST_AUTH_POSTGRES to a fresh disposable database")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(1)
	ctx := context.Background()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	var schemaExists bool
	must(db.Get(&schemaExists, `SELECT to_regclass('public.refresh_tokens') IS NOT NULL`))
	if schemaExists {
		t.Fatal("test requires a fresh database")
	}
	files, err := fs.ReadDir(migrationFiles, "migrations")
	must(err)
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".up.sql") {
			continue
		}
		query, err := migrationFiles.ReadFile("migrations/" + file.Name())
		must(err)
		if _, err := db.ExecContext(ctx, string(query)); err != nil {
			t.Fatalf("migration %s: %v", file.Name(), err)
		}
	}

	tokens := &tokensRepo{db: db}
	oldToken, err := tokens.Create(ctx, 1, time.Hour)
	must(err)
	_, err = db.Exec(`ALTER TABLE refresh_tokens ADD CONSTRAINT reject_new_test_tokens CHECK (user_id <> 1 OR revoked) NOT VALID`)
	must(err)
	if _, _, err = tokens.Rotate(ctx, oldToken, time.Hour); err == nil {
		t.Fatal("rotation unexpectedly succeeded")
	}
	_, err = tokens.Validate(ctx, oldToken)
	must(err)

	_, err = db.Exec(`ALTER TABLE refresh_tokens DROP CONSTRAINT reject_new_test_tokens`)
	must(err)
	userID, replacement, err := tokens.Rotate(ctx, oldToken, time.Hour)
	must(err)
	if userID != 1 || replacement == oldToken || replacement == "" {
		t.Fatal("rotation returned invalid replacement")
	}
	if _, _, err = tokens.Rotate(ctx, oldToken, time.Hour); !errors.Is(err, errs.ErrTokenInvalid) {
		t.Fatalf("old token replay error = %v", err)
	}
	must(tokens.RevokeAll(ctx, 1))
	if _, err = tokens.Validate(ctx, replacement); !errors.Is(err, errs.ErrTokenInvalid) {
		t.Fatalf("replacement survived revocation: %v", err)
	}
}
