package sqlite

import (
	"context"
	"testing"
)

func TestRandomSamplesLibrary(t *testing.T) {
	lib := openTestLibrary(t)
	empty, err := lib.Random(context.Background(), 6)
	if err != nil || len(empty) != 0 {
		t.Fatalf("empty library random = %+v err=%v", empty, err)
	}
	songs := make([]string, 0, 10)
	for _, id := range []string{
		"aaaaaaaaaaa", "bbbbbbbbbbb", "ccccccccccc", "ddddddddddd", "eeeeeeeeeee",
		"fffffffffff", "ggggggggggg", "hhhhhhhhhhh", "iiiiiiiiiii", "jjjjjjjjjjj",
	} {
		songs = append(songs, id)
	}
	for i, id := range songs {
		saveSongs(t, lib, song(id, "Song "+id, "Studio", 200+i))
	}
	got, err := lib.Random(context.Background(), 6)
	if err != nil || len(got) != 6 {
		t.Fatalf("random = %d songs err=%v", len(got), err)
	}
	seen := map[string]bool{}
	for _, result := range got {
		if seen[result.VideoID] {
			t.Fatalf("duplicate sample: %s", result.VideoID)
		}
		seen[result.VideoID] = true
		if result.Title == "" || result.ThumbnailURL == "" {
			t.Fatalf("incomplete sample: %+v", result)
		}
	}
	all, err := lib.Random(context.Background(), 20)
	if err != nil || len(all) != len(songs) {
		t.Fatalf("oversized limit = %d songs err=%v", len(all), err)
	}
}

func TestLookupReturnsStoredMetadata(t *testing.T) {
	lib := openTestLibrary(t)
	saveSongs(t, lib, song("aaaaaaaaaaa", "Stored Karaoke", "Studio", 210))
	meta, ok := lib.Lookup(context.Background(), "aaaaaaaaaaa")
	if !ok || meta.Title != "Stored Karaoke" || meta.Channel != "Studio" || meta.DurationSeconds != 210 {
		t.Fatalf("lookup = %+v ok=%v", meta, ok)
	}
	if _, ok := lib.Lookup(context.Background(), "zzzzzzzzzzz"); ok {
		t.Fatal("unknown song must not be found")
	}
	if _, ok := lib.Lookup(context.Background(), "bad"); ok {
		t.Fatal("invalid id must not be found")
	}
}
