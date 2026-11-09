package ui

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestDayKeyboardOffersEveryDayOfPastMonth(t *testing.T) {
	ui := &MenuUI{}
	kb := ui.BuildDayKeyboard(2024, 2, false)
	seen := map[int]bool{}
	for _, row := range kb.Rows {
		if len(row.Buttons) == 0 || len(row.Buttons) > 7 {
			t.Fatalf("bad row size %d", len(row.Buttons))
		}
		for _, b := range row.Buttons {
			d, _ := strconv.Atoi(b.Text)
			seen[d] = true
		}
	}
	for d := 1; d <= 29; d++ {
		if !seen[d] {
			t.Errorf("day %d is missing", d)
		}
	}
	if len(seen) != 29 {
		t.Errorf("got %d days", len(seen))
	}
}

func TestFutureDatesAreHidden(t *testing.T) {
	ui := &MenuUI{}
	now := time.Now().In(moscow)

	days := 0
	for _, row := range ui.BuildDayKeyboard(now.Year(), int(now.Month()), false).Rows {
		days += len(row.Buttons)
	}
	if days != now.Day() {
		t.Errorf("got %d days, want %d", days, now.Day())
	}

	months := 0
	for _, row := range ui.BuildMonthKeyboard(now.Year(), false).Rows {
		months += len(row.Buttons)
	}
	if months != int(now.Month()) {
		t.Errorf("got %d months, want %d", months, int(now.Month()))
	}

	months = 0
	for _, row := range ui.BuildMonthKeyboard(now.Year()-1, false).Rows {
		months += len(row.Buttons)
	}
	if months != 12 {
		t.Errorf("past year: got %d months", months)
	}
}

func TestEveryYearIsReachableByBackButton(t *testing.T) {
	ui := &MenuUI{}
	current := time.Now().In(moscow).Year()
	seen := map[int]bool{}

	for page, ok := current, true; ok; {
		ok = false
		for _, row := range ui.BuildYearKeyboard(page, false).Rows {
			for _, b := range row.Buttons {
				value, _ := strconv.Atoi(b.Data[strings.LastIndex(b.Data, ":")+1:])
				if strings.Contains(b.Data, "year:select:") {
					seen[value] = true
				}
				if strings.Contains(b.Data, "year:page:") && value < page {
					page, ok = value, true
				}
			}
		}
	}

	for year := YearStart; year <= current; year++ {
		if !seen[year] {
			t.Errorf("year %d is unreachable", year)
		}
	}
	if len(seen) != current-YearStart+1 {
		t.Errorf("got %d years", len(seen))
	}
}
