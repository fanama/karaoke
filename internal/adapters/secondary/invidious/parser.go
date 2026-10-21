package invidious

import (
	"html"
	"strings"

	"fana-karaoke/internal/domain"
)

type searchItem struct {
	Type            string `json:"type"`
	Title           string `json:"title"`
	VideoID         string `json:"videoId"`
	Author          string `json:"author"`
	LengthSeconds   int    `json:"lengthSeconds"`
	LiveNow         bool   `json:"liveNow"`
	VideoThumbnails []struct {
		URL string `json:"url"`
	} `json:"videoThumbnails"`
}

type videoItem struct {
	Type          string `json:"type"`
	Title         string `json:"title"`
	Author        string `json:"author"`
	LengthSeconds int    `json:"lengthSeconds"`
	LiveNow       bool   `json:"liveNow"`
}

func parseSearchItems(items []searchItem) []domain.SearchResult {
	out := make([]domain.SearchResult, 0, 20)
	seen := make(map[string]bool, len(items))
	for _, item := range items {
		if !queueable(item) || seen[item.VideoID] {
			continue
		}
		seen[item.VideoID] = true
		out = append(out, domain.SearchResult{
			VideoID:         item.VideoID,
			Title:           html.UnescapeString(item.Title),
			Channel:         html.UnescapeString(item.Author),
			ThumbnailURL:    thumbnail(item.VideoID),
			DurationSeconds: item.LengthSeconds,
		})
		if len(out) >= 20 {
			break
		}
	}
	return out
}

func queueable(item searchItem) bool {
	if item.Type != "video" || item.LiveNow {
		return false
	}
	if item.LengthSeconds <= domain.MinVideoSeconds || item.LengthSeconds > domain.MaxVideoSeconds {
		return false
	}
	return domain.ValidVideoID(item.VideoID)
}

func parseVideoItem(videoID string, item videoItem) (domain.VideoMeta, bool) {
	if item.Title == "" {
		return domain.VideoMeta{}, false
	}
	return domain.VideoMeta{
		VideoID:         videoID,
		Title:           html.UnescapeString(item.Title),
		Channel:         html.UnescapeString(item.Author),
		ThumbnailURL:    thumbnail(videoID),
		DurationSeconds: item.LengthSeconds,
		Live:            item.LiveNow,
	}, true
}

// normalizeQuery trims control characters, caps the length, and biases results
// towards karaoke versions.
func normalizeQuery(raw string) string {
	fields := strings.Fields(raw)
	cleaned := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.Map(func(r rune) rune {
			if r < 32 || r == 127 {
				return -1
			}
			return r
		}, f)
		if f != "" {
			cleaned = append(cleaned, f)
		}
	}
	q := strings.Join(cleaned, " ")
	if runes := []rune(q); len(runes) > 120 {
		q = string(runes[:120])
	}
	if q == "" {
		return ""
	}
	if !strings.Contains(strings.ToLower(q), "karaoke") {
		q += " karaoke"
	}
	return q
}

func thumbnail(videoID string) string { return domain.ThumbnailURL(videoID) }
