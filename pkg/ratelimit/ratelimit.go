package ratelimit

import (
	"sync"
	"time"
)

const sweepThreshold = 10000

type entry struct {
	count   int
	resetAt time.Time
}

// Limiter - счётчик событий на ключ в фиксированном окне. Живёт в памяти процесса.
type Limiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	entries map[string]*entry
	now     func() time.Time
}

func New(limit int, window time.Duration) *Limiter {
	return &Limiter{limit: limit, window: window, entries: make(map[string]*entry), now: time.Now}
}

func (l *Limiter) current(key string, now time.Time) *entry {
	e, ok := l.entries[key]
	if !ok || !now.Before(e.resetAt) {
		return nil
	}
	return e
}

// Blocked сообщает, исчерпан ли лимит для ключа. Счётчик не меняется.
func (l *Limiter) Blocked(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.current(key, l.now())
	return e != nil && e.count >= l.limit
}

// Record учитывает одно событие для ключа.
func (l *Limiter) Record(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	if len(l.entries) >= sweepThreshold {
		for k, e := range l.entries {
			if !now.Before(e.resetAt) {
				delete(l.entries, k)
			}
		}
	}
	e := l.current(key, now)
	if e == nil {
		e = &entry{resetAt: now.Add(l.window)}
		l.entries[key] = e
	}
	e.count++
}

// Allow атомарно проверяет лимит и учитывает событие.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	now := l.now()
	e := l.current(key, now)
	blocked := e != nil && e.count >= l.limit
	l.mu.Unlock()
	if blocked {
		return false
	}
	l.Record(key)
	return true
}

// Reset сбрасывает счётчик ключа.
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, key)
}
