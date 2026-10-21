package usecase

import (
	"time"

	"fana-karaoke/internal/domain"
)

// VoteService runs skip votes.
type VoteService struct {
	Deps
}

// NewVoteService builds the vote use case.
func NewVoteService(d Deps) *VoteService { return &VoteService{Deps: d} }

// StartSkipVote opens a vote for the current song.
func (s *VoteService) StartSkipVote(code, deviceID, role string) (*domain.Snapshot, error) {
	now := s.Now()
	eligible := s.activeCandidates(code, "", now)
	if eligible == nil {
		return nil, domain.ErrRoomNotFound
	}
	updated, err := s.Rooms.Update(code, func(room domain.Room) (domain.Room, error) {
		if room.SkipVote != nil {
			return room, errNoChange
		}
		return domain.StartSkipVote(room, deviceID, eligible, now)
	})
	if err != nil && err != errNoChange {
		return nil, err
	}
	s.touch(deviceID, role)
	if err == nil {
		s.Events.PublishRoomUpdated(code)
	}
	return s.snapshot(updated, deviceID, role), nil
}

// activeCandidates lists eligible voters that are also present. It returns nil
// when the room does not exist.
func (s *VoteService) activeCandidates(code, exclude string, now time.Time) []string {
	room, ok := s.Rooms.Get(code)
	if !ok {
		return nil
	}
	return s.Devices.Active(domain.EligibleDevices(room), now, ActiveWindow)
}

// CastSkipVote records the viewer's ballot.
func (s *VoteService) CastSkipVote(code, deviceID, role string) (*domain.Snapshot, error) {
	updated, err := s.Rooms.Update(code, func(room domain.Room) (domain.Room, error) {
		result, err := domain.CastVote(room, deviceID, s.Now())
		if err != nil {
			return room, err
		}
		if !result.Recorded {
			return room, errNoChange
		}
		return result.Room, nil
	})
	if err != nil && err != errNoChange {
		return nil, err
	}
	s.touch(deviceID, role)
	if err == nil {
		s.Events.PublishRoomUpdated(code)
	}
	return s.snapshot(updated, deviceID, role), nil
}

// CancelSkipVote drops the active vote (host only at the HTTP layer).
func (s *VoteService) CancelSkipVote(code, role, deviceID string) (*domain.Snapshot, error) {
	updated, err := s.Rooms.Update(code, func(room domain.Room) (domain.Room, error) {
		return domain.CancelSkipVote(room), nil
	})
	if err != nil {
		return nil, err
	}
	s.touch(deviceID, role)
	s.Events.PublishRoomUpdated(code)
	return s.snapshot(updated, deviceID, role), nil
}
