package invidious

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func ytVideo(id, title, channel, length string) map[string]any {
	video := map[string]any{
		"videoId":    id,
		"title":      map[string]any{"runs": []any{map[string]any{"text": title}}},
		"ownerText":  map[string]any{"runs": []any{map[string]any{"text": channel}}},
		"lengthText": map[string]any{"simpleText": length},
	}
	return video
}

func ytPage(t *testing.T, videos ...map[string]any) string {
	t.Helper()
	renderers := make([]any, 0, len(videos))
	for _, video := range videos {
		renderers = append(renderers, map[string]any{"videoRenderer": video})
	}
	data := map[string]any{
		"contents": map[string]any{
			"twoColumnSearchResultsRenderer": map[string]any{
				"primaryContents": map[string]any{
					"sectionListRenderer": map[string]any{
						"contents": []any{map[string]any{
							"itemSectionRenderer": map[string]any{"contents": renderers},
						}},
					},
				},
			},
		},
	}
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	return "<html><script>var ytInitialData = " + string(raw) + ";</script></html>"
}

func liveVideo(id, title, length string) map[string]any {
	video := ytVideo(id, title, "C", length)
	video["badges"] = []any{map[string]any{
		"metadataBadgeRenderer": map[string]any{"style": "BADGE_STYLE_TYPE_LIVE_NOW"},
	}}
	return video
}

func shortsVideo(id, title, length string) map[string]any {
	video := ytVideo(id, title, "E", length)
	video["navigationEndpoint"] = map[string]any{
		"commandMetadata": map[string]any{
			"webCommandMetadata": map[string]any{"url": "/shorts/" + id},
		},
	}
	return video
}

func TestParseInitialDataFiltersAndMaps(t *testing.T) {
	page := ytPage(t,
		ytVideo("aaaaaaaaaaa", "Good Karaoke", "Sing King", "3:33"),
		liveVideo("ccccccccccc", "Live Karaoke", "1:00"),
		ytVideo("bbbbbbbbbbb", "No Length", "B", ""),
		shortsVideo("eeeeeeeeeee", "Short Karaoke", "3:00"),
		ytVideo("fffffffffff", "Too Short", "F", "0:45"),
		ytVideo("ggggggggggg", "Too Long", "G", "10:01"),
		ytVideo("hhhhhhhhhhh", "Exactly A Minute", "H", "1:00"),
		ytVideo("abc", "Bad Id", "I", "3:00"),
		ytVideo("bbbbbbbbbbb", "Ten Minutes", "B", "10:00"),
		ytVideo("aaaaaaaaaaa", "Duplicate Karaoke", "Sing King", "3:33"),
	)

	results, err := parseInitialData(page)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("kept %d results: %+v", len(results), results)
	}
	if results[0].VideoID != "aaaaaaaaaaa" || results[0].DurationSeconds != 213 {
		t.Fatalf("first result: %+v", results[0])
	}
	if results[0].Channel != "Sing King" || !strings.HasPrefix(results[0].ThumbnailURL, "https://i.ytimg.com/") {
		t.Fatalf("first result metadata: %+v", results[0])
	}
	if results[1].VideoID != "bbbbbbbbbbb" || results[1].DurationSeconds != 600 {
		t.Fatalf("second result: %+v", results[1])
	}
}

func TestParseInitialDataAcceptsWindowAssignment(t *testing.T) {
	page := strings.Replace(ytPage(t, ytVideo("aaaaaaaaaaa", "Title", "C", "2:00")),
		"var ytInitialData", `window["ytInitialData"]`, 1)
	results, err := parseInitialData(page)
	if err != nil || len(results) != 1 {
		t.Fatalf("results = %+v err=%v", results, err)
	}
}

func TestParseInitialDataRejectsMissingPayload(t *testing.T) {
	if _, err := parseInitialData("<html>consent</html>"); err == nil {
		t.Fatal("expected a missing ytInitialData error")
	}
}

func TestParseClock(t *testing.T) {
	cases := map[string]int{"3:33": 213, "1:02:05": 3725, "0:45": 45, "10:01": 601, "": 0, "abc": 0}
	for raw, want := range cases {
		if got := parseClock(raw); got != want {
			t.Fatalf("%q: got %d want %d", raw, got, want)
		}
	}
}

func TestSearchYouTubeSendsConsentAndQuery(t *testing.T) {
	var seen *http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r
		_, _ = w.Write([]byte(ytPage(t, ytVideo("aaaaaaaaaaa", "Scraped Karaoke", "C", "3:00"))))
	}))
	defer server.Close()

	client := NewClient(Options{YouTubeURL: server.URL, Region: "FR"})
	results, err := client.Search(context.Background(), "hello karaoke")
	if err != nil || len(results) != 1 || results[0].Title != "Scraped Karaoke" {
		t.Fatalf("results = %+v err=%v", results, err)
	}
	if seen.URL.Query().Get("search_query") != "hello karaoke" || seen.URL.Query().Get("gl") != "FR" {
		t.Fatalf("unexpected query: %s", seen.URL.RawQuery)
	}
	if !strings.Contains(seen.Header.Get("Cookie"), "CONSENT=") {
		t.Fatalf("missing consent cookie: %q", seen.Header.Get("Cookie"))
	}
	if !strings.Contains(seen.Header.Get("User-Agent"), "Mozilla/5.0") {
		t.Fatalf("unexpected user agent: %q", seen.Header.Get("User-Agent"))
	}
}

func TestSearchFallsBackToInvidious(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/yt/results":
			http.Error(w, "blocked", http.StatusTooManyRequests)
		case "/api/v1/search":
			_ = json.NewEncoder(w).Encode([]map[string]any{{
				"type": "video", "title": "Fallback Karaoke", "videoId": "bbbbbbbbbbb",
				"author": "B", "lengthSeconds": 200,
			}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient(Options{Instances: []string{server.URL}, YouTubeURL: server.URL + "/yt/results"})
	results, err := client.Search(context.Background(), "hello karaoke")
	if err != nil || len(results) != 1 || results[0].Title != "Fallback Karaoke" {
		t.Fatalf("results = %+v err=%v", results, err)
	}
}
