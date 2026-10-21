package domain

import "time"

type Status string

const (
	StatusIdle      Status = "idle"
	StatusCountdown Status = "countdown"
	StatusPlaying   Status = "playing"
	StatusPaused    Status = "paused"
)

type QueueEntry struct {
	ID              string          `json:"id"`
	VideoID         string          `json:"videoId"`
	Title           string          `json:"title"`
	Channel         string          `json:"channel"`
	ThumbnailURL    string          `json:"thumbnailUrl"`
	DurationSeconds int             `json:"durationSeconds"`
	Nickname        string          `json:"nickname"`
	DeviceID        string          `json:"-"`
	Votes           map[string]bool `json:"-"`
	InsertedSeq     uint64          `json:"-"`
}

func (e QueueEntry) Score() int { return len(e.Votes) }

type NowPlaying struct {
	Entry      QueueEntry
	Status     Status
	StartedAt  time.Time
	Offset     float64
	Countdown  time.Time
	Generation uint64
}

func (n *NowPlaying) Elapsed(now time.Time) float64 {
	if n == nil {
		return 0
	}
	if n.Status == StatusPlaying && !n.StartedAt.IsZero() {
		v := n.Offset + now.Sub(n.StartedAt).Seconds()
		if n.Entry.DurationSeconds > 0 && v > float64(n.Entry.DurationSeconds) {
			return float64(n.Entry.DurationSeconds)
		}
		if v < 0 {
			return 0
		}
		return v
	}
	return n.Offset
}

type SkipVote struct {
	StartedAt         time.Time
	ExpiresAt         time.Time
	EligibleDevices   map[string]bool
	YesDevices        map[string]bool
	Threshold         int
	InitiatorDevice   string
	InitiatorNickname string
}

type Notice struct {
	ID        string            `json:"id"`
	Code      string            `json:"code"`
	Params    map[string]string `json:"params"`
	CreatedAt time.Time         `json:"createdAt"`
}

type Room struct {
	Code           string
	HostToken      string
	PIN            string
	CreatedAt      time.Time
	HostLastSeenAt time.Time
	QueueLocked    bool
	ManualOrder    bool
	Entries        []QueueEntry
	Now            *NowPlaying
	SkipVote       *SkipVote
	CooldownUntil  time.Time
	Notices        []Notice
	Sequence       uint64
	Generation     uint64
	TVStarted      bool
}

type DeviceInfo struct {
	LastSeen  time.Time
	Host      bool
	Connected bool
}

type VideoMeta struct {
	VideoID         string `json:"videoId"`
	Title           string `json:"title"`
	Channel         string `json:"channel"`
	ThumbnailURL    string `json:"thumbnailUrl"`
	DurationSeconds int    `json:"durationSeconds"`
	Live            bool   `json:"live"`
}

type SearchResult struct {
	VideoID         string `json:"videoId"`
	Title           string `json:"title"`
	Channel         string `json:"channel"`
	ThumbnailURL    string `json:"thumbnailUrl"`
	DurationSeconds int    `json:"durationSeconds"`
}

// Snapshot is the wire representation sent to browsers. It contains no host
// token, PIN, device secrets, or raw vote maps.
type Snapshot struct {
	Seq        uint64          `json:"seq"`
	ServerTime time.Time       `json:"serverTime"`
	Room       RoomView        `json:"room"`
	NowPlaying *NowPlayingView `json:"nowPlaying"`
	Queue      []QueueView     `json:"queue"`
	SkipVote   *SkipVoteView   `json:"skipVote"`
	Notices    []Notice        `json:"notices"`
	You        YouView         `json:"you"`
}

type RoomView struct {
	Code          string `json:"code"`
	QueueLocked   bool   `json:"queueLocked"`
	ManualOrder   bool   `json:"manualOrder"`
	GuestCount    int    `json:"guestCount"`
	GuestCapacity int    `json:"guestCapacity"`
	QueueCapacity int    `json:"queueCapacity"`
	TVStarted     bool   `json:"tvStarted"`
}

type NowPlayingView struct {
	EntryID         string     `json:"entryId"`
	VideoID         string     `json:"videoId"`
	Title           string     `json:"title"`
	Channel         string     `json:"channel"`
	ThumbnailURL    string     `json:"thumbnailUrl"`
	Nickname        string     `json:"nickname"`
	DurationSeconds int        `json:"durationSeconds"`
	Status          Status     `json:"status"`
	ElapsedSeconds  float64    `json:"elapsedSeconds"`
	Generation      uint64     `json:"generation"`
	CountdownEndsAt *time.Time `json:"countdownEndsAt"`
	Mine            bool       `json:"mine"`
}

type QueueView struct {
	ID              string `json:"id"`
	Position        int    `json:"position"`
	VideoID         string `json:"videoId"`
	Title           string `json:"title"`
	Channel         string `json:"channel"`
	ThumbnailURL    string `json:"thumbnailUrl"`
	DurationSeconds int    `json:"durationSeconds"`
	Nickname        string `json:"nickname"`
	Score           int    `json:"score"`
	VotedByMe       bool   `json:"votedByMe"`
	Mine            bool   `json:"mine"`
}

type SkipVoteView struct {
	Active        bool       `json:"active"`
	Yes           int        `json:"yes"`
	Threshold     int        `json:"threshold"`
	Eligible      int        `json:"eligible"`
	ExpiresAt     *time.Time `json:"expiresAt"`
	VotedByMe     bool       `json:"votedByMe"`
	CanVote       bool       `json:"canVote"`
	CooldownUntil *time.Time `json:"cooldownUntil"`
	Initiator     string     `json:"initiator,omitempty"`
}

type YouView struct {
	DeviceID     string `json:"deviceId"`
	Role         string `json:"role"`
	HostUnlocked bool   `json:"hostUnlocked"`
	CanSkipVote  bool   `json:"canSkipVote"`
	InQueue      bool   `json:"inQueue"`
}
