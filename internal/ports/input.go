package ports

import (
	"context"
	"io"

	"fana-karaoke/internal/domain"
)

// RoomCommands creates, resolves and administers rooms.
type RoomCommands interface {
	CreateRoom(ctx context.Context) (code, token, pin string, err error)
	ResolveCode(input string) (string, bool)
	RoomByHostToken(token string) (string, bool)
	Exists(code string) bool
	PIN(code string) (string, bool)
	RotatePIN(code string) (string, error)
	EndRoom(code string) error
	Snapshot(code, deviceID, role string) (*domain.Snapshot, error)
	HostSeen(code string)
	TouchDevice(deviceID string, host bool)
	Config() Config
}

// QueueCommands manages the song queue.
type QueueCommands interface {
	AddSong(ctx context.Context, code, deviceID, role, nickname, videoID, rawURL string) (*domain.Snapshot, error)
	ToggleVote(code, deviceID, role, entryID string) (*domain.Snapshot, error)
	RemoveSong(code, entryID, deviceID, role string) (*domain.Snapshot, error)
	Reorder(code string, ids []string, deviceID, role string) (*domain.Snapshot, error)
	ClearQueue(code, deviceID, role string) (*domain.Snapshot, error)
	LockQueue(code string, locked bool, deviceID, role string) (*domain.Snapshot, error)
}

// PlaybackCommands drives the TV and receives its telemetry.
type PlaybackCommands interface {
	Playback(code, action, deviceID, role string) (*domain.Snapshot, error)
	TVEnded(code, entryID string, generation uint64, deviceID string)
	TVError(code, entryID, reason string, generation uint64)
	TVProgress(code, entryID string, generation uint64, position, duration float64)
}

// VoteCommands runs skip votes.
type VoteCommands interface {
	StartSkipVote(code, deviceID, role string) (*domain.Snapshot, error)
	CastSkipVote(code, deviceID, role string) (*domain.Snapshot, error)
	CancelSkipVote(code, role, deviceID string) (*domain.Snapshot, error)
}

// CatalogQueries searches YouTube karaoke tracks.
type CatalogQueries interface {
	Search(ctx context.Context, query string) ([]domain.SearchResult, bool, error)
}

// LibraryCommands inspects, exports and merges the local song library.
type LibraryCommands interface {
	LibraryStats(ctx context.Context) (int, error)
	LibraryInfo(ctx context.Context) (LibraryStatus, error)
	ExportLibrary(ctx context.Context, format LibraryFormat, w io.Writer) error
	ImportLibrary(ctx context.Context, format LibraryFormat, r io.Reader) (int, error)
}

// KaraokeService is the single facade the primary adapters depend on.
type KaraokeService interface {
	RoomCommands
	QueueCommands
	PlaybackCommands
	VoteCommands
	CatalogQueries
	LibraryCommands
}
