package domain

import "time"

// Status is the playback state of the room's current song.
type Status string

const (
	StatusIdle      Status = "idle"
	StatusCountdown Status = "countdown"
	StatusPlaying   Status = "playing"
	StatusPaused    Status = "paused"
)

// QueueEntry is a song waiting in the room queue.
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

// Score is the number of upvotes on the entry.
func (e QueueEntry) Score() int { return len(e.Votes) }

// Clone returns a deep copy so domain functions never share mutable state.
func (e QueueEntry) Clone() QueueEntry {
	votes := make(map[string]bool, len(e.Votes))
	for k, v := range e.Votes {
		votes[k] = v
	}
	e.Votes = votes
	return e
}

// NowPlaying is the song currently on the TV.
type NowPlaying struct {
	Entry      QueueEntry
	Status     Status
	StartedAt  time.Time
	Offset     float64
	Countdown  time.Time
	Generation uint64
}

// Elapsed returns the playback position in seconds, clamped to the duration.
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

// SkipVote is an active vote to skip the current song.
type SkipVote struct {
	StartedAt         time.Time
	ExpiresAt         time.Time
	EligibleDevices   map[string]bool
	YesDevices        map[string]bool
	Threshold         int
	InitiatorDevice   string
	InitiatorNickname string
}

// Notice is a short room event shown to every client.
type Notice struct {
	ID        string            `json:"id"`
	Code      string            `json:"code"`
	Params    map[string]string `json:"params"`
	CreatedAt time.Time         `json:"createdAt"`
}

// Room is the aggregate root: queue, playback, votes, notices and host state.
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

// Clone returns a deep copy of the room and everything it owns.
func (r Room) Clone() Room {
	entries := make([]QueueEntry, len(r.Entries))
	for i, e := range r.Entries {
		entries[i] = e.Clone()
	}
	r.Entries = entries
	if r.Now != nil {
		now := *r.Now
		now.Entry = now.Entry.Clone()
		r.Now = &now
	}
	if r.SkipVote != nil {
		vote := *r.SkipVote
		vote.EligibleDevices = cloneSet(r.SkipVote.EligibleDevices)
		vote.YesDevices = cloneSet(r.SkipVote.YesDevices)
		r.SkipVote = &vote
	}
	notices := make([]Notice, len(r.Notices))
	copy(notices, r.Notices)
	for i := range notices {
		notices[i].Params = cloneStrings(notices[i].Params)
	}
	r.Notices = notices
	return r
}

// DeviceInfo tracks one browser seen by the room.
type DeviceInfo struct {
	LastSeen  time.Time
	Host      bool
	Connected bool
}

// VideoMeta is the metadata needed to queue a YouTube song.
type VideoMeta struct {
	VideoID         string `json:"videoId"`
	Title           string `json:"title"`
	Channel         string `json:"channel"`
	ThumbnailURL    string `json:"thumbnailUrl"`
	DurationSeconds int    `json:"durationSeconds"`
	Live            bool   `json:"live"`
}

// SearchResult is one karaoke search hit.
type SearchResult struct {
	VideoID         string `json:"videoId"`
	Title           string `json:"title"`
	Channel         string `json:"channel"`
	ThumbnailURL    string `json:"thumbnailUrl"`
	DurationSeconds int    `json:"durationSeconds"`
}

func cloneSet(in map[string]bool) map[string]bool {
	out := make(map[string]bool, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneStrings(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
