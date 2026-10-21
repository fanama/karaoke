package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"fana-karaoke/internal/domain"
)

const (
	activeWindow = 2 * time.Minute
)

type Config struct {
	MaxRooms  int
	MaxGuests int
	MaxQueue  int
	RoomTTL   time.Duration
}

type Catalog interface {
	Search(ctx context.Context, query string) ([]domain.SearchResult, bool, error)
	Lookup(ctx context.Context, videoID string) (domain.VideoMeta, error)
}

type Manager struct {
	mu          sync.Mutex
	rooms       map[string]*domain.Room
	devices     map[string]*domain.DeviceInfo
	hostTokens  map[string]string
	cfg         Config
	catalog     Catalog
	now         func() time.Time
	onChange    func(code string)
	onDelete    func(code string)
	lastCleanup time.Time
}

func NewManager(cfg Config, catalog Catalog) *Manager {
	if cfg.MaxRooms <= 0 {
		cfg.MaxRooms = 1000
	}
	if cfg.MaxGuests <= 0 {
		cfg.MaxGuests = 25
	}
	if cfg.MaxQueue <= 0 {
		cfg.MaxQueue = 100
	}
	if cfg.RoomTTL <= 0 {
		cfg.RoomTTL = 2 * time.Hour
	}
	return &Manager{
		rooms:      make(map[string]*domain.Room),
		devices:    make(map[string]*domain.DeviceInfo),
		hostTokens: make(map[string]string),
		cfg:        cfg,
		catalog:    catalog,
		now:        time.Now,
	}
}

func (m *Manager) SetCallbacks(onChange, onDelete func(string)) {
	m.onChange = onChange
	m.onDelete = onDelete
}

// SetClock is used by tests.
func (m *Manager) SetClock(now func() time.Time) { m.now = now }

func (m *Manager) Config() Config { return m.cfg }

func (m *Manager) CreateRoom() (string, string, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.rooms) >= m.cfg.MaxRooms {
		return "", "", "", domain.ErrServerFull
	}
	now := m.now()
	var code string
	for i := 0; i < 40; i++ {
		candidate := randomCode()
		if _, exists := m.rooms[candidate]; !exists {
			code = candidate
			break
		}
	}
	if code == "" {
		return "", "", "", domain.ErrServerFull
	}
	token := randomToken(32)
	pin, err := randomPIN()
	if err != nil {
		return "", "", "", err
	}
	room := &domain.Room{
		Code:           code,
		HostToken:      token,
		PIN:            pin,
		CreatedAt:      now,
		HostLastSeenAt: time.Time{},
		TVStarted:      true,
		Sequence:       1,
	}
	m.rooms[code] = room
	m.hostTokens[token] = code
	return code, token, pin, nil
}

func (m *Manager) Exists(code string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.rooms[code] != nil
}

// ResolveCode accepts a canonical code or a code with separators removed,
// e.g. both "blue-mic" and "bluemice" resolve to the room "blue-mic".
func (m *Manager) ResolveCode(input string) (string, bool) {
	input = strings.ToLower(strings.TrimSpace(input))
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.rooms[input] != nil {
		return input, true
	}
	compact := strings.NewReplacer("-", "", "_", "", " ", "").Replace(input)
	for code := range m.rooms {
		if strings.NewReplacer("-", "", "_", "", " ", "").Replace(code) == compact {
			return code, true
		}
	}
	return "", false
}

func (m *Manager) RoomByHostToken(token string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	code, ok := m.hostTokens[token]
	return code, ok
}

func (m *Manager) HostSeen(code string) {
	m.mu.Lock()
	if room := m.rooms[code]; room != nil {
		room.HostLastSeenAt = m.now()
	}
	m.mu.Unlock()
}

func (m *Manager) PIN(code string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	room := m.rooms[code]
	if room == nil {
		return "", false
	}
	return room.PIN, true
}

func (m *Manager) TouchDevice(deviceID string, host bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.touchLocked(deviceID, host)
}

func (m *Manager) touchLocked(deviceID string, host bool) {
	if deviceID == "" {
		return
	}
	info := m.devices[deviceID]
	if info == nil {
		info = &domain.DeviceInfo{}
		m.devices[deviceID] = info
	}
	info.LastSeen = m.now()
	if host {
		info.Host = true
	}
}

func (m *Manager) Snapshot(code, deviceID, role string) (*domain.Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	room := m.rooms[code]
	if room == nil {
		return nil, domain.ErrRoomNotFound
	}
	host := role == "tv" || role == "phone"
	info := m.devices[deviceID]
	if !host && (info == nil || (!info.Host && m.now().Sub(info.LastSeen) > activeWindow)) {
		if m.guestCountLocked(deviceID) >= m.cfg.MaxGuests {
			return nil, domain.ErrRoomFull
		}
	}
	m.touchLocked(deviceID, host)
	snap := m.snapshotLocked(room, deviceID, role)
	return &snap, nil
}

func (m *Manager) snapshotLocked(room *domain.Room, deviceID, role string) domain.Snapshot {
	now := m.now()
	host := role == "tv" || role == "phone"
	var nowView *domain.NowPlayingView
	if room.Now != nil {
		n := room.Now
		nv := &domain.NowPlayingView{
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
			Mine:            n.Entry.DeviceID == deviceID,
		}
		if n.Status == domain.StatusCountdown && !n.Countdown.IsZero() {
			t := n.Countdown
			nv.CountdownEndsAt = &t
		}
		nowView = nv
	}
	queue := make([]domain.QueueView, 0, len(room.Entries))
	inQueue := false
	for i, e := range room.Entries {
		if e.DeviceID == deviceID {
			inQueue = true
		}
		queue = append(queue, domain.QueueView{
			ID:              e.ID,
			Position:        i + 1,
			VideoID:         e.VideoID,
			Title:           e.Title,
			Channel:         e.Channel,
			ThumbnailURL:    e.ThumbnailURL,
			DurationSeconds: e.DurationSeconds,
			Nickname:        e.Nickname,
			Score:           e.Score(),
			VotedByMe:       e.Votes[deviceID],
			Mine:            e.DeviceID == deviceID,
		})
	}
	you := domain.YouView{
		DeviceID:     deviceID,
		Role:         role,
		HostUnlocked: host,
		CanSkipVote:  m.canStartSkipLocked(room, deviceID, now),
		InQueue:      inQueue || (room.Now != nil && room.Now.Entry.DeviceID == deviceID),
	}
	var vote *domain.SkipVoteView
	if room.SkipVote != nil {
		sv := room.SkipVote
		expires := sv.ExpiresAt
		vote = &domain.SkipVoteView{
			Active:    true,
			Yes:       len(sv.YesDevices),
			Threshold: sv.Threshold,
			Eligible:  len(sv.EligibleDevices),
			ExpiresAt: &expires,
			VotedByMe: sv.YesDevices[deviceID],
			CanVote:   sv.EligibleDevices[deviceID] && !sv.YesDevices[deviceID],
		}
	} else if now.Before(room.CooldownUntil) {
		t := room.CooldownUntil
		vote = &domain.SkipVoteView{CooldownUntil: &t}
	}
	notices := room.Notices
	if notices == nil {
		notices = []domain.Notice{}
	}
	return domain.Snapshot{
		Seq:        room.Sequence,
		ServerTime: now,
		Room: domain.RoomView{
			Code:          room.Code,
			QueueLocked:   room.QueueLocked,
			ManualOrder:   room.ManualOrder,
			GuestCount:    m.guestCountLocked(""),
			GuestCapacity: m.cfg.MaxGuests,
			QueueCapacity: m.cfg.MaxQueue,
			TVStarted:     room.TVStarted,
		},
		NowPlaying: nowView,
		Queue:      queue,
		SkipVote:   vote,
		Notices:    notices,
		You:        you,
	}
}

func (m *Manager) guestCountLocked(exclude string) int {
	now := m.now()
	count := 0
	for id, info := range m.devices {
		if id == exclude || info.Host {
			continue
		}
		if now.Sub(info.LastSeen) <= activeWindow {
			count++
		}
	}
	return count
}

func (m *Manager) canStartSkipLocked(room *domain.Room, deviceID string, now time.Time) bool {
	if room.Now == nil || room.SkipVote != nil || now.Before(room.CooldownUntil) {
		return false
	}
	for _, id := range m.eligibleDevicesLocked(room, now) {
		if id == deviceID {
			return true
		}
	}
	return false
}

func (m *Manager) eligibleDevicesLocked(room *domain.Room, now time.Time) []string {
	set := make(map[string]bool)
	for _, e := range room.Entries {
		set[e.DeviceID] = true
	}
	if room.Now != nil && room.Now.Entry.DeviceID != "" {
		set[room.Now.Entry.DeviceID] = true
	}
	out := make([]string, 0, len(set))
	for id := range set {
		if id == "" {
			continue
		}
		info := m.devices[id]
		if info != nil && now.Sub(info.LastSeen) <= activeWindow {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

func (m *Manager) notify(code string) {
	if m.onChange != nil {
		m.onChange(code)
	}
}

func (m *Manager) deleted(code string) {
	if m.onDelete != nil {
		m.onDelete(code)
	}
}

func (m *Manager) addNoticeLocked(room *domain.Room, code string, params map[string]string) {
	if params == nil {
		params = map[string]string{}
	}
	room.Notices = append([]domain.Notice{{
		ID:        newID("n_"),
		Code:      code,
		Params:    params,
		CreatedAt: m.now(),
	}}, room.Notices...)
	if len(room.Notices) > 20 {
		room.Notices = room.Notices[:20]
	}
}

func (m *Manager) sortQueueLocked(room *domain.Room) {
	if room.ManualOrder {
		return
	}
	sort.SliceStable(room.Entries, func(i, j int) bool {
		si, sj := room.Entries[i].Score(), room.Entries[j].Score()
		if si != sj {
			return si > sj
		}
		return room.Entries[i].InsertedSeq < room.Entries[j].InsertedSeq
	})
}

func cleanNickname(s string) (string, error) {
	s = strings.TrimSpace(s)
	if n := utf8.RuneCountInString(s); n < 1 || n > 24 {
		return "", domain.ErrInvalidInput
	}
	return s, nil
}

func newID(prefix string) string {
	b := make([]byte, 9)
	_, _ = rand.Read(b)
	return prefix + hex.EncodeToString(b)
}

func randomToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func randomPIN() (string, error) {
	b := make([]byte, 2)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	n := int(b[0])<<8 | int(b[1])
	return leftPad4(n % 10000), nil
}

func leftPad4(n int) string {
	s := make([]byte, 4)
	for i := 3; i >= 0; i-- {
		s[i] = byte('0' + n%10)
		n /= 10
	}
	return string(s)
}

func randomCode() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	a := adjectives[int(b[0])%len(adjectives)]
	n := nouns[int(b[1])%len(nouns)]
	return a + "-" + n
}
