package ratelimit

import (
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {
	now := time.Now()
	l := New(2, time.Minute)
	l.now = func() time.Time { return now }

	if !l.Allow("a") || !l.Allow("a") {
		t.Fatal("first two events must pass")
	}
	if l.Allow("a") {
		t.Fatal("third event must be blocked")
	}
	if !l.Allow("b") {
		t.Fatal("keys must not affect each other")
	}

	now = now.Add(time.Minute)
	if l.Blocked("a") {
		t.Fatal("limit must expire with the window")
	}

	l.Record("c")
	l.Record("c")
	l.Reset("c")
	if l.Blocked("c") {
		t.Fatal("reset must clear the counter")
	}
}
