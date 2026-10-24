package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestLibraryExportImport(t *testing.T) {
	ts, host, guest := newTestServer(t)
	created := createRoom(t, host, ts.URL)
	base := ts.URL + "/api/rooms/" + created.Code
	if status, raw := callJSON(t, guest, http.MethodGet, base+"/library", nil, nil); status != http.StatusForbidden {
		t.Fatalf("guest library access: %d %s", status, raw)
	}
	assertSongAdded(t, guest, ts.URL, created)
	openTV(t, host, created)
	var stats struct {
		Songs int `json:"songs"`
	}
	if status, raw := callJSON(t, host, http.MethodGet, base+"/library", nil, &stats); status != http.StatusOK || stats.Songs != 1 {
		t.Fatalf("library stats: %d %s", status, raw)
	}
	archive := downloadLibrary(t, host, base+"/library/export?format=sql")
	if !bytes.HasPrefix(archive, []byte("SQLite format 3\x00")) {
		t.Fatalf("host export must stay SQLite: %q", archive[:16])
	}
	imported := importLibrary(t, host, base+"/library/import", archive)
	if imported.Imported != 1 || imported.Songs != 1 {
		t.Fatalf("import response: %+v", imported)
	}
}

func openTV(t *testing.T, host *http.Client, created roomCreated) {
	t.Helper()
	res, err := host.Get(created.TVURL)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
}

func downloadLibrary(t *testing.T, client *http.Client, url string) []byte {
	t.Helper()
	res, err := client.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("export: %d %s", res.StatusCode, body)
	}
	if !strings.Contains(res.Header.Get("Content-Disposition"), "attachment") {
		t.Fatalf("export disposition: %q", res.Header.Get("Content-Disposition"))
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func importLibrary(t *testing.T, client *http.Client, url string, archive []byte) struct {
	Imported int `json:"imported"`
	Songs    int `json:"songs"`
} {
	t.Helper()
	res, err := client.Post(url, "application/octet-stream", bytes.NewReader(archive))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var out struct {
		Imported int `json:"imported"`
		Songs    int `json:"songs"`
	}
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("import: %d %s", res.StatusCode, body)
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out
}
