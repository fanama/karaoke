package app

import (
	"context"
	"testing"
	"time"

	"fana-karaoke/internal/domain"
)

type fakeCatalog struct {
	meta map[string]domain.VideoMeta
}

func (f *fakeCatalog) Search(context.Context, string) ([]domain.SearchResult, bool, error) {
	return nil, false, nil
}

func (f *fakeCatalog) Lookup(_ context.Context, id string) (domain.VideoMeta, error) {
	if m, ok := f.meta[id]; ok {
		return m, nil
	}
	return domain.VideoMeta{}, domain.ErrVideoMissing
}

func newTestManager(t *testing.T) (*Manager, *time.Time) {
	t.Helper()
	now := time.Date(2026, 10, 21, 20, 0, 0, 0, time.UTC)
	catalog := &fakeCatalog{meta: map[string]domain.VideoMeta{
		"aaaaaaaaaaa": {VideoID: "aaaaaaaaaaa", Title: "First song", Channel: "Karaoke", DurationSeconds: 200},
		"bbbbbbbbbbb": {VideoID: "bbbbbbbbbbb", Title: "Second song", Channel: "Karaoke", DurationSeconds: 210},
		"ccccccccccc": {VideoID: "ccccccccccc", Title: "Third song", Channel: "Karaoke", DurationSeconds: 220},
	}}
	m := NewManager(Config{MaxRooms: 10, MaxGuests: 25, MaxQueue: 100, RoomTTL: time.Hour}, catalog)
	m.SetClock(func() time.Time { return now })
	return m, &now
}

func mustSnapshot(t *testing.T, m *Manager, code, device, role string) *domain.Snapshot {
	t.Helper()
	snap, err := m.Snapshot(code, device, role)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	return snap
}

func TestQueueSortsByVotesAndManualOrderFreezes(t *testing.T) {
	m, _ := newTestManager(t)
	ctx := context.Background()
	code, _, _, err := m.CreateRoom()
	if err != nil {
		t.Fatal(err)
	}
	first := mustSnapshot(t, m, code, "dev-a", "guest")
	if len(first.Queue) != 0 {
		t.Fatal("new room should be empty")
	}
	if _, err := m.AddSong(ctx, code, "dev-0", "guest", "Host", "ccccccccccc", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := m.AddSong(ctx, code, "dev-a", "guest", "Alex", "aaaaaaaaaaa", ""); err != nil {
		t.Fatal(err)
	}
	snap, err := m.AddSong(ctx, code, "dev-b", "guest", "Sam", "bbbbbbbbbbb", "")
	if err != nil {
		t.Fatal(err)
	}
	secondID := snap.Queue[1].ID
	if _, err := m.ToggleVote(code, "dev-c", "guest", secondID); err != nil {
		t.Fatal(err)
	}
	snap = mustSnapshot(t, m, code, "dev-a", "guest")
	if snap.Queue[0].VideoID != "bbbbbbbbbbb" {
		t.Fatalf("voted song should move first, got %s", snap.Queue[0].VideoID)
	}

	ids := []string{snap.Queue[1].ID, snap.Queue[0].ID}
	snap, err = m.Reorder(code, ids, "dev-tv", "tv")
	if err != nil {
		t.Fatal(err)
	}
	if !snap.Room.ManualOrder {
		t.Fatal("manual order should be set")
	}
	if snap.Queue[0].VideoID != "aaaaaaaaaaa" {
		t.Fatalf("host order should win, got %s", snap.Queue[0].VideoID)
	}
	if _, err := m.ToggleVote(code, "dev-d", "guest", snap.Queue[1].ID); err != nil {
		t.Fatal(err)
	}
	snap = mustSnapshot(t, m, code, "dev-a", "guest")
	if snap.Queue[0].VideoID != "aaaaaaaaaaa" {
		t.Fatal("votes must not reorder after host order")
	}
}

func TestSkipVotePassesAndAdvances(t *testing.T) {
	m, now := newTestManager(t)
	ctx := context.Background()
	code, _, _, _ := m.CreateRoom()
	if _, err := m.AddSong(ctx, code, "dev-a", "guest", "Alex", "aaaaaaaaaaa", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := m.AddSong(ctx, code, "dev-b", "guest", "Sam", "bbbbbbbbbbb", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Playback(code, "play", "dev-host", "tv"); err != nil {
		t.Fatal(err)
	}
	*now = now.Add(6 * time.Second)
	m.Tick()
	snap := mustSnapshot(t, m, code, "dev-a", "guest")
	if snap.NowPlaying == nil || snap.NowPlaying.Status != domain.StatusPlaying {
		t.Fatalf("expected playing, got %+v", snap.NowPlaying)
	}

	snap, err := m.StartSkipVote(code, "dev-a", "guest")
	if err != nil {
		t.Fatal(err)
	}
	if snap.SkipVote == nil || !snap.SkipVote.Active {
		t.Fatal("vote should be active")
	}
	if snap.SkipVote.Threshold != 2 {
		t.Fatalf("expected threshold 2 for two owners, got %d", snap.SkipVote.Threshold)
	}
	snap, err = m.CastSkipVote(code, "dev-a", "guest")
	if err != nil {
		t.Fatal(err)
	}
	if snap.SkipVote == nil || snap.SkipVote.Yes != 1 {
		t.Fatal("first vote should be recorded")
	}
	snap, err = m.CastSkipVote(code, "dev-b", "guest")
	if err != nil {
		t.Fatal(err)
	}
	if snap.SkipVote != nil {
		t.Fatal("vote should clear after passing")
	}
	if snap.NowPlaying == nil || snap.NowPlaying.Status != domain.StatusCountdown {
		t.Fatalf("expected next-song countdown, got %+v", snap.NowPlaying)
	}
}

func TestPlaybackEndsAndAdvancesAfterCountdown(t *testing.T) {
	m, now := newTestManager(t)
	ctx := context.Background()
	code, _, _, _ := m.CreateRoom()
	if _, err := m.AddSong(ctx, code, "dev-a", "guest", "Alex", "aaaaaaaaaaa", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := m.AddSong(ctx, code, "dev-b", "guest", "Sam", "bbbbbbbbbbb", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Playback(code, "play", "dev-host", "tv"); err != nil {
		t.Fatal(err)
	}
	*now = now.Add(5 * time.Second)
	m.Tick()
	snap := mustSnapshot(t, m, code, "dev-a", "guest")
	if snap.NowPlaying == nil || snap.NowPlaying.Status != domain.StatusPlaying {
		t.Fatal("countdown should turn into playback")
	}
	gen := snap.NowPlaying.Generation
	entry := snap.NowPlaying.EntryID
	m.TVEnded(code, entry, gen, "dev-tv")
	snap = mustSnapshot(t, m, code, "dev-a", "guest")
	if snap.NowPlaying == nil || snap.NowPlaying.EntryID == entry {
		t.Fatal("ended song should advance")
	}
	*now = now.Add(5 * time.Second)
	m.Tick()
	snap = mustSnapshot(t, m, code, "dev-a", "guest")
	if snap.NowPlaying.Status != domain.StatusPlaying {
		t.Fatal("second song should play")
	}
}

func TestQueueAndGuestCaps(t *testing.T) {
	m, _ := newTestManager(t)
	ctx := context.Background()
	code, _, _, _ := m.CreateRoom()
	for i := 0; i < 101; i++ {
		if _, err := m.AddSong(ctx, code, "dev-a", "guest", "Alex", "aaaaaaaaaaa", ""); err != nil {
			t.Fatalf("add %d: %v", i, err)
		}
	}
	if _, err := m.AddSong(ctx, code, "dev-a", "guest", "Alex", "aaaaaaaaaaa", ""); err != domain.ErrQueueFull {
		t.Fatalf("expected queue full, got %v", err)
	}
}
