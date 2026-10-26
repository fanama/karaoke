package d1

import (
	"context"
	"testing"

	"fana-karaoke/internal/ports"
)

func TestListPagesTheWarmCache(t *testing.T) {
	worker := &fakeWorker{songs: []apiSong{
		sampleSong("aaaaaaaaaaa", "Zebra Song"),
		sampleSong("bbbbbbbbbbb", "apple Song"),
		sampleSong("ccccccccccc", "Mango Song"),
	}}
	lib, _ := testLibrary(t, worker)
	if err := lib.Warm(context.Background()); err != nil {
		t.Fatal(err)
	}

	first, err := lib.List(context.Background(), ports.SongQuery{Offset: 0, Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if first.Total != 3 || len(first.Songs) != 2 {
		t.Fatalf("first page = %+v", first)
	}
	if first.Songs[0].Title != "apple Song" || first.Songs[1].Title != "Mango Song" {
		t.Fatalf("order = %v, %v", first.Songs[0].Title, first.Songs[1].Title)
	}

	second, err := lib.List(context.Background(), ports.SongQuery{Offset: 2, Limit: 2})
	if err != nil || len(second.Songs) != 1 || second.Songs[0].Title != "Zebra Song" {
		t.Fatalf("second page = %+v err=%v", second.Songs, err)
	}
}

func TestListFiltersAndToleratesAnEmptyCache(t *testing.T) {
	worker := &fakeWorker{songs: []apiSong{sampleSong("aaaaaaaaaaa", "Another Love")}}
	lib, _ := testLibrary(t, worker)
	if err := lib.Warm(context.Background()); err != nil {
		t.Fatal(err)
	}
	filtered, err := lib.List(context.Background(), ports.SongQuery{Offset: 0, Limit: 25, Query: "love"})
	if err != nil || filtered.Total != 1 || len(filtered.Songs) != 1 {
		t.Fatalf("filtered = %+v err=%v", filtered, err)
	}
	missed, err := lib.List(context.Background(), ports.SongQuery{Offset: 0, Limit: 25, Query: "nothing"})
	if err != nil || missed.Total != 0 || len(missed.Songs) != 0 {
		t.Fatalf("no match = %+v err=%v", missed, err)
	}

	empty, _ := testLibrary(t, &fakeWorker{})
	if page, err := empty.List(context.Background(), ports.SongQuery{Offset: 0, Limit: 25}); err != nil || page.Total != 0 {
		t.Fatalf("cold cache = %+v err=%v", page, err)
	}
}
