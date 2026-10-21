package ports

import (
	"context"
	"time"

	"fana-karaoke/internal/domain"
)

// Config holds room limits shared by the use cases and the HTTP adapter.
type Config struct {
	MaxRooms  int
	MaxGuests int
	MaxQueue  int
	RoomTTL   time.Duration
}

// RoomRepository stores rooms and applies atomic read-modify-write updates.
type RoomRepository interface {
	Count() int
	Create(room domain.Room) error
	Get(code string) (domain.Room, bool)
	Exists(code string) bool
	Resolve(input string) (string, bool)
	ByHostToken(token string) (string, bool)
	Update(code string, fn func(domain.Room) (domain.Room, error)) (domain.Room, error)
	Delete(code string) bool
	Codes() []string
}

// DeviceRepository tracks presence so guests, hosts and voters can be counted.
type DeviceRepository interface {
	Touch(id string, host bool, now time.Time)
	SetConnected(id string, connected bool, now time.Time)
	Get(id string) (domain.DeviceInfo, bool)
	ActiveGuests(excludeID string, now time.Time, window time.Duration) int
	Active(ids []string, now time.Time, window time.Duration) []string
}

// EventPublisher notifies connected clients about room changes.
type EventPublisher interface {
	PublishRoomUpdated(code string)
	PublishRoomEnded(code string)
}

// VideoCatalog answers search queries and resolves video metadata.
type VideoCatalog interface {
	Search(ctx context.Context, query string) ([]domain.SearchResult, bool, error)
	Lookup(ctx context.Context, videoID string) (domain.VideoMeta, error)
}

// EmbedProber checks whether a video can be played in an iframe.
type EmbedProber interface {
	Probe(ctx context.Context, videoID string) (domain.VideoMeta, error)
}

// Clock supplies the current time so time-dependent rules stay testable.
type Clock interface {
	Now() time.Time
}

// IDGenerator creates opaque ids, room codes, host tokens and PINs.
type IDGenerator interface {
	NewID(prefix string) string
	NewRoomCode() string
	NewHostToken() string
	NewPIN() string
}

// QRGenerator renders a join URL as an SVG QR code.
type QRGenerator interface {
	EncodeSVG(target string) ([]byte, error)
}

// SessionSigner seals and verifies cookie payloads.
type SessionSigner interface {
	Seal(value string) string
	Verify(signed string) (string, bool)
}

// RateLimiter implements a token bucket per key.
type RateLimiter interface {
	Allow(key string, limit int, window time.Duration) (bool, int)
}
