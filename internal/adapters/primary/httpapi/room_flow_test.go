package httpapi

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"fana-karaoke/internal/domain"
)

func TestHostAuthorizationAndRoomFlow(t *testing.T) {
	ts, host, guest := newTestServer(t)
	created := createRoom(t, host, ts.URL)
	resolveCompactCode(t, guest, ts.URL, created)
	snap := guestSnapshot(t, guest, ts.URL, created)
	verifyGuestIsNotHost(t, guest, ts.URL, created, snap)
	verifyTVSession(t, host, ts.URL, created)
	assertSongAdded(t, guest, ts.URL, created)
	verifyHostPlayback(t, host, ts.URL, created)
	verifyPINUnlock(t, guest, ts.URL, created)
	verifyQR(t, guest, ts.URL, created)
	verifyEndRoom(t, guest, ts.URL, created)
}

type roomCreated struct {
	Code  string `json:"code"`
	TVURL string `json:"tvUrl"`
	PIN   string `json:"pin"`
}

func createRoom(t *testing.T, host *http.Client, base string) roomCreated {
	t.Helper()
	var created roomCreated
	status, raw := callJSON(t, host, http.MethodPost, base+"/api/rooms", map[string]any{}, &created)
	if status != http.StatusCreated {
		t.Fatalf("create: %d %s", status, raw)
	}
	if created.Code == "" || created.TVURL == "" || len(created.PIN) != 4 {
		t.Fatalf("bad create response: %+v", created)
	}
	return created
}

func resolveCompactCode(t *testing.T, guest *http.Client, base string, created roomCreated) {
	t.Helper()
	var resolved struct {
		Code string `json:"code"`
	}
	url := base + "/api/rooms/resolve?code=" + strings.ReplaceAll(created.Code, "-", "")
	status, raw := callJSON(t, guest, http.MethodGet, url, nil, &resolved)
	if status != http.StatusOK || resolved.Code != created.Code {
		t.Fatalf("compact code resolve: %d %s", status, raw)
	}
}

func guestSnapshot(t *testing.T, guest *http.Client, base string, created roomCreated) domain.Snapshot {
	t.Helper()
	var snap domain.Snapshot
	status, raw := callJSON(t, guest, http.MethodGet, base+"/api/rooms/"+created.Code, nil, &snap)
	if status != http.StatusOK {
		t.Fatalf("guest snapshot: %d %s", status, raw)
	}
	return snap
}

func verifyGuestIsNotHost(t *testing.T, guest *http.Client, base string, created roomCreated, snap domain.Snapshot) {
	t.Helper()
	if snap.You.HostUnlocked {
		t.Fatal("guest must not be host")
	}
	url := base + "/api/rooms/" + created.Code + "/playback"
	status, raw := callJSON(t, guest, http.MethodPost, url, map[string]string{"action": "play"}, nil)
	if status != http.StatusForbidden || !strings.Contains(raw, "host_required") {
		t.Fatalf("guest playback should be forbidden: %d %s", status, raw)
	}
}

func verifyTVSession(t *testing.T, host *http.Client, base string, created roomCreated) {
	t.Helper()
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
	status, raw := callJSON(t, host, http.MethodGet, base+"/api/tv/session", nil, &session)
	if status != http.StatusOK || session.Code != created.Code || session.PIN != created.PIN {
		t.Fatalf("TV session: %d %s", status, raw)
	}
}

func assertSongAdded(t *testing.T, guest *http.Client, base string, created roomCreated) {
	t.Helper()
	var added domain.Snapshot
	body := map[string]string{"videoId": "aaaaaaaaaaa", "nickname": "Alex"}
	status, raw := callJSON(t, guest, http.MethodPost, base+"/api/rooms/"+created.Code+"/songs", body, &added)
	if status != http.StatusCreated || added.NowPlaying == nil {
		t.Fatalf("add song: %d %s", status, raw)
	}
	if added.NowPlaying.Nickname != "Alex" || added.NowPlaying.Title != "Test Karaoke" {
		t.Fatalf("unexpected now playing entry: %+v", added.NowPlaying)
	}
}

func verifyHostPlayback(t *testing.T, host *http.Client, base string, created roomCreated) {
	t.Helper()
	var snap domain.Snapshot
	url := base + "/api/rooms/" + created.Code + "/playback"
	status, raw := callJSON(t, host, http.MethodPost, url, map[string]string{"action": "play"}, &snap)
	if status != http.StatusOK || !snap.Room.TVStarted || snap.NowPlaying == nil {
		t.Fatalf("host playback: %d %s", status, raw)
	}
}

func verifyPINUnlock(t *testing.T, guest *http.Client, base string, created roomCreated) {
	t.Helper()
	url := base + "/api/rooms/" + created.Code + "/host/unlock"
	status, raw := callJSON(t, guest, http.MethodPost, url, map[string]string{"pin": "0000"}, nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("wrong PIN should be 401: %d %s", status, raw)
	}
	var snap domain.Snapshot
	status, raw = callJSON(t, guest, http.MethodPost, url, map[string]string{"pin": created.PIN}, &snap)
	if status != http.StatusOK || !snap.You.HostUnlocked {
		t.Fatalf("PIN unlock: %d %s", status, raw)
	}
}

func verifyQR(t *testing.T, guest *http.Client, base string, created roomCreated) {
	t.Helper()
	res, err := guest.Get(base + "/api/rooms/" + created.Code + "/qr.svg")
	if err != nil {
		t.Fatal(err)
	}
	qr, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK || !strings.Contains(res.Header.Get("Content-Type"), "image/svg+xml") || !strings.Contains(string(qr), "<svg") {
		t.Fatalf("QR response: %d %s", res.StatusCode, res.Header.Get("Content-Type"))
	}
}

func verifyEndRoom(t *testing.T, guest *http.Client, base string, created roomCreated) {
	t.Helper()
	status, _ := callJSON(t, guest, http.MethodPost, base+"/api/rooms/"+created.Code+"/end", map[string]any{}, nil)
	if status != http.StatusNoContent {
		t.Fatalf("end room: %d", status)
	}
	status, _ = callJSON(t, guest, http.MethodGet, base+"/api/rooms/"+created.Code, nil, nil)
	if status != http.StatusNotFound {
		t.Fatalf("ended room should be 404: %d", status)
	}
}
