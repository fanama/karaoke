package usecase

import (
	"context"
	"time"

	"fana-karaoke/internal/domain"
)

// LifecycleService advances time-driven room state.
type LifecycleService struct {
	Deps
	lastCleanup time.Time
}

// NewLifecycleService builds the lifecycle use case.
func NewLifecycleService(d Deps) *LifecycleService { return &LifecycleService{Deps: d} }

// Start ticks the lifecycle every 200ms until the context is cancelled.
func (s *LifecycleService) Start(ctx context.Context) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.Tick()
		}
	}
}

// Tick advances countdowns, expires votes, clears cooldowns and deletes rooms.
func (s *LifecycleService) Tick() {
	now := s.Now()
	for _, code := range s.Rooms.Codes() {
		s.tickRoom(code, now)
	}
	s.cleanup(now)
}

func (s *LifecycleService) tickRoom(code string, now time.Time) {
	changed := false
	_, err := s.Rooms.Update(code, func(room domain.Room) (domain.Room, error) {
		room, a := domain.AdvanceCountdownIfDue(room, now)
		room, b := domain.ExpireSkipVoteIfDue(room, now)
		room, c := domain.ClearCooldownIfDue(room, now)
		changed = a || b || c
		if !changed {
			return room, errNoChange
		}
		return room, nil
	})
	if err == nil && changed {
		s.Events.PublishRoomUpdated(code)
	}
}

func (s *LifecycleService) cleanup(now time.Time) {
	if !s.lastCleanup.IsZero() && now.Sub(s.lastCleanup) < time.Minute {
		return
	}
	s.lastCleanup = now
	for _, code := range s.Rooms.Codes() {
		room, ok := s.Rooms.Get(code)
		if !ok {
			continue
		}
		if !domain.IsExpired(room, now, s.Cfg.RoomTTL, NeverSeenRoomTTL) {
			continue
		}
		if s.Rooms.Delete(code) {
			s.Events.PublishRoomEnded(code)
		}
	}
}
