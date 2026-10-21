package domain

import "time"

// SkipVoteDuration is how long a vote stays open.
const SkipVoteDuration = 60 * time.Second

// SkipCooldown is how long voters must wait after a failed vote.
const SkipCooldown = 30 * time.Second

// VoteRetryError is a cooldown error carrying the remaining seconds.
func VoteRetryError(remaining time.Duration) *AppError {
	e := *ErrVoteCooldown
	e.RetryAfter = int(remaining.Seconds()) + 1
	return &e
}

// StartSkipVote opens a vote when eligible devices exist.
func StartSkipVote(room Room, initiator string, eligible []string, now time.Time) (Room, error) {
	if room.Now == nil {
		return room, ErrNoSong
	}
	if len(eligible) == 0 {
		return room, ErrNoEligible
	}
	if now.Before(room.CooldownUntil) {
		return room, VoteRetryError(room.CooldownUntil.Sub(now))
	}
	set := make(map[string]bool, len(eligible))
	for _, id := range eligible {
		set[id] = true
	}
	nickname := DeviceLabel(room, initiator)
	room = room.Clone()
	room.SkipVote = &SkipVote{
		StartedAt:         now,
		ExpiresAt:         now.Add(SkipVoteDuration),
		EligibleDevices:   set,
		YesDevices:        map[string]bool{},
		Threshold:         len(eligible)/2 + 1,
		InitiatorDevice:   initiator,
		InitiatorNickname: nickname,
	}
	room = Record(room, "notice.skipVoteStarted", NoticeStrings("nickname", nickname), now)
	return room, nil
}

// CastSkipVote records one ballot and advances when the threshold is reached.
type CastResult struct {
	Room     Room
	Recorded bool
}

// CastVote validates and records a ballot, skipping the song on success.
func CastVote(room Room, deviceID string, now time.Time) (CastResult, error) {
	vote := room.SkipVote
	if vote == nil {
		return CastResult{Room: room}, ErrNoVote
	}
	if !vote.EligibleDevices[deviceID] {
		return CastResult{Room: room}, ErrNotEligible
	}
	if vote.YesDevices[deviceID] {
		return CastResult{Room: room}, nil
	}
	room = room.Clone()
	vote = room.SkipVote
	vote.YesDevices[deviceID] = true
	room.Sequence++
	if len(vote.YesDevices) < vote.Threshold {
		return CastResult{Room: room, Recorded: true}, nil
	}
	title, nickname := "", ""
	if room.Now != nil {
		title, nickname = room.Now.Entry.Title, room.Now.Entry.Nickname
	}
	room.SkipVote = nil
	room.CooldownUntil = time.Time{}
	room = Record(room, "notice.songSkippedVote", NoticeStrings(
		"title", title,
		"nickname", nickname,
		"yes", Itoa(len(vote.YesDevices)),
		"eligible", Itoa(len(vote.EligibleDevices)),
	), now)
	room = StartNext(room, now)
	return CastResult{Room: room, Recorded: true}, nil
}

// CancelSkipVote drops an active vote.
func CancelSkipVote(room Room) Room {
	room.SkipVote = nil
	room.Sequence++
	return room
}

// ExpireSkipVoteIfDue closes an expired vote and starts its cooldown.
func ExpireSkipVoteIfDue(room Room, now time.Time) (Room, bool) {
	vote := room.SkipVote
	if vote == nil || now.Before(vote.ExpiresAt) {
		return room, false
	}
	yes, threshold := len(vote.YesDevices), vote.Threshold
	room.SkipVote = nil
	room.CooldownUntil = now.Add(SkipCooldown)
	room = Record(room, "notice.skipVoteFailed", NoticeStrings(
		"yes", Itoa(yes), "threshold", Itoa(threshold),
	), now)
	return room, true
}

// ClearCooldownIfDue resets an elapsed cooldown window.
func ClearCooldownIfDue(room Room, now time.Time) (Room, bool) {
	if room.CooldownUntil.IsZero() || now.Before(room.CooldownUntil) {
		return room, false
	}
	room.CooldownUntil = time.Time{}
	room.Sequence++
	return room, true
}

// DeviceLabel is how a device is named in notices.
func DeviceLabel(room Room, deviceID string) string {
	if room.Now != nil && room.Now.Entry.DeviceID == deviceID {
		return room.Now.Entry.Nickname
	}
	for _, e := range room.Entries {
		if e.DeviceID == deviceID {
			return e.Nickname
		}
	}
	return "Guest"
}

// EligibleDevices lists connected singers allowed to vote: everyone with a
// queued song plus the current singer, in a stable order.
func EligibleDevices(room Room) []string {
	set := make(map[string]bool)
	for _, e := range room.Entries {
		set[e.DeviceID] = true
	}
	if room.Now != nil && room.Now.Entry.DeviceID != "" {
		set[room.Now.Entry.DeviceID] = true
	}
	out := make([]string, 0, len(set))
	for id := range set {
		if id != "" {
			out = append(out, id)
		}
	}
	SortStrings(out)
	return out
}
