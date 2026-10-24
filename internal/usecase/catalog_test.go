package usecase

import (
	"context"
	"io"
	"testing"
	"time"

	"fana-karaoke/internal/adapters/secondary/clock"
	"fana-karaoke/internal/adapters/secondary/crypto"
	"fana-karaoke/internal/adapters/secondary/memory"
	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/ports"
)

type fakeLibrary struct {
	songs   []domain.SearchResult
	saved   []domain.SearchResult
	total   int
	readErr error
}

func (f *fakeLibrary) Search(context.Context, string, int) ([]domain.SearchResult, error) {
	if f.readErr != nil {
		return nil, f.readErr
	}
	return append([]domain.SearchResult(nil), f.songs...), nil
}

func (f *fakeLibrary) Save(_ context.Context, songs []domain.SearchResult) error {
	f.saved = append(f.saved, songs...)
	f.total += len(songs)
	return nil
}

func (f *fakeLibrary) Count(context.Context) (int, error) { return f.total, nil }

func (f *fakeLibrary) Export(context.Context, io.Writer) error { return nil }

func (f *fakeLibrary) Import(context.Context, io.Reader) (int, error) { return 0, nil }

type countingCatalog struct {
	results []domain.SearchResult
	hits    int
}

func (c *countingCatalog) Search(context.Context, string) ([]domain.SearchResult, bool, error) {
	c.hits++
	return c.results, false, nil
}

func (c *countingCatalog) Lookup(context.Context, string) (domain.VideoMeta, error) {
	return domain.VideoMeta{}, domain.ErrVideoMissing
}

func TestSearchUsesLibraryBeforeUpstream(t *testing.T) {
	lib := &fakeLibrary{songs: []domain.SearchResult{{VideoID: "aaaaaaaaaaa", Title: "Local hit"}}}
	upstream := &countingCatalog{results: []domain.SearchResult{{VideoID: "bbbbbbbbbbb", Title: "Online hit"}}}
	service := &CatalogService{Deps: Deps{Library: lib, Catalog: upstream}}

	results, stale, err := service.Search(context.Background(), "hello")
	if err != nil || stale || len(results) != 1 || results[0].Title != "Local hit" {
		t.Fatalf("local search = %+v stale=%v err=%v", results, stale, err)
	}
	if upstream.hits != 0 {
		t.Fatalf("upstream queried despite a local hit: %d", upstream.hits)
	}
}

func TestSearchFallsBackOnlineAndRemembersHits(t *testing.T) {
	lib := &fakeLibrary{}
	online := []domain.SearchResult{{VideoID: "bbbbbbbbbbb", Title: "Online hit"}}
	upstream := &countingCatalog{results: online}
	service := &CatalogService{Deps: Deps{Library: lib, Catalog: upstream}}

	results, _, err := service.Search(context.Background(), "hello")
	if err != nil || len(results) != 1 || upstream.hits != 1 {
		t.Fatalf("online search = %+v hits=%d err=%v", results, upstream.hits, err)
	}
	if len(lib.saved) != 1 || lib.saved[0].VideoID != "bbbbbbbbbbb" {
		t.Fatalf("online hits were not saved: %+v", lib.saved)
	}
}

func TestSearchIgnoresLibraryFailures(t *testing.T) {
	lib := &fakeLibrary{readErr: io.ErrUnexpectedEOF}
	upstream := &countingCatalog{results: []domain.SearchResult{{VideoID: "bbbbbbbbbbb"}}}
	service := &CatalogService{Deps: Deps{Library: lib, Catalog: upstream}}
	if _, _, err := service.Search(context.Background(), "hello"); err != nil {
		t.Fatalf("a broken library must not break search: %v", err)
	}
}

func TestAddSongSavesSongToLibrary(t *testing.T) {
	lib := &fakeLibrary{}
	catalog := &fakeCatalog{meta: map[string]domain.VideoMeta{
		"aaaaaaaaaaa": {VideoID: "aaaaaaaaaaa", Title: "Played song", Channel: "Karaoke", DurationSeconds: 200},
	}}
	f := NewFacade(Deps{
		Rooms:   memory.NewRoomRepo(),
		Devices: memory.NewDeviceRepo(),
		Events:  noopPublisher{},
		Catalog: catalog,
		Library: lib,
		Clock:   clock.NewFake(time.Date(2026, 10, 21, 20, 0, 0, 0, time.UTC)),
		IDs:     crypto.NewIDs(),
		Cfg:     ports.Config{MaxRooms: 10, MaxGuests: 25, MaxQueue: 100, RoomTTL: time.Hour},
	})
	code := mustRoom(t, f)
	mustAdd(t, f, code, "dev-a", "Alex", "aaaaaaaaaaa")
	if len(lib.saved) != 1 || lib.saved[0].VideoID != "aaaaaaaaaaa" || lib.saved[0].Title != "Played song" {
		t.Fatalf("played song was not saved: %+v", lib.saved)
	}
}
