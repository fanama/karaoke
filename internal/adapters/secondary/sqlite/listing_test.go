package sqlite

import (
	"context"
	"testing"

	"fana-karaoke/internal/ports"
)

func seedLibrary(t *testing.T, titles ...string) *Library {
	t.Helper()
	lib := openTestLibrary(t)
	for i, title := range titles {
		id := string(rune('a'+i)) + "aaaaaaaaaa"
		saveSongs(t, lib, song(id, title, "Studio", 200))
	}
	return lib
}

func TestListOrdersByTitleAndPages(t *testing.T) {
	lib := seedLibrary(t, "Zebra Song", "apple Song", "Mango Song", "Banana Song")
	first, err := lib.List(context.Background(), ports.SongQuery{Offset: 0, Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if first.Total != 4 || len(first.Songs) != 2 {
		t.Fatalf("first page = %+v", first)
	}
	if first.Songs[0].Title != "apple Song" || first.Songs[1].Title != "Banana Song" {
		t.Fatalf("order = %v, %v", first.Songs[0].Title, first.Songs[1].Title)
	}
	second, err := lib.List(context.Background(), ports.SongQuery{Offset: 2, Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Songs) != 2 || second.Songs[0].Title != "Mango Song" || second.Songs[1].Title != "Zebra Song" {
		t.Fatalf("second page = %+v", second.Songs)
	}
	empty, err := lib.List(context.Background(), ports.SongQuery{Offset: 99, Limit: 2})
	if err != nil || len(empty.Songs) != 0 || empty.Total != 4 {
		t.Fatalf("past the end = %+v err=%v", empty, err)
	}
}

func TestListFiltersAndCountsMatches(t *testing.T) {
	lib := seedLibrary(t, "Céline Dion – My Heart", "Bohemian Rhapsody", "Another Love")
	page, err := lib.List(context.Background(), ports.SongQuery{Offset: 0, Limit: 25, Query: "love"})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || len(page.Songs) != 1 || page.Query != "love" {
		t.Fatalf("filtered page = %+v", page)
	}
	plain, err := lib.List(context.Background(), ports.SongQuery{Offset: 0, Limit: 25, Query: "  "})
	if err != nil || plain.Total != 3 {
		t.Fatalf("blank filter = %+v err=%v", plain, err)
	}
}
