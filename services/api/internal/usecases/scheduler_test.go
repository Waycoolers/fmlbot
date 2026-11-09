package usecases

import (
	"context"
	"database/sql"
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
	messages, err := uc.GetAllImportantDatesMessages(context.Background())
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
