package httpapi

import (
	"net/http"
	"testing"
)

type adminSongsResponse struct {
	Songs []struct {
		VideoID         string `json:"videoId"`
		Title           string `json:"title"`
		Channel         string `json:"channel"`
		DurationSeconds int    `json:"durationSeconds"`
	} `json:"songs"`
	Total  int    `json:"total"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
	Query  string `json:"query"`
}

// loginAdmin authenticates the client against the token newTestServer sets.
func loginAdmin(t *testing.T, client *http.Client, base string) {
	t.Helper()
	status, raw := callJSON(t, client, http.MethodPost, base+"/api/admin/login",
		map[string]string{"token": "test-admin-token-long"}, nil)
	if status != http.StatusOK {
		t.Fatalf("admin login: %d %s", status, raw)
	}
}

func TestAdminSongsRequiresLogin(t *testing.T) {
	ts, _, guest := newTestServer(t)
	status, _ := callJSON(t, guest, http.MethodGet, ts.URL+"/api/admin/library/songs", nil, nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("unauthenticated listing: %d", status)
	}
}

func TestAdminSongsPagesAndFilters(t *testing.T) {
	ts, host, guest := newTestServer(t)
	room := createRoom(t, host, ts.URL)
	assertSongAdded(t, guest, ts.URL, room)
	loginAdmin(t, guest, ts.URL)

	var page adminSongsResponse
	status, raw := callJSON(t, guest, http.MethodGet,
		ts.URL+"/api/admin/library/songs?offset=0&limit=10", nil, &page)
	if status != http.StatusOK {
		t.Fatalf("listing: %d %s", status, raw)
	}
	if page.Total != 1 || len(page.Songs) != 1 || page.Limit != 10 {
		t.Fatalf("page = %+v", page)
	}
	if page.Songs[0].VideoID == "" || page.Songs[0].Title == "" {
		t.Fatalf("song row = %+v", page.Songs[0])
	}

	missed := adminSongsResponse{}
	status, _ = callJSON(t, guest, http.MethodGet,
		ts.URL+"/api/admin/library/songs?q=definitely-not-a-title", nil, &missed)
	if status != http.StatusOK || missed.Total != 0 || len(missed.Songs) != 0 {
		t.Fatalf("filtered listing: %d %+v", status, missed)
	}

	hit := adminSongsResponse{}
	status, _ = callJSON(t, guest, http.MethodGet,
		ts.URL+"/api/admin/library/songs?q=karaoke&limit=5&offset=0", nil, &hit)
	if status != http.StatusOK || hit.Total != 1 || hit.Query != "karaoke" {
		t.Fatalf("matching filter: %d %+v", status, hit)
	}
}

func TestAdminSongsRejectsBadPagination(t *testing.T) {
	ts, _, guest := newTestServer(t)
	loginAdmin(t, guest, ts.URL)
	for _, query := range []string{"limit=0", "limit=101", "limit=abc", "offset=-1", "offset=1000001"} {
		status, _ := callJSON(t, guest, http.MethodGet,
			ts.URL+"/api/admin/library/songs?"+query, nil, nil)
		if status != http.StatusBadRequest {
			t.Fatalf("%s: %d", query, status)
		}
	}
}
