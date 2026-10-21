package usecase

import (
	"fana-karaoke/internal/domain"
)

// PlaybackService drives the TV and records its telemetry.
type PlaybackService struct {
	Deps
}

// NewPlaybackService builds the playback use case.
func NewPlaybackService(d Deps) *PlaybackService { return &PlaybackService{Deps: d} }

// Playback applies a host action (play, pause, next, cancelCountdown).
func (s *PlaybackService) Playback(code, action, deviceID, role string) (*domain.Snapshot, error) {
	updated, err := s.Rooms.Update(code, func(room domain.Room) (domain.Room, error) {
		next, err := domain.ApplyPlayback(room, action, s.Now())
		if err != nil {
			return room, err
		}
		next.Sequence++
		return next, nil
	})
	if err != nil {
		return nil, err
	}
	s.touch(deviceID, role)
	s.Events.PublishRoomUpdated(code)
	return s.snapshot(updated, deviceID, role), nil
}

// TVEnded advances the queue when the TV reports the current song finished.
func (s *PlaybackService) TVEnded(code, entryID string, generation uint64, deviceID string) {
	s.mutate(code, deviceID, func(room domain.Room) (domain.Room, bool) {
		return domain.AdvanceOnEnded(room, entryID, generation, s.Now())
	})
}

// TVError records a failed song and skips it.
func (s *PlaybackService) TVError(code, entryID, reason string, generation uint64) {
	s.mutate(code, "", func(room domain.Room) (domain.Room, bool) {
		return domain.AdvanceOnError(room, entryID, reason, generation, s.Now())
	})
}

// TVProgress stores drift-correcting playback telemetry.
func (s *PlaybackService) TVProgress(code, entryID string, generation uint64, position, duration float64) {
	s.mutate(code, "", func(room domain.Room) (domain.Room, bool) {
		return domain.UpdateProgress(room, entryID, generation, position, duration, s.Now())
	})
}

func (s *PlaybackService) mutate(code, deviceID string, fn func(domain.Room) (domain.Room, bool)) {
	updated, err := s.Rooms.Update(code, func(room domain.Room) (domain.Room, error) {
		next, changed := fn(room)
		if !changed {
			return room, errNoChange
		}
		return next, nil
	})
	if err != nil {
		return
	}
	if deviceID != "" {
		s.Devices.Touch(deviceID, false, s.Now())
	}
	s.Events.PublishRoomUpdated(updated.Code)
}
