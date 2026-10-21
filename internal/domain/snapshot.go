package domain

import (
	"sort"
	"time"
)

// ProjectSnapshot renders the wire snapshot for one viewer.
func ProjectSnapshot(room Room, guestCount int, caps Caps, viewer Viewer, now time.Time) Snapshot {
	return Snapshot{
		Seq:        room.Sequence,
		ServerTime: now,
		Room:       projectRoom(room, guestCount, caps),
		NowPlaying: projectNowPlaying(room, viewer, now),
		Queue:      projectQueue(room, viewer),
		SkipVote:   projectVote(room, viewer, now),
		Notices:    projectNotices(room),
		You:        projectYou(room, viewer),
	}
}

func projectRoom(room Room, guestCount int, caps Caps) RoomView {
	return RoomView{
		Code:          room.Code,
		QueueLocked:   room.QueueLocked,
		ManualOrder:   room.ManualOrder,
		GuestCount:    guestCount,
		GuestCapacity: caps.Guests,
		QueueCapacity: caps.Queue,
		TVStarted:     room.TVStarted,
	}
}

func projectNowPlaying(room Room, viewer Viewer, now time.Time) *NowPlayingView {
	if room.Now == nil {
		return nil
	}
	n := room.Now
	view := &NowPlayingView{
		EntryID:         n.Entry.ID,
		VideoID:         n.Entry.VideoID,
		Title:           n.Entry.Title,
		Channel:         n.Entry.Channel,
		ThumbnailURL:    n.Entry.ThumbnailURL,
		Nickname:        n.Entry.Nickname,
		DurationSeconds: n.Entry.DurationSeconds,
		Status:          n.Status,
		ElapsedSeconds:  n.Elapsed(now),
		Generation:      n.Generation,
		Mine:            n.Entry.DeviceID == viewer.DeviceID,
	}
	if n.Status == StatusCountdown && !n.Countdown.IsZero() {
		t := n.Countdown
		view.CountdownEndsAt = &t
	}
	return view
}

func projectQueue(room Room, viewer Viewer) []QueueView {
	queue := make([]QueueView, 0, len(room.Entries))
	for i, e := range room.Entries {
		queue = append(queue, QueueView{
			ID:              e.ID,
			Position:        i + 1,
			VideoID:         e.VideoID,
			Title:           e.Title,
			Channel:         e.Channel,
			ThumbnailURL:    e.ThumbnailURL,
			DurationSeconds: e.DurationSeconds,
			Nickname:        e.Nickname,
			Score:           e.Score(),
			VotedByMe:       e.Votes[viewer.DeviceID],
			Mine:            e.DeviceID == viewer.DeviceID,
		})
	}
	return queue
}

func projectVote(room Room, viewer Viewer, now time.Time) *SkipVoteView {
	if room.SkipVote != nil {
		vote := room.SkipVote
		expires := vote.ExpiresAt
		return &SkipVoteView{
			Active:    true,
			Yes:       len(vote.YesDevices),
			Threshold: vote.Threshold,
			Eligible:  len(vote.EligibleDevices),
			ExpiresAt: &expires,
			VotedByMe: vote.YesDevices[viewer.DeviceID],
			CanVote:   vote.EligibleDevices[viewer.DeviceID] && !vote.YesDevices[viewer.DeviceID],
		}
	}
	if now.Before(room.CooldownUntil) {
		until := room.CooldownUntil
		return &SkipVoteView{CooldownUntil: &until}
	}
	return nil
}

func projectNotices(room Room) []Notice {
	if room.Notices == nil {
		return []Notice{}
	}
	return room.Notices
}

func projectYou(room Room, viewer Viewer) YouView {
	inQueue := false
	for _, e := range room.Entries {
		if e.DeviceID == viewer.DeviceID {
			inQueue = true
			break
		}
	}
	inQueue = inQueue || (room.Now != nil && room.Now.Entry.DeviceID == viewer.DeviceID)
	return YouView{
		DeviceID:     viewer.DeviceID,
		Role:         viewer.Role,
		HostUnlocked: viewer.IsHost(),
		CanSkipVote:  CanStartSkipVote(room, viewer.DeviceID),
		InQueue:      inQueue,
	}
}

// CanStartSkipVote reports whether a viewer may open a vote right now.
func CanStartSkipVote(room Room, deviceID string) bool {
	if room.Now == nil || room.SkipVote != nil {
		return false
	}
	for _, id := range EligibleDevices(room) {
		if id == deviceID {
			return true
		}
	}
	return false
}

// SortStrings sorts a copy-free slice of ids in place.
func SortStrings(values []string) { sort.Strings(values) }
