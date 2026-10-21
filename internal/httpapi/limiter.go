package httpapi

import (
	"sync"
	"time"
)

type rateEntry struct {
	start time.Time
	count int
}

type Limiter struct {
	mu      sync.Mutex
	entries map[string]*rateEntry
	now     func() time.Time
}

func NewLimiter() *Limiter {
	return &Limiter{entries: make(map[string]*rateEntry), now: time.Now}
}

func (s *Server) limited(key string, limit int, window time.Duration) (bool, int) {
	if s.dev {
		return true, 0
	}
	return s.limiter.Allow(key, limit, window)
}

// Allow returns whether an action is allowed and, when it is not, the number
// of seconds until the window resets.
func (l *Limiter) Allow(key string, limit int, window time.Duration) (bool, int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	e := l.entries[key]
	if e == nil || now.Sub(e.start) >= window {
		if len(l.entries) > 10000 {
			for k := range l.entries {
				delete(l.entries, k)
			}
		}
		l.entries[key] = &rateEntry{start: now, count: 1}
		return true, 0
	}
	if e.count >= limit {
		retry := int(window.Seconds()-now.Sub(e.start).Seconds()) + 1
		return false, retry
	}
	e.count++
	return true, 0
}
