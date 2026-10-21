package domain

import (
	"sort"
	"time"
)

// EnqueueSong appends a song, keeps the queue sorted, and starts playback when
// the room is already on and idle.
func EnqueueSong(room Room, entry QueueEntry, now time.Time, maxQueue int) (Room, error) {
	if room.QueueLocked {
		return room, ErrQueueLocked
	}
	if len(room.Entries) >= maxQueue {
		return room, ErrQueueFull
	}
	room = room.Clone()
	room.Sequence++
	entry.InsertedSeq = room.Sequence
	if entry.Votes == nil {
		entry.Votes = map[string]bool{}
	}
	room.Entries = append(room.Entries, entry)
	room.Entries = SortQueue(room.Entries, room.ManualOrder)
	room = Record(room, "notice.songAdded", NoticeStrings(
		"nickname", entry.Nickname, "title", entry.Title,
	), now)
	if room.Now == nil && room.TVStarted {
		room = StartNext(room, now)
	}
	return room, nil
}

// ToggleEntryVote flips one device's upvote and re-sorts an automatic queue.
func ToggleEntryVote(room Room, entryID, deviceID string) (Room, error) {
	idx := IndexOfEntry(room.Entries, entryID)
	if idx < 0 {
		return room, ErrInvalidInput
	}
	room = room.Clone()
	entry := room.Entries[idx]
	if entry.Votes == nil {
		entry.Votes = map[string]bool{}
	}
	if entry.Votes[deviceID] {
		delete(entry.Votes, deviceID)
	} else {
		entry.Votes[deviceID] = true
	}
	room.Entries[idx] = entry
	room.Entries = SortQueue(room.Entries, room.ManualOrder)
	room.Sequence++
	return room, nil
}

// RemoveEntry deletes a queued song and records a notice.
func RemoveEntry(room Room, entryID string, now time.Time) (Room, error) {
	idx := IndexOfEntry(room.Entries, entryID)
	if idx < 0 {
		return room, ErrInvalidInput
	}
	room = room.Clone()
	title := room.Entries[idx].Title
	room.Entries = append(room.Entries[:idx], room.Entries[idx+1:]...)
	room = Record(room, "notice.songRemoved", NoticeStrings("title", title), now)
	return room, nil
}

// ReorderQueue applies a host-provided order and freezes automatic sorting.
func ReorderQueue(room Room, orderedIDs []string, now time.Time) (Room, error) {
	ordered, err := orderByID(room.Entries, orderedIDs)
	if err != nil {
		return room, err
	}
	room = room.Clone()
	room.Entries = ordered
	room.ManualOrder = true
	room = Record(room, "notice.orderLocked", nil, now)
	return room, nil
}

// ClearEntries empties the queue and reports how many songs were dropped.
func ClearEntries(room Room, now time.Time) Room {
	room = room.Clone()
	count := len(room.Entries)
	room.Entries = nil
	return Record(room, "notice.queueCleared", NoticeStrings("count", Itoa(count)), now)
}

// SetQueueLock locks or unlocks the queue and records a notice.
func SetQueueLock(room Room, locked bool, now time.Time) Room {
	room.QueueLocked = locked
	code := "notice.queueUnlocked"
	if locked {
		code = "notice.queueLocked"
	}
	return Record(room, code, nil, now)
}

// SetQueueLocked changes the lock without recording a notice.
func SetQueueLocked(room Room, locked bool) Room {
	room.QueueLocked = locked
	room.Sequence++
	return room
}

// SortQueue orders by upvotes then insertion time unless the host chose an order.
func SortQueue(entries []QueueEntry, manualOrder bool) []QueueEntry {
	if manualOrder {
		return entries
	}
	sort.SliceStable(entries, func(i, j int) bool {
		si, sj := entries[i].Score(), entries[j].Score()
		if si != sj {
			return si > sj
		}
		return entries[i].InsertedSeq < entries[j].InsertedSeq
	})
	return entries
}

// IndexOfEntry returns the queue position of an entry or -1.
func IndexOfEntry(entries []QueueEntry, id string) int {
	for i := range entries {
		if entries[i].ID == id {
			return i
		}
	}
	return -1
}

func orderByID(entries []QueueEntry, ids []string) ([]QueueEntry, error) {
	if len(ids) != len(entries) {
		return nil, ErrInvalidInput
	}
	byID := make(map[string]QueueEntry, len(entries))
	for _, e := range entries {
		byID[e.ID] = e
	}
	ordered := make([]QueueEntry, 0, len(ids))
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		e, ok := byID[id]
		if !ok || seen[id] {
			return nil, ErrInvalidInput
		}
		seen[id] = true
		ordered = append(ordered, e)
	}
	return ordered, nil
}
