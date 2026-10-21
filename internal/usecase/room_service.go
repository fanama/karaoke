package usecase

import (
	"context"
	"time"

	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/ports"
)

// RoomService owns the room lifecycle and presence.
type RoomService struct {
	Deps
}

// NewRoomService builds the room use case.
func NewRoomService(d Deps) *RoomService { return &RoomService{Deps: d} }

// CreateRoom allocates a room, a host token and a PIN.
func (s *RoomService) CreateRoom(_ context.Context) (string, string, string, error) {
	if s.Rooms.Count() >= s.Cfg.MaxRooms {
		return "", "", "", domain.ErrServerFull
	}
	code, err := s.uniqueCode()
	if err != nil {
		return "", "", "", err
	}
	token := s.IDs.NewHostToken()
	pin := s.IDs.NewPIN()
	room := domain.NewRoom(code, token, pin, s.Now())
	if err := s.Rooms.Create(room); err != nil {
		return "", "", "", domain.ErrServerFull
	}
	return code, token, pin, nil
}

func (s *RoomService) uniqueCode() (string, error) {
	for i := 0; i < 40; i++ {
		candidate := s.IDs.NewRoomCode()
		if !s.Rooms.Exists(candidate) {
			return candidate, nil
		}
	}
	return "", domain.ErrServerFull
}

// ResolveCode maps a compact or canonical code onto a live room.
func (s *RoomService) ResolveCode(input string) (string, bool) { return s.Rooms.Resolve(input) }

// Exists reports whether a room code is live.
func (s *RoomService) Exists(code string) bool { return s.Rooms.Exists(code) }

// RoomByHostToken finds the room behind a secret TV token.
func (s *RoomService) RoomByHostToken(token string) (string, bool) {
	return s.Rooms.ByHostToken(token)
}

// PIN returns the current room PIN.
func (s *RoomService) PIN(code string) (string, bool) {
	room, ok := s.Rooms.Get(code)
	if !ok {
		return "", false
	}
	return room.PIN, true
}

// RotatePIN generates and stores a fresh PIN.
func (s *RoomService) RotatePIN(code string) (string, error) {
	pin := s.IDs.NewPIN()
	updated, err := s.Rooms.Update(code, func(room domain.Room) (domain.Room, error) {
		room = domain.RotatePIN(room, pin)
		room.Sequence++
		return room, nil
	})
	if err != nil {
		return "", err
	}
	s.Events.PublishRoomUpdated(updated.Code)
	return pin, nil
}

// EndRoom deletes a room and notifies its clients.
func (s *RoomService) EndRoom(code string) error {
	if !s.Rooms.Delete(code) {
		return domain.ErrRoomNotFound
	}
	s.Events.PublishRoomEnded(code)
	return nil
}

// Snapshot renders the room for one viewer, enforcing the guest cap.
func (s *RoomService) Snapshot(code, deviceID, role string) (*domain.Snapshot, error) {
	room, ok := s.Rooms.Get(code)
	if !ok {
		return nil, domain.ErrRoomNotFound
	}
	now := s.Now()
	viewer := Viewer(deviceID, role)
	if err := s.admit(viewer, now); err != nil {
		return nil, err
	}
	s.Devices.Touch(deviceID, viewer.IsHost(), now)
	return s.snapshot(room, deviceID, role), nil
}

func (s *RoomService) admit(viewer domain.Viewer, now time.Time) error {
	if viewer.IsHost() {
		return nil
	}
	device := s.deviceAt(viewer.DeviceID, now)
	if device.Known && (device.Host || device.Active) {
		return nil
	}
	if s.Devices.ActiveGuests(viewer.DeviceID, now, ActiveWindow) >= s.Cfg.MaxGuests {
		return domain.ErrRoomFull
	}
	return nil
}

// HostSeen records host activity.
func (s *RoomService) HostSeen(code string) {
	_, _ = s.Rooms.Update(code, func(room domain.Room) (domain.Room, error) {
		return domain.TouchHost(room, s.Now()), nil
	})
}

// TouchDevice records device presence.
func (s *RoomService) TouchDevice(deviceID string, host bool) {
	s.Devices.Touch(deviceID, host, s.Now())
}

// Config exposes the room limits for /api/config.
func (s *RoomService) Config() ports.Config { return s.Cfg }
