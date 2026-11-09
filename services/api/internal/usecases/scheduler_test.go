package usecases

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
)

func day(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestIsEventAndNotifyDay(t *testing.T) {
	cases := []struct {
		name         string
		date, today  time.Time
		before       int
		event, nofiy bool
	}{
		{"anniversary from a past year", day(2013, 5, 25), day(2026, 5, 25), 0, true, true},
		{"reminder for a past year", day(2013, 5, 25), day(2026, 5, 18), 7, false, true},
		{"ordinary day", day(2013, 5, 25), day(2026, 5, 19), 7, false, false},
		{"reminder crosses new year", day(2020, 1, 2), day(2026, 12, 26), 7, false, true},
		{"29 february in a common year", day(2024, 2, 29), day(2026, 2, 28), 0, true, true},
		{"29 february in a leap year", day(2024, 2, 29), day(2028, 2, 29), 0, true, true},
		{"28 february is not the date in a leap year", day(2024, 2, 29), day(2028, 2, 28), 0, false, false},
	}
	for _, c := range cases {
		event, notify := isEventAndNotifyDay(c.date, c.before, c.today)
		if event != c.event || notify != c.nofiy {
			t.Errorf("%s: got event=%v notify=%v", c.name, event, notify)
		}
	}
}

type datesStub struct {
	domain.ImportantDatesRepo
	dates []domain.ImportantDate
}

func (s *datesStub) GetAllActiveImportantDates(context.Context) ([]domain.ImportantDate, error) {
	return s.dates, nil
}

func messagesFor(t *testing.T, date domain.ImportantDate) []domain.ImportantDateMessage {
	t.Helper()
	uc := New(&domain.Repos{ImportantDates: &datesStub{dates: []domain.ImportantDate{date}}}, nil, nil)
	messages, err := uc.GetAllImportantDatesMessages(context.Background(), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	return messages
}

func TestMessagesAreEscapedForTelegramAndPlainForPush(t *testing.T) {
	now := time.Now()
	messages := messagesFor(t, domain.ImportantDate{
		ID:       1,
		UserID:   sql.NullInt64{Int64: 1, Valid: true},
		Title:    "Мы <3 друг друга",
		Date:     time.Date(now.Year()-3, now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
		IsActive: true,
	})
	if len(messages) != 1 {
		t.Fatalf("got %d messages", len(messages))
	}
	if !strings.Contains(messages[0].Message, "<b>Мы &lt;3 друг друга</b>") {
		t.Errorf("message: %q", messages[0].Message)
	}
	if !strings.Contains(messages[0].PlainMessage, "Мы <3 друг друга") || strings.Contains(messages[0].PlainMessage, "<b>") {
		t.Errorf("plain message: %q", messages[0].PlainMessage)
	}
}

func TestDateCreatedTodayIsNotNotified(t *testing.T) {
	now := time.Now()
	messages := messagesFor(t, domain.ImportantDate{
		ID:        1,
		UserID:    sql.NullInt64{Int64: 1, Valid: true},
		Title:     "x",
		Date:      time.Date(now.Year()-1, now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
		IsActive:  true,
		CreatedAt: now,
	})
	if len(messages) != 0 {
		t.Fatalf("got %d messages", len(messages))
	}
}

func TestDateCreatedJustAfterMoscowMidnightIsNotNotified(t *testing.T) {
	moscow := time.FixedZone("MSK", 3*60*60)
	now := time.Date(2026, 11, 9, 12, 0, 0, 0, moscow)
	uc := New(&domain.Repos{ImportantDates: &datesStub{dates: []domain.ImportantDate{{
		ID: 1, UserID: sql.NullInt64{Int64: 1, Valid: true}, IsActive: true,
		Date: day(2025, 11, 9), CreatedAt: time.Date(2026, 11, 8, 21, 30, 0, 0, time.UTC),
	}}}}, nil, nil)

	messages, err := uc.GetAllImportantDatesMessages(context.Background(), now)
	if err != nil || len(messages) != 0 {
		t.Fatalf("new date generated a notification: %v, %v", messages, err)
	}
}

type deliveryStub struct {
	domain.SchedulerRepo
	pending []domain.ImportantDateDelivery
}

func (s *deliveryStub) PendingImportantDateDeliveries(context.Context, bool) ([]domain.ImportantDateDelivery, error) {
	return append([]domain.ImportantDateDelivery(nil), s.pending...), nil
}

func (s *deliveryStub) DeliverImportantDate(ctx context.Context, delivery domain.ImportantDateDelivery, send func(context.Context) error) error {
	if err := send(ctx); err != nil {
		return err
	}
	for i, pending := range s.pending {
		if pending.ID == delivery.ID {
			s.pending = append(s.pending[:i], s.pending[i+1:]...)
			break
		}
	}
	return nil
}

type notificationSenderStub struct {
	domain.Sender
	fail map[int64]bool
	sent []int64
}

func (s *notificationSenderStub) SendImportantDatesNotification(_ context.Context, msg domain.ImportantDateMessage) error {
	if len(msg.UserIDs) != 1 {
		return errors.New("expected one recipient")
	}
	userID := msg.UserIDs[0]
	if s.fail[userID] {
		return errors.New("temporary delivery failure")
	}
	s.sent = append(s.sent, userID)
	return nil
}

func TestDeliveryRetriesOnlyFailedRecipientAndChannel(t *testing.T) {
	store := &deliveryStub{pending: []domain.ImportantDateDelivery{
		{ID: 1, DateID: 1, UserID: 1, Channel: "telegram"},
		{ID: 2, DateID: 1, UserID: 2, Channel: "telegram"},
		{ID: 3, DateID: 1, UserID: 2, Channel: "push"},
	}}
	uc := New(&domain.Repos{Scheduler: store}, nil, nil)
	telegram := &notificationSenderStub{fail: map[int64]bool{2: true}}
	push := &notificationSenderStub{}
	if err := uc.DeliverImportantDates(context.Background(), telegram, push); err != nil {
		t.Fatal(err)
	}
	if len(store.pending) != 1 || store.pending[0].ID != 2 || len(push.sent) != 1 {
		t.Fatalf("partial failure affected another delivery: pending=%+v push=%v", store.pending, push.sent)
	}
	delete(telegram.fail, 2)
	if err := uc.DeliverImportantDates(context.Background(), telegram, push); err != nil {
		t.Fatal(err)
	}
	if len(store.pending) != 0 || len(push.sent) != 1 || len(telegram.sent) != 2 {
		t.Fatalf("retry sent duplicates or skipped delivery: pending=%+v push=%v telegram=%v", store.pending, push.sent, telegram.sent)
	}
}
