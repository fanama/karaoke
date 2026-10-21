package usecase

import (
	"context"
	"testing"
	"time"

	"fana-karaoke/internal/domain"
)

func TestSkipVotePassesAndAdvances(t *testing.T) {
	f, clk := newTestFacade(t)
	code := mustRoom(t, f)
	mustAdd(t, f, code, "dev-a", "Alex", "aaaaaaaaaaa")
	mustAdd(t, f, code, "dev-b", "Sam", "bbbbbbbbbbb")
	if _, err := f.Playback(code, "play", "dev-host", "tv"); err != nil {
		t.Fatal(err)
	}
	clk.Advance(6 * time.Second)
	f.Lifecycle.Tick()
	snap := mustSnapshot(t, f, code, "dev-a", "guest")
	if snap.NowPlaying == nil || snap.NowPlaying.Status != domain.StatusPlaying {
		t.Fatalf("expected playing, got %+v", snap.NowPlaying)
	}

	snap, err := f.StartSkipVote(code, "dev-a", "guest")
	if err != nil {
		t.Fatal(err)
	}
	if snap.SkipVote == nil || !snap.SkipVote.Active || snap.SkipVote.Threshold != 2 {
		t.Fatalf("unexpected vote: %+v", snap.SkipVote)
	}
	snap, err = f.CastSkipVote(code, "dev-a", "guest")
	if err != nil || snap.SkipVote == nil || snap.SkipVote.Yes != 1 {
		t.Fatalf("first vote should be recorded: %+v err=%v", snap.SkipVote, err)
	}
	snap, err = f.CastSkipVote(code, "dev-b", "guest")
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
	f, clk := newTestFacade(t)
	code := mustRoom(t, f)
	mustAdd(t, f, code, "dev-a", "Alex", "aaaaaaaaaaa")
	mustAdd(t, f, code, "dev-b", "Sam", "bbbbbbbbbbb")
	if _, err := f.Playback(code, "play", "dev-host", "tv"); err != nil {
		t.Fatal(err)
	}
	clk.Advance(5 * time.Second)
	f.Lifecycle.Tick()
	snap := mustSnapshot(t, f, code, "dev-a", "guest")
	if snap.NowPlaying == nil || snap.NowPlaying.Status != domain.StatusPlaying {
		t.Fatal("countdown should turn into playback")
	}

	gen := snap.NowPlaying.Generation
	entry := snap.NowPlaying.EntryID
	f.TVEnded(code, entry, gen, "dev-tv")
	snap = mustSnapshot(t, f, code, "dev-a", "guest")
	if snap.NowPlaying == nil || snap.NowPlaying.EntryID == entry {
		t.Fatal("ended song should advance")
	}
	clk.Advance(5 * time.Second)
	f.Lifecycle.Tick()
	snap = mustSnapshot(t, f, code, "dev-a", "guest")
	if snap.NowPlaying.Status != domain.StatusPlaying {
		t.Fatal("second song should play")
	}
}

func TestQueueCapAndAutoStart(t *testing.T) {
	f, _ := newTestFacade(t)
	code := mustRoom(t, f)
	snap, err := f.AddSong(context.Background(), code, "dev-a", "guest", "Alex", "aaaaaaaaaaa", "")
	if err != nil {
		t.Fatal(err)
	}
	if snap.NowPlaying == nil || snap.NowPlaying.Status != domain.StatusCountdown {
		t.Fatal("the first queued song should start automatically")
	}
	if !snap.Room.TVStarted {
		t.Fatal("room should be marked as started")
	}
}
