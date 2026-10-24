package d1

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/ports"
)

type fakeWorker struct {
	mu          sync.Mutex
	songs       []apiSong
	fail        bool
	upserts     int
	importCalls int
}

func (f *fakeWorker) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		switch {
		case r.URL.Path == "/v1/health":
			f.health(w)
		case r.URL.Path == "/v1/songs" && r.Method == http.MethodGet:
			f.list(w, r)
		case r.URL.Path == "/v1/songs/upsert":
			f.upserts++
			f.write(w)
		case r.URL.Path == "/v1/songs/import":
			f.importCalls++
			f.write(w)
		default:
			http.NotFound(w, r)
		}
	})
}

func (f *fakeWorker) health(w http.ResponseWriter) {
	if f.fail {
		writeFakeError(w, http.StatusServiceUnavailable, "d1_unavailable")
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (f *fakeWorker) list(w http.ResponseWriter, r *http.Request) {
	if f.fail {
		writeFakeError(w, http.StatusServiceUnavailable, "d1_unavailable")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 500
	}
	end := offset + limit
	if end > len(f.songs) {
		end = len(f.songs)
	}
	page := map[string]any{"songs": f.songs[offset:end]}
	if end < len(f.songs) {
		page["nextOffset"] = end
	}
	_ = json.NewEncoder(w).Encode(page)
}

func (f *fakeWorker) write(w http.ResponseWriter) {
	if f.fail {
		writeFakeError(w, http.StatusServiceUnavailable, "d1_unavailable")
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]int{"written": 1})
}

func writeFakeError(w http.ResponseWriter, status int, code string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": code, "message": "unavailable"},
	})
}

func testLibrary(t *testing.T, worker *fakeWorker) (*Library, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(worker.handler())
	t.Cleanup(server.Close)
	lib := New(Config{
		BaseURL: server.URL, Token: "test-token",
		WarmupTimeout: time.Second, WriteTimeout: time.Second,
		CacheWriteTimeout: time.Second, Now: func() time.Time {
			return time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)
		},
	})
	t.Cleanup(func() { _ = lib.Close() })
	return lib, server
}

func sampleSong(id, title string) apiSong {
	return apiSong{
		VideoID: id, Title: title, Channel: "Studio", DurationSeconds: 200,
		SearchText:  domain.SearchText(title, "Studio"),
		FirstSeenAt: "2026-10-24T10:00:00Z", LastSeenAt: "2026-10-24T10:00:00Z",
	}
}

func TestWarmLoadsPagesAndSearches(t *testing.T) {
	worker := &fakeWorker{songs: []apiSong{
		sampleSong("aaaaaaaaaaa", "Été – Toto (karaoké)"),
		sampleSong("bbbbbbbbbbb", "Bohemian Rhapsody"),
	}}
	lib, _ := testLibrary(t, worker)
	worker.songs = append(worker.songs, sampleSong("ccccccccccc", "Third"))
	if err := lib.Warm(context.Background()); err != nil {
		t.Fatal(err)
	}
	if total, _ := lib.Count(context.Background()); total != 3 {
		t.Fatalf("count = %d, want 3", total)
	}
	found, _ := lib.Search(context.Background(), "ete", 20)
	if len(found) != 1 || found[0].VideoID != "aaaaaaaaaaa" {
		t.Fatalf("accent search = %+v", found)
	}
}

func TestSaveWaitsAndRejectsFailure(t *testing.T) {
	worker := &fakeWorker{fail: true}
	lib, _ := testLibrary(t, worker)
	err := lib.Save(context.Background(), []domain.SearchResult{{VideoID: "aaaaaaaaaaa", Title: "Song"}})
	if !errors.Is(err, domain.ErrLibraryUnavailable) {
		t.Fatalf("save error = %v", err)
	}
	if total, _ := lib.Count(context.Background()); total != 0 {
		t.Fatalf("failed save changed cache: %d", total)
	}
	worker.fail = false
	if err := lib.Save(context.Background(), []domain.SearchResult{{VideoID: "aaaaaaaaaaa", Title: "Song"}}); err != nil {
		t.Fatal(err)
	}
	if total, _ := lib.Count(context.Background()); total != 1 {
		t.Fatalf("successful save count = %d", total)
	}
}

func TestCacheIsBestEffort(t *testing.T) {
	worker := &fakeWorker{fail: true}
	lib, _ := testLibrary(t, worker)
	song := domain.SearchResult{VideoID: "aaaaaaaaaaa", Title: "Cached Song"}
	if err := lib.Cache(context.Background(), []domain.SearchResult{song}); err != nil {
		t.Fatal(err)
	}
	found, _ := lib.Search(context.Background(), "cached", 20)
	if len(found) != 1 {
		t.Fatalf("cached search = %+v", found)
	}
	time.Sleep(20 * time.Millisecond)
	if lib.Status(context.Background()).LastError == "" {
		t.Fatal("background failure did not set a status error")
	}
}

func TestExportAndImportBothFormats(t *testing.T) {
	worker := &fakeWorker{songs: []apiSong{sampleSong("aaaaaaaaaaa", "First Song")}}
	lib, _ := testLibrary(t, worker)
	if err := lib.Warm(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, format := range []ports.LibraryFormat{ports.FormatSQL, ports.FormatSQLite} {
		var buf bytes.Buffer
		if err := lib.Export(context.Background(), format, &buf); err != nil {
			t.Fatalf("export %s: %v", format, err)
		}
		target, _ := testLibrary(t, worker)
		imported, err := target.Import(context.Background(), format, bytes.NewReader(buf.Bytes()))
		if err != nil || imported != 1 {
			t.Fatalf("import %s: %d %v", format, imported, err)
		}
	}
}
