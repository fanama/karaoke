package memory

import (
	"sort"
	"sync"
	"time"

	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/ports"
)

// DeviceRepo tracks browser presence in memory.
type DeviceRepo struct {
	mu      sync.RWMutex
	devices map[string]domain.DeviceInfo
}

// NewDeviceRepo builds an empty presence store.
func NewDeviceRepo() *DeviceRepo {
	return &DeviceRepo{devices: map[string]domain.DeviceInfo{}}
}

// Touch marks a device as seen now.
func (r *DeviceRepo) Touch(id string, host bool, now time.Time) {
	if id == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	info := r.devices[id]
	info.LastSeen = now
	if host {
		info.Host = true
	}
	r.devices[id] = info
}

// SetConnected records whether a device currently holds a socket.
func (r *DeviceRepo) SetConnected(id string, connected bool, now time.Time) {
	if id == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	info := r.devices[id]
	info.Connected = connected
	if connected {
		info.LastSeen = now
	}
	r.devices[id] = info
}

// Get returns a device's presence record.
func (r *DeviceRepo) Get(id string) (domain.DeviceInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, ok := r.devices[id]
	return info, ok
}

// ActiveGuests counts non-host devices seen inside the window.
func (r *DeviceRepo) ActiveGuests(excludeID string, now time.Time, window time.Duration) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for id, info := range r.devices {
		if id == excludeID || info.Host {
			continue
		}
		if now.Sub(info.LastSeen) <= window {
			count++
		}
	}
	return count
}

// Active filters ids down to devices seen inside the window, sorted.
func (r *DeviceRepo) Active(ids []string, now time.Time, window time.Duration) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		info, ok := r.devices[id]
		if ok && now.Sub(info.LastSeen) <= window {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

var _ ports.DeviceRepository = (*DeviceRepo)(nil)
