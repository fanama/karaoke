package memory

import (
	"sync"
	"time"

	"fana-karaoke/internal/ports"
)

type rateEntry struct {
	start time.Time
	count int
}

// RateLimiter is a fixed-window counter keyed by an arbitrary string.
type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateEntry
	now     func() time.Time
}

// NewRateLimiter builds an empty limiter.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{entries: map[string]*rateEntry{}, now: time.Now}
}

// Allow reports whether the action is allowed and, if not, the retry delay.
func (l *RateLimiter) Allow(key string, limit int, window time.Duration) (bool, int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	entry := l.entries[key]
	if entry == nil || now.Sub(entry.start) >= window {
		l.compactLocked()
		l.entries[key] = &rateEntry{start: now, count: 1}
		return true, 0
	}
	if entry.count >= limit {
		return false, int(window.Seconds()-now.Sub(entry.start).Seconds()) + 1
	}
	entry.count++
	return true, 0
}

func (l *RateLimiter) compactLocked() {
	if len(l.entries) <= 10000 {
		return
	}
	for key := range l.entries {
		delete(l.entries, key)
	}
}

var _ ports.RateLimiter = (*RateLimiter)(nil)
