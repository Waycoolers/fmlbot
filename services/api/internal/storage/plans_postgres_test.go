package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"github.com/Waycoolers/fmlbot/pkg/plans"
	"github.com/jmoiron/sqlx"
)

func testPlanLifecycle(t *testing.T, db *sqlx.DB, pairVersion int64) int64 {
	t.Helper()
	ctx := context.Background()
	store := &plansRepo{db: db}
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	start := time.Now().UTC().Add(2 * time.Hour).Truncate(time.Minute)
	proposal := plans.Proposal{
		Pair:     plans.Pair{PartnerID: 2, Version: pairVersion},
		DraftID:  "9a01020304050607",
		StartsAt: start,
		Card: plans.Card{
			Title: "Пикник у окна", DurationMinutes: 45, EstimatedCost: "до 500 ₽",
			Preparation: "Заварите чай", Steps: []string{"Накройте стол", "Включите любимую музыку"},
			Backup: "Если устали, оставьте только чай и музыку",
		},
	}
	created, err := store.CreatePlan(ctx, 1, proposal)
	must(err)
	duplicate, err := store.CreatePlan(ctx, 1, proposal)
	must(err)
	if created.ID != duplicate.ID || created.Status != plans.Pending || created.ProposedBy != 1 {
		t.Fatalf("unexpected proposal or duplicate: %+v, %+v", created, duplicate)
	}
	if _, err = store.GetPlan(ctx, 3, created.ID); !errors.Is(err, errs.ErrPlanNotFound) {
		t.Fatalf("outsider could read plan: %v", err)
	}
	listed, err := store.ListPlans(ctx, 2, time.Now())
	must(err)
	if len(listed) != 1 || listed[0].ID != created.ID {
		t.Fatalf("proposal absent from partner's plans: %+v", listed)
	}
	if _, err = store.ChangePlan(ctx, 1, created.ID, plans.Change{Action: plans.Accept, Revision: created.Revision}); !errors.Is(err, errs.ErrForbidden) {
		t.Fatalf("author could approve their own proposal: %v", err)
	}

	newTime := start.Add(30 * time.Minute)
	counter, err := store.ChangePlan(ctx, 2, created.ID, plans.Change{
		Action: plans.Reschedule, Revision: created.Revision, StartsAt: &newTime,
	})
	must(err)
	if counter.Revision != created.Revision+1 || counter.ProposedBy != 2 || !counter.StartsAt.Equal(newTime) {
		t.Fatalf("invalid counteroffer: %+v", counter)
	}
	if _, err = store.ChangePlan(ctx, 2, created.ID, plans.Change{Action: plans.Accept, Revision: created.Revision}); !errors.Is(err, errs.ErrPlanChanged) {
		t.Fatalf("obsolete button accepted the plan: %v", err)
	}

	var notices []plans.Notice
	record := func(_ context.Context, notice plans.Notice) error {
		notices = append(notices, notice)
		return nil
	}
	probe := time.Now().Add(time.Second)
	jobs, err := store.DuePlanNotices(ctx, probe)
	must(err)
	if len(jobs) != 2 {
		t.Fatalf("want initial offer and counteroffer, got %d jobs", len(jobs))
	}
	for _, job := range jobs {
		must(store.DeliverPlanNotice(ctx, job, probe, record))
	}
	if len(notices) != 1 || notices[0].ChatID != 11 || notices[0].View.Revision != counter.Revision {
		t.Fatalf("an obsolete notice was sent or the latest notice used the wrong chat: %+v", notices)
	}

	confirmed, err := store.ChangePlan(ctx, 1, created.ID, plans.Change{Action: plans.Accept, Revision: counter.Revision})
	must(err)
	if confirmed.Status != plans.Confirmed || !confirmed.StartsAt.Equal(newTime) {
		t.Fatalf("confirmation didn't keep the agreed time: %+v", confirmed)
	}
	remindAt := newTime.Add(-time.Hour)
	jobs, err = store.DuePlanNotices(ctx, remindAt.Add(-time.Second))
	must(err)
	if len(jobs) != 1 {
		t.Fatalf("reminder arrived before its time: %+v", jobs)
	}
	must(store.DeliverPlanNotice(ctx, jobs[0], time.Now().Add(time.Second), record))
	if len(notices) != 2 || notices[1].ChatID != 22 || notices[1].Reminder {
		t.Fatalf("confirmation notice missing: %+v", notices)
	}

	jobs, err = store.DuePlanNotices(ctx, remindAt)
	must(err)
	if len(jobs) != 2 {
		t.Fatalf("expected a reminder for each partner, got %d", len(jobs))
	}
	failed := errors.New("temporary Telegram failure")
	for _, job := range jobs {
		err = store.DeliverPlanNotice(ctx, job, remindAt, func(ctx context.Context, notice plans.Notice) error {
			if !notice.Reminder || notice.View.Revision != confirmed.Revision {
				t.Fatalf("invalid reminder: %+v", notice)
			}
			if notice.ChatID == 22 {
				return failed
			}
			return record(ctx, notice)
		})
		if err != nil && !errors.Is(err, failed) {
			t.Fatal(err)
		}
	}
	retryAt := remindAt.Add(time.Minute)
	jobs, err = store.DuePlanNotices(ctx, retryAt)
	must(err)
	if len(jobs) != 1 {
		t.Fatalf("expected only the failed reminder, got %d jobs", len(jobs))
	}
	must(store.DeliverPlanNotice(ctx, jobs[0], retryAt, record))
	if len(notices) != 4 || notices[2].ChatID != 11 || notices[3].ChatID != 22 || !notices[3].Reminder {
		t.Fatalf("retry duplicated a successful reminder or used the wrong chat: %+v", notices)
	}

	cancelled, err := store.ChangePlan(ctx, 2, created.ID, plans.Change{Action: plans.Cancel, Revision: confirmed.Revision})
	must(err)
	if cancelled.Status != plans.Cancelled {
		t.Fatalf("plan was not cancelled: %+v", cancelled)
	}
	return created.ID
}
