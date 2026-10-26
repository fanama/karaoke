package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/ports"
)

func TestLibrarySongsClampsPagination(t *testing.T) {
	lib := &fakeLibrary{songs: []domain.SearchResult{{VideoID: "aaaaaaaaaaa", Title: "Song"}}}
	service := &CatalogService{Deps: Deps{Library: lib}}

	page, err := service.LibrarySongs(context.Background(), ports.SongQuery{Offset: -5})
	if err != nil {
		t.Fatal(err)
	}
	if page.Offset != 0 || page.Limit != ports.DefaultSongPageSize {
		t.Fatalf("default page = %+v", page)
	}

	page, err = service.LibrarySongs(context.Background(), ports.SongQuery{Limit: 5_000})
	if err != nil || page.Limit != ports.MaxSongPageSize {
		t.Fatalf("clamped limit = %+v err=%v", page, err)
	}

	long := strings.Repeat("x", maxSongFilterLength+50)
	page, err = service.LibrarySongs(context.Background(), ports.SongQuery{Limit: 10, Query: "  " + long + "  "})
	if err != nil || len(page.Query) != maxSongFilterLength {
		t.Fatalf("trimmed filter = %q err=%v", len(page.Query), err)
	}
}

func TestLibrarySongsReportsDisabledAndBrokenLibraries(t *testing.T) {
	disabled := &CatalogService{Deps: Deps{}}
	if _, err := disabled.LibrarySongs(context.Background(), ports.SongQuery{}); !errors.Is(err, domain.ErrLibraryDisabled) {
		t.Fatalf("disabled library error = %v", err)
	}

	broken := &CatalogService{Deps: Deps{Library: &fakeLibrary{readErr: errors.New("boom")}}}
	if _, err := broken.LibrarySongs(context.Background(), ports.SongQuery{}); err == nil {
		t.Fatal("expected the read error to surface")
	}
}
