package usecase

import (
	"errors"
	"time"

	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/ports"
)

// errNoChange lets an update callback report "nothing to do" without surfacing
// an error to the caller.
var errNoChange = errors.New("no change")

// ActiveWindow is how long a device counts as present.
const ActiveWindow = 2 * time.Minute

// NeverSeenRoomTTL is how long a room survives when no host ever connects.
const NeverSeenRoomTTL = 15 * time.Minute

// Deps is the dependency bundle every use case receives.
type Deps struct {
	Rooms   ports.RoomRepository
	Devices ports.DeviceRepository
	Events  ports.EventPublisher
	Catalog ports.VideoCatalog
	Clock   ports.Clock
	IDs     ports.IDGenerator
	Cfg     ports.Config
}

// Viewer builds the domain viewer from a device id and role.
func Viewer(deviceID, role string) domain.Viewer {
	return domain.Viewer{DeviceID: deviceID, Role: role}
}

// Now returns the injected clock reading.
func (d Deps) Now() time.Time { return d.Clock.Now() }

// Caps exposes the room limits in the shape the domain wants.
func (d Deps) Caps() domain.Caps {
	return domain.Caps{Guests: d.Cfg.MaxGuests, Queue: d.Cfg.MaxQueue}
}

// snapshot renders a room without touching the repository.
func (d Deps) snapshot(room domain.Room, deviceID, role string) *domain.Snapshot {
	now := d.Now()
	guestCount := d.Devices.ActiveGuests("", now, ActiveWindow)
	snap := domain.ProjectSnapshot(room, guestCount, d.Caps(), Viewer(deviceID, role), now)
	return &snap
}

// touch records device activity, treating host roles as hosts.
func (d Deps) touch(deviceID, role string) {
	d.Devices.Touch(deviceID, role == "tv" || role == "phone", d.Now())
}

// deviceView describes a device inside the update callback.
type deviceView struct {
	ID     string
	Known  bool
	Host   bool
	Active bool
}

// deviceAt inspects presence for guest-cap checks.
func (d Deps) deviceAt(deviceID string, now time.Time) deviceView {
	info, ok := d.Devices.Get(deviceID)
	return deviceView{
		ID:     deviceID,
		Known:  ok,
		Host:   ok && info.Host,
		Active: ok && now.Sub(info.LastSeen) <= ActiveWindow,
	}
}
