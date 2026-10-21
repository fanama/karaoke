package domain

import "fmt"

type AppError struct {
	Code       string
	Message    string
	Status     int
	RetryAfter int
}

func (e *AppError) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }

func Err(code, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, Status: status}
}

var (
	ErrRoomNotFound = Err("room_not_found", "Room not found.", 404)
	ErrRoomFull     = Err("room_full", "Room is full.", 409)
	ErrQueueFull    = Err("queue_full", "The queue is full.", 409)
	ErrQueueLocked  = Err("queue_locked", "The queue is locked.", 409)
	ErrInvalidInput = Err("invalid_input", "Invalid input.", 400)
	ErrHostRequired = Err("host_required", "Host access required.", 403)
	ErrInvalidPIN   = Err("invalid_pin", "Wrong PIN.", 401)
	ErrNoSong       = Err("no_song", "No song is playing.", 409)
	ErrNoEligible   = Err("no_eligible_voters", "No eligible voters in the queue.", 409)
	ErrVoteActive   = Err("vote_active", "A skip vote is already active.", 409)
	ErrVoteCooldown = Err("vote_cooldown", "Skip voting is cooling down.", 409)
	ErrNotEligible  = Err("not_eligible", "Only queued singers can vote.", 403)
	ErrAlreadyVoted = Err("already_voted", "You already voted.", 409)
	ErrNoVote       = Err("no_vote", "No skip vote is active.", 409)
	ErrVideoLive    = Err("video_live", "Live streams cannot be queued.", 422)
	ErrVideoShort   = Err("video_short", "Shorts cannot be queued.", 422)
	ErrVideoTooLong = Err("video_too_long", "Songs must be 10 minutes or shorter.", 422)
	ErrVideoMissing = Err("video_unavailable", "This video is unavailable.", 422)
	ErrEmbedBlocked = Err("video_not_embeddable", "This video does not allow embedded playback.", 422)
	ErrInvalidURL   = Err("invalid_url", "Paste a valid YouTube link.", 400)
	ErrServerFull   = Err("server_full", "The server is at capacity.", 503)
	ErrRateLimited  = Err("rate_limited", "Too many requests.", 429)
	ErrSearchFailed = Err("search_unavailable", "Search is unavailable.", 503)
)
