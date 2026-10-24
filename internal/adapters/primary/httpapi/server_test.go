package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"fana-karaoke/internal/adapters/secondary/clock"
	"fana-karaoke/internal/adapters/secondary/crypto"
	"fana-karaoke/internal/adapters/secondary/memory"
	"fana-karaoke/internal/adapters/secondary/qrgen"
	songdb "fana-karaoke/internal/adapters/secondary/sqlite"
	"fana-karaoke/internal/domain"
	"fana-karaoke/internal/ports"
	"fana-karaoke/internal/usecase"
)

type catalogStub struct{}

func (catalogStub) Search(context.Context, string) ([]domain.SearchResult, bool, error) {
	return []domain.SearchResult{}, false, nil
}

func (catalogStub) Lookup(_ context.Context, id string) (domain.VideoMeta, error) {
	return domain.VideoMeta{
		VideoID:         id,
		Title:           "Test Karaoke",
		Channel:         "Karaoke Channel",
		ThumbnailURL:    "https://i.ytimg.com/vi/" + id + "/mqdefault.jpg",
		DurationSeconds: 200,
	}, nil
}

func newTestServer(t *testing.T) (*httptest.Server, *http.Client, *http.Client) {
	t.Helper()
	return newTestServerWithAdmin(t, "test-admin-token-long")
}

func newTestServerWithAdmin(t *testing.T, adminToken string) (*httptest.Server, *http.Client, *http.Client) {
	t.Helper()
	library, err := songdb.Open(filepath.Join(t.TempDir(), "library.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = library.Close() })
	svc := usecase.NewFacade(usecase.Deps{
		Rooms:   memory.NewRoomRepo(),
		Devices: memory.NewDeviceRepo(),
		Events:  noopEvents{},
		Catalog: catalogStub{},
		Library: library,
		Clock:   clock.NewFake(time.Date(2026, 10, 21, 20, 0, 0, 0, time.UTC)),
		IDs:     crypto.NewIDs(),
		Cfg:     ports.Config{MaxRooms: 10, MaxGuests: 25, MaxQueue: 100, RoomTTL: time.Hour},
	})
	srv := New(svc, NewHub(svc, nil), crypto.NewSigner([]byte("test-secret-test-secret-test-secret")),
		memory.NewRateLimiter(), qrgen.New(), nil, Config{Dev: true, AdminToken: adminToken, Version: "test"},
		slog.New(slog.NewTextHandler(io.Discard, nil)))
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts, newClient(), newClient()
}

type noopEvents struct{}

func (noopEvents) PublishRoomUpdated(string) {}
func (noopEvents) PublishRoomEnded(string)   {}

func newClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{Jar: jar}
}

func callJSON(t *testing.T, client *http.Client, method, url string, body any, out any) (int, string) {
	t.Helper()
	req, err := newRequest(t, method, url, body)
	if err != nil {
		t.Fatal(err)
	}
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)
	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			t.Fatalf("decode %s: %v", string(data), err)
		}
	}
	return res.StatusCode, string(data)
}

func newRequest(t *testing.T, method, url string, body any) (*http.Request, error) {
	t.Helper()
	if body == nil {
		return http.NewRequest(method, url, nil)
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func TestReferrerPolicyAllowsYouTubeEmbed(t *testing.T) {
	ts, _, _ := newTestServer(t)
	res, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	// YouTube rejects embeds whose requests carry no referrer (error 153), so
	// pages must not opt out of sending one. Only the origin is shared
	// cross-origin, which keeps the secret /tv/<token> path private.
	if got := res.Header.Get("Referrer-Policy"); got != "strict-origin-when-cross-origin" {
		t.Fatalf("Referrer-Policy = %q, want strict-origin-when-cross-origin", got)
	}
}
