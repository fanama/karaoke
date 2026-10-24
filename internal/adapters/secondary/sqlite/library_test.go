package sqlite

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/ports"
)

func openTestLibrary(t *testing.T) *Library {
	t.Helper()
	lib, err := Open(filepath.Join(t.TempDir(), "nested", "library.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = lib.Close() })
	return lib
}

func saveSongs(t *testing.T, lib *Library, songs ...domain.SearchResult) {
	t.Helper()
	if err := lib.Save(context.Background(), songs); err != nil {
		t.Fatal(err)
	}
}

func song(id, title, channel string, duration int) domain.SearchResult {
	return domain.SearchResult{
		VideoID:         id,
		Title:           title,
		Channel:         channel,
		ThumbnailURL:    domain.ThumbnailURL(id),
		DurationSeconds: duration,
	}
}

func TestSaveDeduplicatesAndRefreshes(t *testing.T) {
	lib := openTestLibrary(t)
	original := song("aaaaaaaaaaa", "Hello Karaoke", "Studio", 200)
	saveSongs(t, lib, original, original)
	updated := original
	updated.Title = "Hello (Karaoke Version)"
	saveSongs(t, lib, updated)
	total, err := lib.Count(context.Background())
	if err != nil || total != 1 {
		t.Fatalf("count = %d err=%v, want 1", total, err)
	}
	found, err := lib.Search(context.Background(), "hello version", 20)
	if err != nil || len(found) != 1 || found[0].Title != updated.Title {
		t.Fatalf("search = %+v err=%v", found, err)
	}
}

func TestSearchFoldsAccentsAndRequiresEveryWord(t *testing.T) {
	lib := openTestLibrary(t)
	saveSongs(t, lib,
		song("aaaaaaaaaaa", "Été – Toto (karaoké version)", "Studio", 200),
		song("bbbbbbbbbbb", "Bohemian Rhapsody Karaoke", "Queen", 300),
	)
	cases := map[string]int{
		"ete":             1,
		"ETE KARAOKE":     1,
		"ete queen":       0,
		"été":             1,
		"bohemian karaok": 1,
		"toto":            1,
		"unknown":         0,
	}
	for query, want := range cases {
		found, err := lib.Search(context.Background(), query, 20)
		if err != nil {
			t.Fatalf("%q: %v", query, err)
		}
		if len(found) != want {
			t.Fatalf("%q: got %d results, want %d", query, len(found), want)
		}
	}
}

func TestSearchEscapesWildcards(t *testing.T) {
	lib := openTestLibrary(t)
	saveSongs(t, lib,
		song("aaaaaaaaaaa", "100% Karaoke", "Studio", 200),
		song("bbbbbbbbbbb", "Hello Karaoke", "Studio", 200),
	)
	found, err := lib.Search(context.Background(), "100%", 20)
	if err != nil || len(found) != 1 || found[0].VideoID != "aaaaaaaaaaa" {
		t.Fatalf("wildcard search = %+v err=%v", found, err)
	}
}

func TestExportImportRoundTrip(t *testing.T) {
	source := openTestLibrary(t)
	saveSongs(t, source,
		song("aaaaaaaaaaa", "First Karaoke", "A", 200),
		song("bbbbbbbbbbb", "Second Karaoke", "B", 210),
	)
	var archive bytes.Buffer
	if err := source.Export(context.Background(), ports.FormatSQLite, &archive); err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(archive.Bytes(), []byte("SQLite format 3\x00")) {
		t.Fatalf("export is not a SQLite file: %q", archive.Bytes()[:16])
	}
	target := openTestLibrary(t)
	imported, err := target.Import(context.Background(), ports.FormatSQLite, bytes.NewReader(archive.Bytes()))
	if err != nil || imported != 2 {
		t.Fatalf("imported = %d err=%v", imported, err)
	}
	if _, err := target.Import(context.Background(), ports.FormatSQLite, bytes.NewReader(archive.Bytes())); err != nil {
		t.Fatal(err)
	}
	total, err := target.Count(context.Background())
	if err != nil || total != 2 {
		t.Fatalf("import must not duplicate: count=%d err=%v", total, err)
	}
	found, err := target.Search(context.Background(), "second", 20)
	if err != nil || len(found) != 1 {
		t.Fatalf("imported song is not searchable: %+v err=%v", found, err)
	}
}

func TestSQLExportImportRoundTrip(t *testing.T) {
	source := openTestLibrary(t)
	saveSongs(t, source, song("aaaaaaaaaaa", "L'été Karaoke", "Studio", 200))
	var dump bytes.Buffer
	if err := source.Export(context.Background(), ports.FormatSQL, &dump); err != nil {
		t.Fatal(err)
	}
	target := openTestLibrary(t)
	imported, err := target.Import(context.Background(), ports.FormatSQL, bytes.NewReader(dump.Bytes()))
	if err != nil || imported != 1 {
		t.Fatalf("SQL import = %d %v", imported, err)
	}
	found, _ := target.Search(context.Background(), "ete", 20)
	if len(found) != 1 {
		t.Fatalf("SQL-imported song = %+v", found)
	}
}

func TestImportRejectsInvalidFile(t *testing.T) {
	lib := openTestLibrary(t)
	_, err := lib.Import(context.Background(), ports.FormatSQLite, strings.NewReader("not a database"))
	if !errors.Is(err, domain.ErrLibraryInvalid) {
		t.Fatalf("invalid import error = %v", err)
	}
}
