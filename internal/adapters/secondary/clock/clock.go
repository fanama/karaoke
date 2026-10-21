package clock

import (
	"sync"
	"time"

	"fana-karaoke/internal/ports"
)

// System returns the wall clock.
type System struct{}

// Now returns the current time.
func (System) Now() time.Time { return time.Now() }

// Fake is a manually advanced clock for tests.
type Fake struct {
	mu  sync.Mutex
	now time.Time
}

// NewFake starts a fake clock at the given instant.
func NewFake(at time.Time) *Fake { return &Fake{now: at} }

// Now returns the fake instant.
func (f *Fake) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.now
}

// Advance moves the fake clock forward.
func (f *Fake) Advance(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.now = f.now.Add(d)
}

var (
	_ ports.Clock = System{}
	_ ports.Clock = (*Fake)(nil)
)
