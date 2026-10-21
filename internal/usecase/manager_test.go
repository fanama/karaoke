package usecase

import (
	"context"
	"testing"
	"time"

	"fana-karaoke/internal/adapters/secondary/clock"
	"fana-karaoke/internal/adapters/secondary/crypto"
	"fana-karaoke/internal/adapters/secondary/memory"
	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/ports"
)

type fakeCatalog struct {
	meta map[string]domain.VideoMeta
}

func (f *fakeCatalog) Search(context.Context, string) ([]domain.SearchResult, bool, error) {
	return nil, false, nil
}

func (f *fakeCatalog) Lookup(_ context.Context, id string) (domain.VideoMeta, error) {
	if meta, ok := f.meta[id]; ok {
		return meta, nil
	}
	return domain.VideoMeta{}, domain.ErrVideoMissing
}

type noopPublisher struct{}

func (noopPublisher) PublishRoomUpdated(string) {}
func (noopPublisher) PublishRoomEnded(string)   {}

func newTestFacade(t *testing.T) (*Facade, *clock.Fake) {
	t.Helper()
	clockFake := clock.NewFake(time.Date(2026, 10, 21, 20, 0, 0, 0, time.UTC))
	catalog := &fakeCatalog{meta: map[string]domain.VideoMeta{
		"aaaaaaaaaaa": {VideoID: "aaaaaaaaaaa", Title: "First song", Channel: "Karaoke", DurationSeconds: 200},
		"bbbbbbbbbbb": {VideoID: "bbbbbbbbbbb", Title: "Second song", Channel: "Karaoke", DurationSeconds: 210},
		"ccccccccccc": {VideoID: "ccccccccccc", Title: "Third song", Channel: "Karaoke", DurationSeconds: 220},
	}}
	facade := NewFacade(Deps{
		Rooms:   memory.NewRoomRepo(),
		Devices: memory.NewDeviceRepo(),
		Events:  noopPublisher{},
		Catalog: catalog,
		Clock:   clockFake,
		IDs:     crypto.NewIDs(),
		Cfg:     ports.Config{MaxRooms: 10, MaxGuests: 25, MaxQueue: 100, RoomTTL: time.Hour},
	})
	return facade, clockFake
}

func mustRoom(t *testing.T, f *Facade) string {
	t.Helper()
	code, _, _, err := f.CreateRoom(context.Background())
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	return code
}

func mustAdd(t *testing.T, f *Facade, code, device, nickname, videoID string) *domain.Snapshot {
	t.Helper()
	snap, err := f.AddSong(context.Background(), code, device, "guest", nickname, videoID, "")
	if err != nil {
		t.Fatalf("add %s: %v", videoID, err)
	}
	return snap
}

func mustSnapshot(t *testing.T, f *Facade, code, device, role string) *domain.Snapshot {
	t.Helper()
	snap, err := f.Snapshot(code, device, role)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	return snap
}

func TestQueueSortsByVotesAndManualOrderFreezes(t *testing.T) {
	f, _ := newTestFacade(t)
	code := mustRoom(t, f)
	mustAdd(t, f, code, "dev-0", "Host", "ccccccccccc")
	mustAdd(t, f, code, "dev-a", "Alex", "aaaaaaaaaaa")
	snap := mustAdd(t, f, code, "dev-b", "Sam", "bbbbbbbbbbb")
	secondID := snap.Queue[1].ID

	if _, err := f.ToggleVote(code, "dev-c", "guest", secondID); err != nil {
		t.Fatal(err)
	}
	snap = mustSnapshot(t, f, code, "dev-a", "guest")
	if snap.Queue[0].VideoID != "bbbbbbbbbbb" {
		t.Fatalf("voted song should move first, got %s", snap.Queue[0].VideoID)
	}

	ids := []string{snap.Queue[1].ID, snap.Queue[0].ID}
	snap, err := f.Reorder(code, ids, "dev-tv", "tv")
	if err != nil {
		t.Fatal(err)
	}
	if !snap.Room.ManualOrder || snap.Queue[0].VideoID != "aaaaaaaaaaa" {
		t.Fatalf("host order should win, got %+v", snap.Queue)
	}
	if _, err := f.ToggleVote(code, "dev-d", "guest", snap.Queue[1].ID); err != nil {
		t.Fatal(err)
	}
	snap = mustSnapshot(t, f, code, "dev-a", "guest")
	if snap.Queue[0].VideoID != "aaaaaaaaaaa" {
		t.Fatal("votes must not reorder after host order")
	}
}
