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
	"strings"
	"testing"
	"time"

	"fana-karaoke/internal/app"
	"fana-karaoke/internal/domain"
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
	mgr := app.NewManager(app.Config{MaxRooms: 10, MaxGuests: 25, MaxQueue: 100, RoomTTL: time.Hour}, catalogStub{})
	srv, err := New(mgr, catalogStub{}, nil, Config{
		Dev:     true,
		Secret:  []byte("test-secret-test-secret-test-secret"),
		Version: "test",
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	newClient := func() *http.Client {
		jar, _ := cookiejar.New(nil)
		return &http.Client{Jar: jar}
	}
	return ts, newClient(), newClient()
}

func callJSON(t *testing.T, client *http.Client, method, url string, body any, out any) (int, string) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
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

func TestHostAuthorizationAndRoomFlow(t *testing.T) {
	ts, host, guest := newTestServer(t)

	var created struct {
		Code  string `json:"code"`
		TVURL string `json:"tvUrl"`
		PIN   string `json:"pin"`
	}
	status, raw := callJSON(t, host, http.MethodPost, ts.URL+"/api/rooms", map[string]any{}, &created)
	if status != http.StatusCreated {
		t.Fatalf("create: %d %s", status, raw)
	}
	if created.Code == "" || created.TVURL == "" || len(created.PIN) != 4 {
		t.Fatalf("bad create response: %+v", created)
	}

	var resolved struct {
		Code string `json:"code"`
	}
	status, raw = callJSON(t, guest, http.MethodGet, ts.URL+"/api/rooms/resolve?code="+strings.ReplaceAll(created.Code, "-", ""), nil, &resolved)
	if status != http.StatusOK || resolved.Code != created.Code {
		t.Fatalf("compact code resolve: %d %s", status, raw)
	}

	var snap domain.Snapshot
	status, raw = callJSON(t, guest, http.MethodGet, ts.URL+"/api/rooms/"+created.Code, nil, &snap)
	if status != http.StatusOK {
		t.Fatalf("guest snapshot: %d %s", status, raw)
	}
	if snap.You.HostUnlocked {
		t.Fatal("guest must not be host")
	}

	status, raw = callJSON(t, guest, http.MethodPost, ts.URL+"/api/rooms/"+created.Code+"/playback", map[string]string{"action": "play"}, nil)
	if status != http.StatusForbidden || !strings.Contains(raw, "host_required") {
		t.Fatalf("guest playback should be forbidden: %d %s", status, raw)
	}

	res, err := host.Get(created.TVURL)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("TV page: %d", res.StatusCode)
	}

	var session struct {
		Code string `json:"code"`
		PIN  string `json:"pin"`
	}
	status, raw = callJSON(t, host, http.MethodGet, ts.URL+"/api/tv/session", nil, &session)
	if status != http.StatusOK || session.Code != created.Code || session.PIN != created.PIN {
		t.Fatalf("TV session: %d %s", status, raw)
	}

	var added domain.Snapshot
	status, raw = callJSON(t, guest, http.MethodPost, ts.URL+"/api/rooms/"+created.Code+"/songs",
		map[string]string{"videoId": "aaaaaaaaaaa", "nickname": "Alex"}, &added)
	if status != http.StatusCreated || added.NowPlaying == nil {
		t.Fatalf("add song: %d %s", status, raw)
	}
	if added.NowPlaying.Nickname != "Alex" || added.NowPlaying.Title != "Test Karaoke" {
		t.Fatalf("unexpected now playing entry: %+v", added.NowPlaying)
	}

	status, raw = callJSON(t, host, http.MethodPost, ts.URL+"/api/rooms/"+created.Code+"/playback", map[string]string{"action": "play"}, &snap)
	if status != http.StatusOK || !snap.Room.TVStarted || snap.NowPlaying == nil {
		t.Fatalf("host playback: %d %s", status, raw)
	}

	status, raw = callJSON(t, guest, http.MethodPost, ts.URL+"/api/rooms/"+created.Code+"/host/unlock", map[string]string{"pin": "0000"}, nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("wrong PIN should be 401: %d %s", status, raw)
	}
	status, raw = callJSON(t, guest, http.MethodPost, ts.URL+"/api/rooms/"+created.Code+"/host/unlock", map[string]string{"pin": created.PIN}, &snap)
	if status != http.StatusOK || !snap.You.HostUnlocked {
		t.Fatalf("PIN unlock: %d %s", status, raw)
	}

	res, err = guest.Get(ts.URL + "/api/rooms/" + created.Code + "/qr.svg")
	if err != nil {
		t.Fatal(err)
	}
	qr, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK || !strings.Contains(res.Header.Get("Content-Type"), "image/svg+xml") || !strings.Contains(string(qr), "<svg") {
		t.Fatalf("QR response: %d %s", res.StatusCode, res.Header.Get("Content-Type"))
	}

	status, _ = callJSON(t, guest, http.MethodPost, ts.URL+"/api/rooms/"+created.Code+"/end", map[string]any{}, nil)
	if status != http.StatusNoContent {
		t.Fatalf("end room: %d", status)
	}
	status, _ = callJSON(t, guest, http.MethodGet, ts.URL+"/api/rooms/"+created.Code, nil, nil)
	if status != http.StatusNotFound {
		t.Fatalf("ended room should be 404: %d", status)
	}
}
