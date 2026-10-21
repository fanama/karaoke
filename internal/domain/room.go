package domain

import (
	"strings"
	"time"
)

// NewRoom builds a fresh room that is ready to accept songs immediately.
func NewRoom(code, hostToken, pin string, now time.Time) Room {
	return Room{
		Code:           code,
		HostToken:      hostToken,
		PIN:            pin,
		CreatedAt:      now,
		HostLastSeenAt: time.Time{},
		TVStarted:      true,
		Sequence:       1,
	}
}

// TouchHost records host activity so cleanup does not delete a live room.
func TouchHost(room Room, now time.Time) Room {
	room.HostLastSeenAt = now
	return room
}

// IsExpired reports whether a room should be cleaned up at now.
func IsExpired(room Room, now time.Time, ttl, neverSeenTTL time.Duration) bool {
	if room.HostLastSeenAt.IsZero() {
		return now.Sub(room.CreatedAt) > neverSeenTTL
	}
	return now.Sub(room.HostLastSeenAt) > ttl
}

// NormalizeCompactCode strips separators so "bluemice" matches "blue-mice".
func NormalizeCompactCode(input string) string {
	return strings.NewReplacer("-", "", "_", "", " ", "").Replace(strings.ToLower(strings.TrimSpace(input)))
}

// HasPIN reports whether a candidate PIN matches the room PIN in constant time.
func HasPIN(room Room, candidate string) bool { return ConstantTimeEqual(room.PIN, candidate) }

// RotatePIN stores a freshly generated PIN on the room.
func RotatePIN(room Room, pin string) Room {
	room.PIN = pin
	return room
}

// AddNotice prepends a notice and bumps the room sequence.
func AddNotice(room Room, id, code string, params map[string]string, now time.Time) Room {
	room = PrependNotice(room, id, code, params, now)
	room.Sequence++
	return room
}
