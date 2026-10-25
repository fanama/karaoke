package httpapi

import (
	"net/http"
	"testing"
)

type suggestionBody struct {
	Results []struct {
		VideoID         string `json:"videoId"`
		Title           string `json:"title"`
		ThumbnailURL    string `json:"thumbnailUrl"`
		DurationSeconds int    `json:"durationSeconds"`
	} `json:"results"`
}

func TestRandomSongsEndpoint(t *testing.T) {
	ts, host, guest := newTestServer(t)
	created := createRoom(t, host, ts.URL)
	url := ts.URL + "/api/songs/random?limit=6"
	var empty suggestionBody
	if status, raw := callJSON(t, guest, http.MethodGet, url, nil, &empty); status != http.StatusOK {
		t.Fatalf("empty random: %d %s", status, raw)
	}
	if len(empty.Results) != 0 {
		t.Fatalf("empty library returned songs: %+v", empty.Results)
	}
	assertSongAdded(t, guest, ts.URL, created)
	var body suggestionBody
	status, raw := callJSON(t, guest, http.MethodGet, url, nil, &body)
	if status != http.StatusOK || len(body.Results) != 1 {
		t.Fatalf("random: %d %s", status, raw)
	}
	got := body.Results[0]
	if got.VideoID != "aaaaaaaaaaa" || got.Title == "" || got.ThumbnailURL == "" {
		t.Fatalf("unexpected suggestion: %+v", got)
	}
	if status, _ := callJSON(t, guest, http.MethodGet, ts.URL+"/api/songs/random?limit=99", nil, &body); status != http.StatusOK {
		t.Fatalf("clamped limit: %d", status)
	}
	if status, _ := callJSON(t, guest, http.MethodGet, ts.URL+"/api/songs/random?limit=abc", nil, &body); status != http.StatusOK {
		t.Fatalf("invalid limit: %d", status)
	}
}
