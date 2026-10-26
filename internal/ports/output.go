package ports

import (
	"context"
	"io"
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

// LibraryFormat selects the portable transfer representation.
type LibraryFormat string

const (
	// FormatSQLite is a SQLite database file.
	FormatSQLite LibraryFormat = "sqlite"
	// FormatSQL is the app's own SQL dump format.
	FormatSQL LibraryFormat = "sql"
)

// LibraryStatus describes the active library without exposing its secrets.
type LibraryStatus struct {
	Backend    string
	Connected  bool
	CacheOnly  bool
	Songs      int
	LastSyncAt time.Time
	LastError  string
}

// Admin song list bounds, shared by the use case and the HTTP adapter.
const (
	// DefaultSongPageSize is the page size used when a request omits one.
	DefaultSongPageSize = 25
	// MaxSongPageSize caps how many songs one request may read.
	MaxSongPageSize = 100
)

// SongQuery pages and filters the stored library. Offset is the number of songs
// to skip, Limit the page size, and Query an optional substring filter.
type SongQuery struct {
	Offset int
	Limit  int
	Query  string
}

// SongPage is one window of the library plus the total matching the filter.
type SongPage struct {
	Songs  []domain.SearchResult
	Total  int
	Offset int
	Limit  int
	Query  string
}

// SongLibrary is the local song store searched before the upstream catalog.
// Saving the same video twice updates it instead of creating a duplicate.
type SongLibrary interface {
	Search(ctx context.Context, query string, limit int) ([]domain.SearchResult, error)
	// List returns one page of stored songs ordered by title, with the total.
	List(ctx context.Context, query SongQuery) (SongPage, error)
	// Random samples up to limit songs so the guest screen can suggest tracks.
	Random(ctx context.Context, limit int) ([]domain.SearchResult, error)
	// Lookup returns stored metadata so queueing never needs the network.
	Lookup(ctx context.Context, videoID string) (domain.VideoMeta, bool)
	// Cache updates memory immediately and writes durably best-effort.
	Cache(ctx context.Context, songs []domain.SearchResult) error
	// Save waits for durable storage; the cache is unchanged on failure.
	Save(ctx context.Context, songs []domain.SearchResult) error
	Count(ctx context.Context) (int, error)
	Status(ctx context.Context) LibraryStatus
	Export(ctx context.Context, format LibraryFormat, w io.Writer) error
	Import(ctx context.Context, format LibraryFormat, r io.Reader) (int, error)
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
