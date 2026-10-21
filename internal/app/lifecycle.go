package app

import (
	"context"
	"time"

	"fana-karaoke/internal/domain"
)

func (m *Manager) Start(ctx context.Context) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.Tick()
		}
	}
}

// Tick advances countdowns, expires votes, and cleans abandoned rooms. It is
// exported so tests can drive time without a real ticker.
func (m *Manager) Tick() {
	now := m.now()
	changed := make(map[string]bool)
	var deleted []string

	m.mu.Lock()
	for code, room := range m.rooms {
		if room.Now != nil && room.Now.Status == domain.StatusCountdown &&
			!room.Now.Countdown.IsZero() && !now.Before(room.Now.Countdown) {
			room.Now.Status = domain.StatusPlaying
			room.Now.StartedAt = room.Now.Countdown
			room.Now.Countdown = time.Time{}
			room.Sequence++
			changed[code] = true
		}
		if room.SkipVote != nil && !now.Before(room.SkipVote.ExpiresAt) {
			yes := len(room.SkipVote.YesDevices)
			threshold := room.SkipVote.Threshold
			room.SkipVote = nil
			room.CooldownUntil = now.Add(30 * time.Second)
			m.addNoticeLocked(room, "notice.skipVoteFailed", map[string]string{
				"yes": itoa(yes), "threshold": itoa(threshold),
			})
			room.Sequence++
			changed[code] = true
		}
		if !room.CooldownUntil.IsZero() && !now.Before(room.CooldownUntil) {
			room.CooldownUntil = time.Time{}
			room.Sequence++
			changed[code] = true
		}
	}

	if m.lastCleanup.IsZero() || now.Sub(m.lastCleanup) >= time.Minute {
		m.lastCleanup = now
		for code, room := range m.rooms {
			expired := false
			if room.HostLastSeenAt.IsZero() {
				expired = now.Sub(room.CreatedAt) > 15*time.Minute
			} else {
				expired = now.Sub(room.HostLastSeenAt) > m.cfg.RoomTTL
			}
			if expired {
				delete(m.rooms, code)
				delete(m.hostTokens, room.HostToken)
				deleted = append(deleted, code)
			}
		}
	}
	m.mu.Unlock()

	for code := range changed {
		m.notify(code)
	}
	for _, code := range deleted {
		m.deleted(code)
	}
}
