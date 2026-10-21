package domain

import "time"

// CountdownDuration is the gap between songs before playback starts.
const CountdownDuration = 5 * time.Second

// StartNext moves the head of the queue into the player with a countdown.
func StartNext(room Room, now time.Time) Room {
	room = room.Clone()
	room.SkipVote = nil
	room.CooldownUntil = time.Time{}
	if len(room.Entries) == 0 {
		room.Now = nil
		room.Generation++
		return room
	}
	entry := room.Entries[0]
	room.Entries = room.Entries[1:]
	room.Generation++
	room.Now = &NowPlaying{
		Entry:      entry,
		Status:     StatusCountdown,
		Offset:     0,
		Countdown:  now.Add(CountdownDuration),
		Generation: room.Generation,
	}
	room.Sequence++
	return room
}

// Play starts the current song or resumes it when paused.
func Play(room Room, now time.Time) Room {
	room.TVStarted = true
	if room.Now == nil {
		return StartNext(room, now)
	}
	switch room.Now.Status {
	case StatusPaused, StatusCountdown:
		room.Now.Status = StatusPlaying
		room.Now.StartedAt = now
		room.Now.Countdown = time.Time{}
	}
	return room
}

// Pause freezes playback at the current offset.
func Pause(room Room, now time.Time) Room {
	if room.Now == nil || room.Now.Status != StatusPlaying {
		return room
	}
	room.Now.Offset = room.Now.Elapsed(now)
	room.Now.StartedAt = time.Time{}
	room.Now.Status = StatusPaused
	return room
}

// SkipToNext records a host skip notice and advances the queue.
func SkipToNext(room Room, now time.Time) Room {
	if room.Now == nil {
		return StartNext(room, now)
	}
	room = Record(room, "notice.songSkippedHost", NoticeStrings(
		"title", room.Now.Entry.Title,
		"nickname", room.Now.Entry.Nickname,
	), now)
	return StartNext(room, now)
}

// CancelCountdown pauses the next song before it starts.
func CancelCountdown(room Room) Room {
	if room.Now == nil || room.Now.Status != StatusCountdown {
		return room
	}
	room.Now.Status = StatusPaused
	room.Now.Countdown = time.Time{}
	room.Now.Offset = 0
	room.Now.StartedAt = time.Time{}
	return room
}

// ApplyPlayback runs a host playback action.
func ApplyPlayback(room Room, action string, now time.Time) (Room, error) {
	switch action {
	case "play":
		return Play(room, now), nil
	case "pause":
		return Pause(room, now), nil
	case "next":
		return SkipToNext(room, now), nil
	case "cancelCountdown":
		return CancelCountdown(room), nil
	}
	return room, ErrInvalidInput
}

// AdvanceOnEnded advances only when the TV reports the current generation.
func AdvanceOnEnded(room Room, entryID string, generation uint64, now time.Time) (Room, bool) {
	if !isCurrent(room, entryID, generation) || room.Now.Status != StatusPlaying {
		return room, false
	}
	return StartNext(room, now), true
}

// AdvanceOnError records a failed song and moves on.
func AdvanceOnError(room Room, entryID, reason string, generation uint64, now time.Time) (Room, bool) {
	if !isCurrent(room, entryID, generation) || room.Now.Status == StatusCountdown {
		return room, false
	}
	room = Record(room, "notice.songFailed", NoticeStrings("title", room.Now.Entry.Title, "reason", reason), now)
	return StartNext(room, now), true
}

// UpdateProgress stores the TV's reported position when it drifts by 2s.
func UpdateProgress(room Room, entryID string, generation uint64, pos, dur float64, now time.Time) (Room, bool) {
	if !isCurrent(room, entryID, generation) {
		return room, false
	}
	changed := false
	if dur > 0 && room.Now.Entry.DurationSeconds == 0 {
		room.Now.Entry.DurationSeconds = int(dur)
		room.Sequence++
		changed = true
	}
	if room.Now.Status == StatusPlaying {
		if diff := pos - room.Now.Elapsed(now); diff > 2 || diff < -2 {
			room.Now.Offset = pos
			room.Now.StartedAt = now
		}
	}
	return room, changed
}

// AdvanceCountdownIfDue flips a due countdown into playback.
func AdvanceCountdownIfDue(room Room, now time.Time) (Room, bool) {
	if room.Now == nil || room.Now.Status != StatusCountdown {
		return room, false
	}
	if room.Now.Countdown.IsZero() || now.Before(room.Now.Countdown) {
		return room, false
	}
	room.Now.Status = StatusPlaying
	room.Now.StartedAt = room.Now.Countdown
	room.Now.Countdown = time.Time{}
	room.Sequence++
	return room, true
}

func isCurrent(room Room, entryID string, generation uint64) bool {
	return room.Now != nil && room.Now.Entry.ID == entryID && room.Now.Generation == generation
}
