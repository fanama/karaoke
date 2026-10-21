package invidious

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"fana-karaoke/internal/domain"
)

func newTestCatalog(t *testing.T, serverURL string, now *time.Time) *Catalog {
	t.Helper()
	upstream := NewClient(Options{Instances: []string{serverURL}})
	prober := NewProber(&http.Client{Timeout: time.Second}, serverURL+"/oembed")
	cache := NewCache(time.Second, time.Hour, 7*24*time.Hour, func() time.Time { return *now })
	return NewCatalog(upstream, prober, cache)
}

func searchServer(t *testing.T, hits *int32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oembed":
			probeOEmbed(t, w, r)
		case "/api/v1/search":
			atomic.AddInt32(hits, 1)
			if !strings.Contains(r.URL.Query().Get("q"), "karaoke") {
				t.Errorf("query was not normalized: %s", r.URL.Query().Get("q"))
			}
			_ = json.NewEncoder(w).Encode(searchPayload())
		default:
			http.NotFound(w, r)
		}
	}))
}

func probeOEmbed(t *testing.T, w http.ResponseWriter, r *http.Request) {
	t.Helper()
	if strings.Contains(r.URL.Query().Get("url"), "fffffffffff") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"title": "OEmbed karaoke", "author_name": "OEmbed channel",
		"thumbnail_url": "https://i.ytimg.com/vi/aaaaaaaaaaa/mqdefault.jpg",
	})
}

func searchPayload() []map[string]any {
	return []map[string]any{
		{"type": "video", "title": "Good Karaoke", "videoId": "aaaaaaaaaaa", "author": "A", "lengthSeconds": 300},
		{"type": "video", "title": "Embed blocked", "videoId": "fffffffffff", "author": "F", "lengthSeconds": 300},
		{"type": "video", "title": "Live", "videoId": "bbbbbbbbbbb", "author": "B", "lengthSeconds": 300, "liveNow": true},
		{"type": "video", "title": "Short", "videoId": "ccccccccccc", "author": "C", "lengthSeconds": 40},
		{"type": "video", "title": "Too long", "videoId": "ddddddddddd", "author": "D", "lengthSeconds": 720},
		{"type": "video", "title": "Duplicate", "videoId": "aaaaaaaaaaa", "author": "A", "lengthSeconds": 300},
		{"type": "channel", "title": "Not a video", "videoId": "eeeeeeeeeee"},
	}
}

func TestSearchFiltersAndCaches(t *testing.T) {
	var hits int32
	server := searchServer(t, &hits)
	defer server.Close()
	now := time.Date(2026, 10, 21, 20, 0, 0, 0, time.UTC)
	catalog := newTestCatalog(t, server.URL, &now)

	results, stale, err := catalog.Search(context.Background(), "  bohemian   rhapsody ")
	if err != nil || stale || len(results) != 1 || results[0].VideoID != "aaaaaaaaaaa" {
		t.Fatalf("search: results=%+v stale=%v err=%v", results, stale, err)
	}
	meta, err := catalog.Lookup(context.Background(), "aaaaaaaaaaa")
	if err != nil || meta.Title != "Good Karaoke" || meta.DurationSeconds != 300 {
		t.Fatalf("search metadata should be reusable: %+v err=%v", meta, err)
	}
	if _, _, err := catalog.Search(context.Background(), "bohemian rhapsody"); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&hits); got != 1 {
		t.Fatalf("second search should hit cache, upstream hits=%d", got)
	}
	now = now.Add(2 * time.Second)
	server.Close()
	results, stale, err = catalog.Search(context.Background(), "bohemian rhapsody")
	if err != nil || !stale || len(results) != 1 {
		t.Fatalf("stale fallback: results=%d stale=%v err=%v", len(results), stale, err)
	}
}

func TestLookupMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/videos/aaaaaaaaaaa":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"type": "video", "title": "Test Song", "author": "Channel", "lengthSeconds": 210,
			})
		case "/oembed":
			_ = json.NewEncoder(w).Encode(map[string]any{"title": "Test Song", "author_name": "Channel"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	now := time.Now()
	meta, err := newTestCatalog(t, server.URL, &now).Lookup(context.Background(), "aaaaaaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	if meta.Title != "Test Song" || meta.Channel != "Channel" || meta.DurationSeconds != 210 {
		t.Fatalf("unexpected metadata: %+v", meta)
	}
}

func TestLookupRejectsEmbedBlockedVideo(t *testing.T) {
	var oembedHits int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/videos/aaaaaaaaaaa":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"type": "video", "title": "Blocked Song", "author": "Channel", "lengthSeconds": 210,
			})
		case "/oembed":
			atomic.AddInt32(&oembedHits, 1)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	now := time.Now()
	catalog := newTestCatalog(t, server.URL, &now)
	if _, err := catalog.Lookup(context.Background(), "aaaaaaaaaaa"); !errors.Is(err, domain.ErrEmbedBlocked) {
		t.Fatalf("expected embed blocked, got %v", err)
	}
	if _, err := catalog.Lookup(context.Background(), "aaaaaaaaaaa"); !errors.Is(err, domain.ErrEmbedBlocked) {
		t.Fatalf("expected cached embed blocked, got %v", err)
	}
	if hits := atomic.LoadInt32(&oembedHits); hits != 1 {
		t.Fatalf("blocked lookup should be cached, oembed hits=%d", hits)
	}
}

func TestLookupFallsBackToOEmbed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/videos/aaaaaaaaaaa":
			http.Error(w, "down", http.StatusInternalServerError)
		case "/oembed":
			_ = json.NewEncoder(w).Encode(map[string]any{"title": "OEmbed Song", "author_name": "OEmbed Channel"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	now := time.Now()
	meta, err := newTestCatalog(t, server.URL, &now).Lookup(context.Background(), "aaaaaaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	if meta.Title != "OEmbed Song" || meta.Channel != "OEmbed Channel" {
		t.Fatalf("unexpected oembed metadata: %+v", meta)
	}
}

func TestNormalizeQuery(t *testing.T) {
	cases := map[string]string{
		"  Hello   World ": "Hello World karaoke",
		"Karaoke Queen":    "Karaoke Queen",
		"bohemian karaoke": "bohemian karaoke",
		"\x01tune\x02":     "tune karaoke",
		"":                 "",
	}
	for input, want := range cases {
		if got := normalizeQuery(input); got != want {
			t.Fatalf("%q: got %q want %q", input, got, want)
		}
	}
}
