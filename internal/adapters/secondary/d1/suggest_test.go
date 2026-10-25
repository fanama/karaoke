package d1

import (
	"context"
	"testing"
)

func TestRandomSamplesWarmCache(t *testing.T) {
	worker := &fakeWorker{songs: []apiSong{
		sampleSong("aaaaaaaaaaa", "First Song"),
		sampleSong("bbbbbbbbbbb", "Second Song"),
		sampleSong("ccccccccccc", "Third Song"),
	}}
	lib, _ := testLibrary(t, worker)
	if err := lib.Warm(context.Background()); err != nil {
		t.Fatal(err)
	}
	got, err := lib.Random(context.Background(), 2)
	if err != nil || len(got) != 2 {
		t.Fatalf("random = %d songs err=%v", len(got), err)
	}
	if got[0].VideoID == got[1].VideoID {
		t.Fatalf("duplicate sample: %+v", got)
	}
	worker.mu.Lock()
	upserts := worker.upserts
	worker.mu.Unlock()
	if upserts != 0 {
		t.Fatalf("random sampling wrote to the worker: %d", upserts)
	}
	all, err := lib.Random(context.Background(), 10)
	if err != nil || len(all) != 3 {
		t.Fatalf("oversized limit = %d songs err=%v", len(all), err)
	}
}

func TestLookupAnswersFromCache(t *testing.T) {
	worker := &fakeWorker{songs: []apiSong{sampleSong("aaaaaaaaaaa", "Stored Song")}}
	lib, _ := testLibrary(t, worker)
	if err := lib.Warm(context.Background()); err != nil {
		t.Fatal(err)
	}
	meta, ok := lib.Lookup(context.Background(), "aaaaaaaaaaa")
	if !ok || meta.Title != "Stored Song" || meta.DurationSeconds != 200 {
		t.Fatalf("lookup = %+v ok=%v", meta, ok)
	}
	if _, ok := lib.Lookup(context.Background(), "zzzzzzzzzzz"); ok {
		t.Fatal("unknown song must not be found")
	}
}

func TestRandomOnEmptyCache(t *testing.T) {
	lib, _ := testLibrary(t, &fakeWorker{})
	got, err := lib.Random(context.Background(), 6)
	if err != nil || len(got) != 0 {
		t.Fatalf("empty random = %+v err=%v", got, err)
	}
}
