package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type adminSessionResponse struct {
	Enabled       bool `json:"enabled"`
	Authenticated bool `json:"authenticated"`
	Library       bool `json:"library"`
	Songs         int  `json:"songs"`
	Storage       *struct {
		Backend   string `json:"backend"`
		Connected bool   `json:"connected"`
		CacheOnly bool   `json:"cacheOnly"`
		Songs     int    `json:"songs"`
		LastError string `json:"lastError"`
	} `json:"storage"`
}

func readAdminSession(t *testing.T, client *http.Client, base string) adminSessionResponse {
	t.Helper()
	var out adminSessionResponse
	status, raw := callJSON(t, client, http.MethodGet, base+"/api/admin/session", nil, &out)
	if status != http.StatusOK {
		t.Fatalf("admin session: %d %s", status, raw)
	}
	return out
}

func TestAdminLibraryTransfer(t *testing.T) {
	ts, host, guest := newTestServer(t)
	created := createRoom(t, host, ts.URL)
	assertSongAdded(t, guest, ts.URL, created)

	session := readAdminSession(t, guest, ts.URL)
	if !session.Enabled || session.Authenticated || session.Library {
		t.Fatalf("initial admin session: %+v", session)
	}
	if status, _ := callJSON(t, guest, http.MethodGet, ts.URL+"/api/admin/library/export", nil, nil); status != http.StatusUnauthorized {
		t.Fatalf("export without login: %d", status)
	}
	status, raw := callJSON(t, guest, http.MethodPost, ts.URL+"/api/admin/login", map[string]string{"token": "wrong"}, nil)
	if status != http.StatusUnauthorized || !strings.Contains(raw, "invalid_token") {
		t.Fatalf("wrong token: %d %s", status, raw)
	}
	status, raw = callJSON(t, guest, http.MethodPost, ts.URL+"/api/admin/login",
		map[string]string{"token": "test-admin-token-long"}, &session)
	if status != http.StatusOK || !session.Authenticated || !session.Library || session.Songs != 1 {
		t.Fatalf("login: %d %s (%+v)", status, raw, session)
	}
	if session.Storage == nil || session.Storage.Backend != "sqlite" || !session.Storage.Connected {
		t.Fatalf("storage status: %+v", session.Storage)
	}
	archive := downloadLibrary(t, guest, ts.URL+"/api/admin/library/export")
	imported := importLibrary(t, guest, ts.URL+"/api/admin/library/import", archive)
	if imported.Imported != 1 || imported.Songs != 1 {
		t.Fatalf("admin import: %+v", imported)
	}
	status, _ = callJSON(t, guest, http.MethodPost, ts.URL+"/api/admin/logout", map[string]any{}, nil)
	if status != http.StatusNoContent {
		t.Fatalf("logout: %d", status)
	}
	if after := readAdminSession(t, guest, ts.URL); after.Authenticated {
		t.Fatal("logout must clear the admin cookie")
	}
}

func TestAdminSQLTransfer(t *testing.T) {
	ts, host, guest := newTestServer(t)
	created := createRoom(t, host, ts.URL)
	assertSongAdded(t, guest, ts.URL, created)
	callJSON(t, guest, http.MethodPost, ts.URL+"/api/admin/login",
		map[string]string{"token": "test-admin-token-long"}, nil)
	dump := downloadLibrary(t, guest, ts.URL+"/api/admin/library/export?format=sql")
	if !bytes.HasPrefix(dump, []byte("-- fana-karaoke-library v1")) {
		t.Fatalf("SQL export header: %q", dump[:min(len(dump), 40)])
	}
	res, err := guest.Post(ts.URL+"/api/admin/library/import?format=sql", "application/sql", bytes.NewReader(dump))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var out struct {
		Imported int `json:"imported"`
		Songs    int `json:"songs"`
	}
	if res.StatusCode != http.StatusOK || json.NewDecoder(res.Body).Decode(&out) != nil || out.Imported != 1 {
		t.Fatalf("SQL import: %d %+v", res.StatusCode, out)
	}
	if status, _ := callJSON(t, guest, http.MethodGet, ts.URL+"/api/admin/library/export?format=nope", nil, nil); status != http.StatusBadRequest {
		t.Fatalf("invalid format: %d", status)
	}
}

func TestAdminDisabledWithoutToken(t *testing.T) {
	ts, _, _ := newTestServerWithAdmin(t, "")
	session := readAdminSession(t, http.DefaultClient, ts.URL)
	if session.Enabled || session.Authenticated {
		t.Fatalf("disabled admin session: %+v", session)
	}
	status, _ := callJSON(t, http.DefaultClient, http.MethodPost, ts.URL+"/api/admin/login",
		map[string]string{"token": "anything"}, nil)
	if status != http.StatusNotFound {
		t.Fatalf("disabled login: %d", status)
	}
	res, err := http.Get(ts.URL + "/admin")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK || !strings.Contains(string(body), "<!doctype html>") {
		t.Fatalf("admin page: %d", res.StatusCode)
	}
}
