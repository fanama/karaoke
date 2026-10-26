package usecase

import (
	"context"
	"errors"
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
	songs    []domain.SearchResult
	saved    []domain.SearchResult
	cached   []domain.SearchResult
	total    int
	readErr  error
	saveErr  error
	cacheErr error
}

func (f *fakeLibrary) Search(context.Context, string, int) ([]domain.SearchResult, error) {
	if f.readErr != nil {
		return nil, f.readErr
	}
	return append([]domain.SearchResult(nil), f.songs...), nil
}

func (f *fakeLibrary) Random(_ context.Context, limit int) ([]domain.SearchResult, error) {
	if f.readErr != nil {
		return nil, f.readErr
	}
	if limit <= 0 || limit > len(f.songs) {
		limit = len(f.songs)
	}
	return append([]domain.SearchResult(nil), f.songs[:limit]...), nil
}

func (f *fakeLibrary) Lookup(_ context.Context, videoID string) (domain.VideoMeta, bool) {
	for _, song := range f.songs {
		if song.VideoID != videoID {
			continue
		}
		return domain.VideoMeta{
			VideoID: song.VideoID, Title: song.Title, Channel: song.Channel,
			ThumbnailURL: song.ThumbnailURL, DurationSeconds: song.DurationSeconds,
		}, true
	}
	return domain.VideoMeta{}, false
}

func (f *fakeLibrary) Save(_ context.Context, songs []domain.SearchResult) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = append(f.saved, songs...)
	f.total += len(songs)
	return nil
}

func (f *fakeLibrary) Cache(_ context.Context, songs []domain.SearchResult) error {
	if f.cacheErr != nil {
		return f.cacheErr
	}
	f.cached = append(f.cached, songs...)
	return nil
}

func (f *fakeLibrary) Count(context.Context) (int, error) { return f.total, nil }

func (f *fakeLibrary) List(_ context.Context, query ports.SongQuery) (ports.SongPage, error) {
	if f.readErr != nil {
		return ports.SongPage{}, f.readErr
	}
	return ports.SongPage{
		Songs: append([]domain.SearchResult(nil), f.songs...), Total: len(f.songs),
		Offset: query.Offset, Limit: query.Limit, Query: query.Query,
	}, nil
}

func (f *fakeLibrary) Status(context.Context) ports.LibraryStatus {
	return ports.LibraryStatus{Backend: "fake", Connected: true, Songs: f.total}
}

func (f *fakeLibrary) Export(context.Context, ports.LibraryFormat, io.Writer) error { return nil }

func (f *fakeLibrary) Import(context.Context, ports.LibraryFormat, io.Reader) (int, error) {
	return 0, nil
}

type failingCatalog struct{ lookups int }

func (f *failingCatalog) Search(context.Context, string) ([]domain.SearchResult, bool, error) {
	return nil, false, domain.ErrSearchFailed
}

func (f *failingCatalog) Lookup(context.Context, string) (domain.VideoMeta, error) {
	f.lookups++
	return domain.VideoMeta{}, domain.ErrVideoMissing
}

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

func TestRandomSongsSamplesLibrary(t *testing.T) {
	lib := &fakeLibrary{songs: []domain.SearchResult{
		{VideoID: "aaaaaaaaaaa", Title: "First"},
		{VideoID: "bbbbbbbbbbb", Title: "Second"},
	}}
	service := &CatalogService{Deps: Deps{Library: lib}}
	results, err := service.RandomSongs(context.Background(), 6)
	if err != nil || len(results) != 2 {
		t.Fatalf("random songs = %+v err=%v", results, err)
	}
	if results, err := service.RandomSongs(context.Background(), 0); err != nil || len(results) != 2 {
		t.Fatalf("default limit = %+v err=%v", results, err)
	}
}

func TestRandomSongsWithoutLibraryIsEmpty(t *testing.T) {
	service := &CatalogService{Deps: Deps{}}
	results, err := service.RandomSongs(context.Background(), 6)
	if err != nil || results == nil || len(results) != 0 {
		t.Fatalf("disabled library = %+v err=%v", results, err)
	}
	broken := &CatalogService{Deps: Deps{Library: &fakeLibrary{readErr: io.ErrUnexpectedEOF}}}
	if results, err := broken.RandomSongs(context.Background(), 6); err != nil || len(results) != 0 {
		t.Fatalf("broken library = %+v err=%v", results, err)
	}
}

func TestAddSongUsesLibraryMetadataWithoutUpstream(t *testing.T) {
	lib := &fakeLibrary{songs: []domain.SearchResult{
		{VideoID: "aaaaaaaaaaa", Title: "Known song", Channel: "Karaoke", DurationSeconds: 200},
	}}
	catalog := &failingCatalog{}
	f := NewFacade(Deps{
		Rooms: memory.NewRoomRepo(), Devices: memory.NewDeviceRepo(), Events: noopPublisher{},
		Catalog: catalog, Library: lib, Clock: clock.NewFake(time.Now()), IDs: crypto.NewIDs(),
		Cfg: ports.Config{MaxRooms: 10, MaxGuests: 25, MaxQueue: 100, RoomTTL: time.Hour},
	})
	code := mustRoom(t, f)
	snap, err := f.AddSong(context.Background(), code, "dev-a", "guest", "Alex", "aaaaaaaaaaa", "")
	if err != nil {
		t.Fatal(err)
	}
	if snap.NowPlaying == nil || snap.NowPlaying.Title != "Known song" || snap.NowPlaying.Channel != "Karaoke" {
		t.Fatalf("now playing = %+v", snap.NowPlaying)
	}
	if catalog.lookups != 0 {
		t.Fatalf("library hit still used the upstream catalog: %d", catalog.lookups)
	}
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
	if len(lib.cached) != 1 || lib.cached[0].VideoID != "bbbbbbbbbbb" {
		t.Fatalf("online hits were not cached: %+v", lib.cached)
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

func TestAddSongRejectsFailedDurableSave(t *testing.T) {
	lib := &fakeLibrary{saveErr: domain.ErrLibraryUnavailable}
	catalog := &fakeCatalog{meta: map[string]domain.VideoMeta{
		"aaaaaaaaaaa": {VideoID: "aaaaaaaaaaa", Title: "Played song", Channel: "Karaoke", DurationSeconds: 200},
	}}
	f := NewFacade(Deps{
		Rooms: memory.NewRoomRepo(), Devices: memory.NewDeviceRepo(), Events: noopPublisher{},
		Catalog: catalog, Library: lib, Clock: clock.NewFake(time.Now()), IDs: crypto.NewIDs(),
		Cfg: ports.Config{MaxRooms: 10, MaxGuests: 25, MaxQueue: 100, RoomTTL: time.Hour},
	})
	code := mustRoom(t, f)
	_, err := f.AddSong(context.Background(), code, "dev-a", "guest", "Alex", "aaaaaaaaaaa", "")
	if !errors.Is(err, domain.ErrLibraryUnavailable) {
		t.Fatalf("add error = %v", err)
	}
	if snap := mustSnapshot(t, f, code, "dev-a", "guest"); len(snap.Queue) != 0 {
		t.Fatalf("failed save queued a song: %+v", snap.Queue)
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
