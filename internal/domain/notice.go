package domain

import (
	"strconv"
	"time"
)

// MaxNotices is how many notices a room keeps.
const MaxNotices = 20

// NoticeID derives a unique-per-room id from the sequence read before the
// mutation advances it. Every mutation advances the sequence, so ids never
// repeat inside a room.
func NoticeID(room Room) string { return "n_" + Itoa(int(room.Sequence)) }

// PrependNotice adds a notice at the head of the room history and trims it.
func PrependNotice(room Room, id, code string, params map[string]string, now time.Time) Room {
	if params == nil {
		params = map[string]string{}
	}
	notice := Notice{ID: id, Code: code, Params: cloneStrings(params), CreatedAt: now}
	room.Notices = append([]Notice{notice}, room.Notices...)
	if len(room.Notices) > MaxNotices {
		room.Notices = room.Notices[:MaxNotices]
	}
	return room
}

// Record prepends a notice and advances the room sequence.
func Record(room Room, code string, params map[string]string, now time.Time) Room {
	id := NoticeID(room)
	room = PrependNotice(room, id, code, params, now)
	room.Sequence++
	return room
}

// NoticeStrings makes a params map from alternating key/value pairs.
func NoticeStrings(pairs ...string) map[string]string {
	out := map[string]string{}
	for i := 0; i+1 < len(pairs); i += 2 {
		out[pairs[i]] = pairs[i+1]
	}
	return out
}

// Itoa renders an integer for notice params.
func Itoa(n int) string { return strconv.Itoa(n) }
