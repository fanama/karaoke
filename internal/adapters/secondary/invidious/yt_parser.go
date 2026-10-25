package invidious

import (
	"encoding/json"
	"errors"
	"html"
	"regexp"
	"strconv"
	"strings"

	"fana-karaoke/internal/domain"
)

// ytMaxResults mirrors the twenty-hit cap of the Invidious client.
const ytMaxResults = 20

var ytInitialDataRE = regexp.MustCompile(
	`(?s)(?:var\s+ytInitialData|window\["ytInitialData"\])\s*=\s*(\{.*?\});\s*</script>`)

type ytText struct {
	SimpleText string `json:"simpleText"`
	Runs       []struct {
		Text string `json:"text"`
	} `json:"runs"`
}

type ytVideoRenderer struct {
	VideoID     string `json:"videoId"`
	Title       ytText `json:"title"`
	OwnerText   ytText `json:"ownerText"`
	LongByline  ytText `json:"longBylineText"`
	ShortByline ytText `json:"shortBylineText"`
	LengthText  ytText `json:"lengthText"`
	Badges      []struct {
		MetadataBadgeRenderer struct {
			Style string `json:"style"`
		} `json:"metadataBadgeRenderer"`
	} `json:"badges"`
	ThumbnailOverlays []struct {
		ThumbnailOverlayTimeStatusRenderer struct {
			Style string `json:"style"`
			Text  ytText `json:"text"`
		} `json:"thumbnailOverlayTimeStatusRenderer"`
	} `json:"thumbnailOverlays"`
	NavigationEndpoint struct {
		CommandMetadata struct {
			WebCommandMetadata struct {
				URL string `json:"url"`
			} `json:"webCommandMetadata"`
		} `json:"commandMetadata"`
	} `json:"navigationEndpoint"`
}

type ytSection struct {
	ItemSectionRenderer struct {
		Contents []struct {
			VideoRenderer *ytVideoRenderer `json:"videoRenderer"`
		} `json:"contents"`
	} `json:"itemSectionRenderer"`
}

type ytInitialData struct {
	Contents struct {
		TwoColumnSearchResultsRenderer struct {
			PrimaryContents struct {
				SectionListRenderer struct {
					Contents []ytSection `json:"contents"`
				} `json:"sectionListRenderer"`
			} `json:"primaryContents"`
		} `json:"twoColumnSearchResultsRenderer"`
		SectionListRenderer struct {
			Contents []ytSection `json:"contents"`
		} `json:"sectionListRenderer"`
	} `json:"contents"`
}

// sections returns the result sections whatever the page layout is.
func (d ytInitialData) sections() []ytSection {
	two := d.Contents.TwoColumnSearchResultsRenderer.PrimaryContents.SectionListRenderer.Contents
	if len(two) > 0 {
		return two
	}
	return d.Contents.SectionListRenderer.Contents
}

// parseInitialData extracts the search results embedded in YouTube's HTML.
func parseInitialData(page string) ([]domain.SearchResult, error) {
	match := ytInitialDataRE.FindStringSubmatch(page)
	if len(match) < 2 {
		return nil, errors.New("ytInitialData not found")
	}
	var data ytInitialData
	if err := json.Unmarshal([]byte(match[1]), &data); err != nil {
		return nil, err
	}
	return collectResults(data.sections()), nil
}

// collectResults keeps queueable videos, deduplicated and capped.
func collectResults(sections []ytSection) []domain.SearchResult {
	out := make([]domain.SearchResult, 0, ytMaxResults)
	seen := make(map[string]bool, ytMaxResults)
	for _, section := range sections {
		for _, item := range section.ItemSectionRenderer.Contents {
			if item.VideoRenderer == nil {
				continue
			}
			result, ok := rendererToResult(*item.VideoRenderer)
			if !ok || seen[result.VideoID] {
				continue
			}
			seen[result.VideoID] = true
			out = append(out, result)
			if len(out) >= ytMaxResults {
				return out
			}
		}
	}
	return out
}

func rendererToResult(v ytVideoRenderer) (domain.SearchResult, bool) {
	title := ytClean(v.Title)
	if title == "" || !domain.ValidVideoID(v.VideoID) || !queueableYouTube(v) {
		return domain.SearchResult{}, false
	}
	return domain.SearchResult{
		VideoID:         v.VideoID,
		Title:           title,
		Channel:         ytChannel(v),
		ThumbnailURL:    domain.ThumbnailURL(v.VideoID),
		DurationSeconds: ytDuration(v),
	}, true
}

// queueableYouTube applies the room's duration, live and Shorts rules. Live
// streams carry no duration at all, so the range check also drops them.
func queueableYouTube(v ytVideoRenderer) bool {
	if ytIsLive(v) || ytIsShort(v) {
		return false
	}
	seconds := ytDuration(v)
	return seconds > domain.MinVideoSeconds && seconds <= domain.MaxVideoSeconds
}

func ytIsLive(v ytVideoRenderer) bool {
	for _, badge := range v.Badges {
		if badge.MetadataBadgeRenderer.Style == "BADGE_STYLE_TYPE_LIVE_NOW" {
			return true
		}
	}
	for _, overlay := range v.ThumbnailOverlays {
		switch overlay.ThumbnailOverlayTimeStatusRenderer.Style {
		case "LIVE", "SHORTS":
			return true
		}
	}
	return false
}

func ytIsShort(v ytVideoRenderer) bool {
	return strings.Contains(v.NavigationEndpoint.CommandMetadata.WebCommandMetadata.URL, "/shorts/")
}

func ytChannel(v ytVideoRenderer) string {
	for _, candidate := range []ytText{v.OwnerText, v.LongByline, v.ShortByline} {
		if name := ytClean(candidate); name != "" {
			return name
		}
	}
	return ""
}

func ytDuration(v ytVideoRenderer) int {
	if seconds := parseClock(ytClean(v.LengthText)); seconds > 0 {
		return seconds
	}
	for _, overlay := range v.ThumbnailOverlays {
		status := overlay.ThumbnailOverlayTimeStatusRenderer
		if seconds := parseClock(ytClean(status.Text)); seconds > 0 {
			return seconds
		}
	}
	return 0
}

// parseClock turns "3:33" or "1:02:05" into seconds.
func parseClock(raw string) int {
	if raw == "" {
		return 0
	}
	parts := strings.Split(raw, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return 0
	}
	total := 0
	for _, part := range parts {
		value, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || value < 0 {
			return 0
		}
		total = total*60 + value
	}
	return total
}

// text joins the runs YouTube uses instead of a plain string.
func (t ytText) text() string {
	if t.SimpleText != "" {
		return t.SimpleText
	}
	parts := make([]string, 0, len(t.Runs))
	for _, run := range t.Runs {
		parts = append(parts, run.Text)
	}
	return strings.Join(parts, "")
}

func ytClean(t ytText) string {
	return strings.TrimSpace(html.UnescapeString(t.text()))
}
