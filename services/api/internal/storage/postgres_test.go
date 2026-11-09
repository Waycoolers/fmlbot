package storage

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"github.com/jmoiron/sqlx"
)

func TestPostgresPairDeliveryAndTelegramLink(t *testing.T) {
	dsn := os.Getenv("FMLBOT_TEST_API_POSTGRES")
	if dsn == "" {
		t.Skip("set FMLBOT_TEST_API_POSTGRES to a fresh disposable database")
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
	must(db.Get(&schemaExists, `SELECT to_regclass('public.users') IS NOT NULL`))
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

	users := &usersRepo{db: db}
	dates := &importantDatesRepo{db: db}
	compliments := &complimentsRepo{db: db}
	deliveries := &schedulerRepo{db: db}
	must(users.AddUser(ctx, 1, "author", []byte("test-hash"), sql.NullInt64{Int64: 11, Valid: true}))
	must(users.AddUser(ctx, 2, "partner", []byte("test-hash"), sql.NullInt64{}))
	must(users.AddUser(ctx, 3, "new_partner", []byte("test-hash"), sql.NullInt64{}))
	linkCode := []byte("fixture code hash")
	must(users.SaveTelegramLink(ctx, 2, linkCode, time.Now().Add(time.Minute)))
	must(users.UseTelegramLink(ctx, 22, linkCode))
	linked, err := users.GetTelegramUserID(ctx, 22)
	must(err)
	if linked != 2 || !errors.Is(users.UseTelegramLink(ctx, 33, linkCode), errs.ErrTelegramLinkInvalid) {
		t.Fatal("Telegram code linked the wrong account or could be reused")
	}

	pair, err := users.CreatePartnerRequest(ctx, 1, 2)
	must(err)
	_, err = users.AcceptPartnerRequest(ctx, pair, 2)
	must(err)
	partner := sql.NullInt64{Int64: 2, Valid: true}
	date, err := dates.AddImportantDate(ctx, 1, partner, pair, "anniversary", time.Date(2020, 2, 29, 0, 0, 0, 0, time.UTC), 0)
	must(err)

	now := time.Now()
	must(deliveries.QueueImportantDateDelivery(ctx, date.ID, 1, now, "telegram", "owner notification"))
	must(deliveries.QueueImportantDateDelivery(ctx, date.ID, 2, now, "telegram", "partner notification"))
	must(deliveries.QueueImportantDateDelivery(ctx, date.ID, 2, now, "telegram", "duplicate"))
	queued, err := deliveries.PendingImportantDateDeliveries(ctx, false)
	must(err)
	if len(queued) != 2 {
		t.Fatalf("expected two deliveries, got %d", len(queued))
	}
	must(deliveries.DeliverImportantDate(ctx, queued[0], func(context.Context) error { return nil }))
	offline := errors.New("temporary network failure")
	if err := deliveries.DeliverImportantDate(ctx, queued[1], func(context.Context) error { return offline }); !errors.Is(err, offline) {
		t.Fatalf("delivery error = %v", err)
	}
	pending, err := deliveries.PendingImportantDateDeliveries(ctx, false)
	must(err)
	if len(pending) != 1 || pending[0].UserID != 2 {
		t.Fatalf("wrong pending deliveries: %+v", pending)
	}
	must(dates.MakeImportantDatePrivate(ctx, date.ID, 2))
	private, err := dates.GetImportantDateByID(ctx, date.ID, 1)
	must(err)
	if private.PartnerID.Valid || private.UserID.Int64 != 1 || !private.LastNotificationAt.Valid {
		t.Fatalf("unsharing transferred ownership: %+v", private)
	}
	must(deliveries.DeliverImportantDate(ctx, pending[0], func(context.Context) error {
		t.Fatal("sent a reminder to a user who lost access")
		return nil
	}))

	_, err = compliments.AddCompliment(ctx, 1, "test compliment")
	must(err)
	_, err = compliments.AcquireCompliment(ctx, 2, 1, pair, func(context.Context, string) error { return offline })
	if !errors.Is(err, errs.ErrComplimentDelivery) {
		t.Fatalf("compliment delivery error = %v", err)
	}
	text, err := compliments.AcquireCompliment(ctx, 2, 1, pair, func(context.Context, string) error { return nil })
	must(err)
	if text != "test compliment" {
		t.Fatalf("failed delivery consumed the compliment: %q", text)
	}

	planID := testPlanLifecycle(t, db, pair)
	must(users.RemovePartners(ctx, 1, 2, pair))
	planStore := &plansRepo{db: db}
	if _, err := planStore.GetPlan(ctx, 2, planID); !errors.Is(err, errs.ErrPlanNotFound) {
		t.Fatalf("former partner can still read the plan: %v", err)
	}
	var remainingPlanNotices int
	must(db.Get(&remainingPlanNotices, `SELECT COUNT(*) FROM couple_plan_notices WHERE plan_id = $1`, planID))
	if remainingPlanNotices != 0 {
		t.Fatal("unlink left plan notifications behind")
	}
	nextPair, err := users.CreatePartnerRequest(ctx, 1, 3)
	must(err)
	_, err = users.AcceptPartnerRequest(ctx, nextPair, 3)
	must(err)
	if err := users.RemovePartners(ctx, 1, 2, pair); !errors.Is(err, errs.ErrPairChanged) {
		t.Fatalf("stale unpair error = %v", err)
	}
	if _, err := dates.AddImportantDate(ctx, 1, partner, pair, "stale date", now, 0); !errors.Is(err, errs.ErrPairChanged) {
		t.Fatalf("stale shared date error = %v", err)
	}
}
