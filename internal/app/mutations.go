package app

import (
	"context"
	"strings"
	"time"

	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/video"
)

func (m *Manager) AddSong(ctx context.Context, code, deviceID, role, nickname, videoID, rawURL string) (*domain.Snapshot, error) {
	nickname, err := cleanNickname(nickname)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}
	if rawURL != "" {
		id, parseErr := video.ParseURL(rawURL)
		if parseErr != nil {
			return nil, parseErr
		}
		if strings.Contains(strings.ToLower(rawURL), "/shorts/") {
			return nil, domain.ErrVideoShort
		}
		videoID = id
	}
	if !video.ValidID(videoID) {
		return nil, domain.ErrInvalidURL
	}

	m.mu.Lock()
	room := m.rooms[code]
	if room == nil {
		m.mu.Unlock()
		return nil, domain.ErrRoomNotFound
	}
	if room.QueueLocked {
		m.mu.Unlock()
		return nil, domain.ErrQueueLocked
	}
	if len(room.Entries) >= m.cfg.MaxQueue {
		m.mu.Unlock()
		return nil, domain.ErrQueueFull
	}
	m.mu.Unlock()

	if m.catalog == nil {
		return nil, domain.ErrSearchFailed
	}
	meta, lookupErr := m.catalog.Lookup(ctx, videoID)
	if lookupErr != nil {
		return nil, lookupErr
	}
	if meta.Live {
		return nil, domain.ErrVideoLive
	}
	if meta.DurationSeconds > 600 {
		return nil, domain.ErrVideoTooLong
	}
	if meta.DurationSeconds > 0 && meta.DurationSeconds <= 60 {
		return nil, domain.ErrVideoShort
	}
	if meta.Title == "" {
		return nil, domain.ErrVideoMissing
	}
	if meta.ThumbnailURL == "" {
		meta.ThumbnailURL = "https://i.ytimg.com/vi/" + videoID + "/mqdefault.jpg"
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	room = m.rooms[code]
	if room == nil {
		return nil, domain.ErrRoomNotFound
	}
	if room.QueueLocked {
		return nil, domain.ErrQueueLocked
	}
	if len(room.Entries) >= m.cfg.MaxQueue {
		return nil, domain.ErrQueueFull
	}
	now := m.now()
	room.Sequence++
	entry := domain.QueueEntry{
		ID:              newID("q_"),
		VideoID:         videoID,
		Title:           meta.Title,
		Channel:         meta.Channel,
		ThumbnailURL:    meta.ThumbnailURL,
		DurationSeconds: meta.DurationSeconds,
		Nickname:        nickname,
		DeviceID:        deviceID,
		Votes:           map[string]bool{},
		InsertedSeq:     room.Sequence,
	}
	room.Entries = append(room.Entries, entry)
	m.sortQueueLocked(room)
	m.addNoticeLocked(room, "notice.songAdded", map[string]string{
		"nickname": nickname,
		"title":    meta.Title,
	})
	if room.Now == nil && room.TVStarted {
		m.startNextLocked(room, now)
	}
	m.touchLocked(deviceID, role == "tv" || role == "phone")
	snap := m.snapshotLocked(room, deviceID, role)
	go m.notify(code)
	return &snap, nil
}

func (m *Manager) ToggleVote(code, deviceID, role, entryID string) (*domain.Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	room := m.rooms[code]
	if room == nil {
		return nil, domain.ErrRoomNotFound
	}
	idx := -1
	for i := range room.Entries {
		if room.Entries[i].ID == entryID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil, domain.ErrInvalidInput
	}
	if room.Entries[idx].Votes[deviceID] {
		delete(room.Entries[idx].Votes, deviceID)
	} else {
		room.Entries[idx].Votes[deviceID] = true
	}
	m.sortQueueLocked(room)
	room.Sequence++
	m.touchLocked(deviceID, role == "tv" || role == "phone")
	snap := m.snapshotLocked(room, deviceID, role)
	go m.notify(code)
	return &snap, nil
}

func (m *Manager) StartSkipVote(code, deviceID, role string) (*domain.Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	room := m.rooms[code]
	if room == nil {
		return nil, domain.ErrRoomNotFound
	}
	m.touchLocked(deviceID, role == "tv" || role == "phone")
	if room.SkipVote != nil {
		return ptr(m.snapshotLocked(room, deviceID, role)), nil
	}
	now := m.now()
	if room.Now == nil {
		return nil, domain.ErrNoSong
	}
	if now.Before(room.CooldownUntil) {
		e := *domain.ErrVoteCooldown
		e.RetryAfter = int(room.CooldownUntil.Sub(now).Seconds()) + 1
		return nil, &e
	}
	eligible := m.eligibleDevicesLocked(room, now)
	if len(eligible) == 0 {
		return nil, domain.ErrNoEligible
	}
	set := make(map[string]bool, len(eligible))
	for _, id := range eligible {
		set[id] = true
	}
	label := m.deviceLabelLocked(room, deviceID)
	room.SkipVote = &domain.SkipVote{
		StartedAt:         now,
		ExpiresAt:         now.Add(60 * time.Second),
		EligibleDevices:   set,
		YesDevices:        map[string]bool{},
		Threshold:         len(eligible)/2 + 1,
		InitiatorDevice:   deviceID,
		InitiatorNickname: label,
	}
	room.Sequence++
	m.addNoticeLocked(room, "notice.skipVoteStarted", map[string]string{"nickname": label})
	snap := m.snapshotLocked(room, deviceID, role)
	go m.notify(code)
	return &snap, nil
}

func (m *Manager) CastSkipVote(code, deviceID, role string) (*domain.Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	room := m.rooms[code]
	if room == nil {
		return nil, domain.ErrRoomNotFound
	}
	sv := room.SkipVote
	if sv == nil {
		return nil, domain.ErrNoVote
	}
	if !sv.EligibleDevices[deviceID] {
		return nil, domain.ErrNotEligible
	}
	if sv.YesDevices[deviceID] {
		return ptr(m.snapshotLocked(room, deviceID, role)), nil
	}
	sv.YesDevices[deviceID] = true
	room.Sequence++
	if len(sv.YesDevices) >= sv.Threshold {
		title, nickname := "", ""
		if room.Now != nil {
			title = room.Now.Entry.Title
			nickname = room.Now.Entry.Nickname
		}
		room.SkipVote = nil
		room.CooldownUntil = time.Time{}
		m.addNoticeLocked(room, "notice.songSkippedVote", map[string]string{
			"title":    title,
			"nickname": nickname,
			"yes":      itoa(len(sv.YesDevices)),
			"eligible": itoa(len(sv.EligibleDevices)),
		})
		m.startNextLocked(room, m.now())
	}
	m.touchLocked(deviceID, role == "tv" || role == "phone")
	snap := m.snapshotLocked(room, deviceID, role)
	go m.notify(code)
	return &snap, nil
}

func (m *Manager) CancelSkipVote(code, role string, deviceID string) (*domain.Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	room := m.rooms[code]
	if room == nil {
		return nil, domain.ErrRoomNotFound
	}
	room.SkipVote = nil
	room.Sequence++
	snap := m.snapshotLocked(room, deviceID, role)
	go m.notify(code)
	return &snap, nil
}

func (m *Manager) Playback(code, action, deviceID, role string) (*domain.Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	room := m.rooms[code]
	if room == nil {
		return nil, domain.ErrRoomNotFound
	}
	now := m.now()
	switch action {
	case "play":
		room.TVStarted = true
		if room.Now == nil {
			m.startNextLocked(room, now)
		} else {
			switch room.Now.Status {
			case domain.StatusPaused, domain.StatusCountdown:
				room.Now.Status = domain.StatusPlaying
				room.Now.StartedAt = now
				room.Now.Countdown = time.Time{}
			}
		}
	case "pause":
		if room.Now != nil && room.Now.Status == domain.StatusPlaying {
			room.Now.Offset = room.Now.Elapsed(now)
			room.Now.StartedAt = time.Time{}
			room.Now.Status = domain.StatusPaused
		}
	case "next":
		if room.Now == nil {
			m.startNextLocked(room, now)
		} else {
			title := room.Now.Entry.Title
			nickname := room.Now.Entry.Nickname
			m.addNoticeLocked(room, "notice.songSkippedHost", map[string]string{
				"title": title, "nickname": nickname,
			})
			m.startNextLocked(room, now)
		}
	case "cancelCountdown":
		if room.Now != nil && room.Now.Status == domain.StatusCountdown {
			room.Now.Status = domain.StatusPaused
			room.Now.Countdown = time.Time{}
			room.Now.Offset = 0
			room.Now.StartedAt = time.Time{}
		}
	default:
		return nil, domain.ErrInvalidInput
	}
	room.Sequence++
	m.touchLocked(deviceID, role == "tv" || role == "phone")
	snap := m.snapshotLocked(room, deviceID, role)
	go m.notify(code)
	return &snap, nil
}

func (m *Manager) RemoveSong(code, entryID, deviceID, role string) (*domain.Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	room := m.rooms[code]
	if room == nil {
		return nil, domain.ErrRoomNotFound
	}
	idx := -1
	for i := range room.Entries {
		if room.Entries[i].ID == entryID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil, domain.ErrInvalidInput
	}
	title := room.Entries[idx].Title
	room.Entries = append(room.Entries[:idx], room.Entries[idx+1:]...)
	m.addNoticeLocked(room, "notice.songRemoved", map[string]string{"title": title})
	room.Sequence++
	snap := m.snapshotLocked(room, deviceID, role)
	go m.notify(code)
	return &snap, nil
}

func (m *Manager) Reorder(code string, ids []string, deviceID, role string) (*domain.Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	room := m.rooms[code]
	if room == nil {
		return nil, domain.ErrRoomNotFound
	}
	if len(ids) != len(room.Entries) {
		return nil, domain.ErrInvalidInput
	}
	byID := make(map[string]domain.QueueEntry, len(room.Entries))
	for _, e := range room.Entries {
		byID[e.ID] = e
	}
	ordered := make([]domain.QueueEntry, 0, len(ids))
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		e, ok := byID[id]
		if !ok || seen[id] {
			return nil, domain.ErrInvalidInput
		}
		seen[id] = true
		ordered = append(ordered, e)
	}
	room.Entries = ordered
	room.ManualOrder = true
	m.addNoticeLocked(room, "notice.orderLocked", nil)
	room.Sequence++
	snap := m.snapshotLocked(room, deviceID, role)
	go m.notify(code)
	return &snap, nil
}

func (m *Manager) ClearQueue(code, deviceID, role string) (*domain.Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	room := m.rooms[code]
	if room == nil {
		return nil, domain.ErrRoomNotFound
	}
	count := len(room.Entries)
	room.Entries = nil
	m.addNoticeLocked(room, "notice.queueCleared", map[string]string{"count": itoa(count)})
	room.Sequence++
	snap := m.snapshotLocked(room, deviceID, role)
	go m.notify(code)
	return &snap, nil
}

func (m *Manager) LockQueue(code string, locked bool, deviceID, role string) (*domain.Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	room := m.rooms[code]
	if room == nil {
		return nil, domain.ErrRoomNotFound
	}
	room.QueueLocked = locked
	if locked {
		m.addNoticeLocked(room, "notice.queueLocked", nil)
	} else {
		m.addNoticeLocked(room, "notice.queueUnlocked", nil)
	}
	room.Sequence++
	snap := m.snapshotLocked(room, deviceID, role)
	go m.notify(code)
	return &snap, nil
}

func (m *Manager) RotatePIN(code string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	room := m.rooms[code]
	if room == nil {
		return "", domain.ErrRoomNotFound
	}
	pin, err := randomPIN()
	if err != nil {
		return "", err
	}
	room.PIN = pin
	return pin, nil
}

func (m *Manager) EndRoom(code, deviceID, role string) error {
	m.mu.Lock()
	room := m.rooms[code]
	if room == nil {
		m.mu.Unlock()
		return domain.ErrRoomNotFound
	}
	delete(m.rooms, code)
	delete(m.hostTokens, room.HostToken)
	m.mu.Unlock()
	m.deleted(code)
	return nil
}

func (m *Manager) TVEnded(code, entryID string, generation uint64, deviceID string) {
	m.mu.Lock()
	room := m.rooms[code]
	if room == nil || room.Now == nil {
		m.mu.Unlock()
		return
	}
	if room.Now.Entry.ID != entryID || room.Now.Generation != generation {
		m.mu.Unlock()
		return
	}
	if room.Now.Status != domain.StatusPlaying {
		m.mu.Unlock()
		return
	}
	m.startNextLocked(room, m.now())
	room.Sequence++
	go m.notify(code)
	m.mu.Unlock()
}

func (m *Manager) TVError(code, entryID, reason string, generation uint64) {
	m.mu.Lock()
	room := m.rooms[code]
	if room == nil || room.Now == nil {
		m.mu.Unlock()
		return
	}
	if room.Now.Entry.ID != entryID || room.Now.Generation != generation {
		m.mu.Unlock()
		return
	}
	if room.Now.Status == domain.StatusCountdown {
		m.mu.Unlock()
		return
	}
	title := room.Now.Entry.Title
	m.addNoticeLocked(room, "notice.songFailed", map[string]string{
		"title": title, "reason": reason,
	})
	m.startNextLocked(room, m.now())
	room.Sequence++
	go m.notify(code)
	m.mu.Unlock()
}

func (m *Manager) TVProgress(code, entryID string, generation uint64, position float64, duration float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	room := m.rooms[code]
	if room == nil || room.Now == nil || room.Now.Entry.ID != entryID || room.Now.Generation != generation {
		return
	}
	changed := false
	if duration > 0 && room.Now.Entry.DurationSeconds == 0 {
		room.Now.Entry.DurationSeconds = int(duration)
		room.Sequence++
		changed = true
	}
	if room.Now.Status == domain.StatusPlaying {
		if diff := position - room.Now.Elapsed(m.now()); diff > 2 || diff < -2 {
			room.Now.Offset = position
			room.Now.StartedAt = m.now()
		}
	}
	if changed {
		go m.notify(code)
	}
}

func (m *Manager) deviceLabelLocked(room *domain.Room, deviceID string) string {
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

func (m *Manager) startNextLocked(room *domain.Room, now time.Time) {
	room.SkipVote = nil
	room.CooldownUntil = time.Time{}
	if len(room.Entries) == 0 {
		room.Now = nil
		room.Generation++
		return
	}
	entry := room.Entries[0]
	room.Entries = room.Entries[1:]
	room.Generation++
	room.Now = &domain.NowPlaying{
		Entry:      entry,
		Status:     domain.StatusCountdown,
		Offset:     0,
		Countdown:  now.Add(5 * time.Second),
		Generation: room.Generation,
	}
	room.Sequence++
}

func ptr[T any](v T) *T { return &v }

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
