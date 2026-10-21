package domain

import "time"

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

// RoomView is the public room state.
type RoomView struct {
	Code          string `json:"code"`
	QueueLocked   bool   `json:"queueLocked"`
	ManualOrder   bool   `json:"manualOrder"`
	GuestCount    int    `json:"guestCount"`
	GuestCapacity int    `json:"guestCapacity"`
	QueueCapacity int    `json:"queueCapacity"`
	TVStarted     bool   `json:"tvStarted"`
}

// NowPlayingView is the public playback state.
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

// QueueView is one entry of the public queue.
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

// SkipVoteView is the public skip-vote state.
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

// YouView is the per-viewer slice of a snapshot.
type YouView struct {
	DeviceID     string `json:"deviceId"`
	Role         string `json:"role"`
	HostUnlocked bool   `json:"hostUnlocked"`
	CanSkipVote  bool   `json:"canSkipVote"`
	InQueue      bool   `json:"inQueue"`
}

// Viewer identifies who a snapshot is rendered for.
type Viewer struct {
	DeviceID string
	Role     string
}

// IsHost reports whether the viewer holds host powers on the TV or a phone.
func (v Viewer) IsHost() bool { return v.Role == "tv" || v.Role == "phone" }

// Caps are the room capacities shown in a snapshot.
type Caps struct {
	Guests int
	Queue  int
}
