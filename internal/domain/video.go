package domain

import (
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

var idPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

var allowedHosts = map[string]bool{
	"youtube.com":       true,
	"www.youtube.com":   true,
	"m.youtube.com":     true,
	"music.youtube.com": true,
	"youtu.be":          true,
}

// MaxVideoSeconds caps a queueable song at ten minutes.
const MaxVideoSeconds = 600

// MinVideoSeconds rejects Shorts, which cannot be sung to.
const MinVideoSeconds = 60

// ParseYouTubeURL extracts a video id from a pasted link. It never fetches the
// URL, so arbitrary hosts cannot be used for SSRF.
func ParseYouTubeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 2048 {
		return "", ErrInvalidURL
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || !allowedHosts[strings.ToLower(u.Hostname())] {
		return "", ErrInvalidURL
	}
	if strings.ToLower(u.Hostname()) == "youtu.be" {
		return shortLinkID(u)
	}
	if id := u.Query().Get("v"); idPattern.MatchString(id) {
		return id, nil
	}
	return pathID(u)
}

func shortLinkID(u *url.URL) (string, error) {
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) > 0 && idPattern.MatchString(strings.TrimSpace(parts[0])) {
		return strings.TrimSpace(parts[0]), nil
	}
	return "", ErrInvalidURL
}

func pathID(u *url.URL) (string, error) {
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) >= 2 && (parts[0] == "shorts" || parts[0] == "embed" || parts[0] == "live") {
		if idPattern.MatchString(parts[1]) {
			return parts[1], nil
		}
	}
	return "", ErrInvalidURL
}

// ValidVideoID reports whether an id looks like a YouTube id.
func ValidVideoID(id string) bool { return idPattern.MatchString(id) }

// IsShortLink reports whether a pasted URL points at a Shorts page.
func IsShortLink(raw string) bool { return strings.Contains(strings.ToLower(raw), "/shorts/") }

// CleanNickname trims and validates a singer name.
func CleanNickname(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if n := utf8.RuneCountInString(name); n < 1 || n > 24 {
		return "", ErrInvalidInput
	}
	return name, nil
}

// ValidateVideoMeta enforces the queueable-video rules.
func ValidateVideoMeta(meta VideoMeta) error {
	if meta.Live {
		return ErrVideoLive
	}
	if meta.DurationSeconds > MaxVideoSeconds {
		return ErrVideoTooLong
	}
	if meta.DurationSeconds > 0 && meta.DurationSeconds <= MinVideoSeconds {
		return ErrVideoShort
	}
	if meta.Title == "" {
		return ErrVideoMissing
	}
	return nil
}

// ThumbnailURL is the default YouTube thumbnail for a video id.
func ThumbnailURL(videoID string) string {
	return "https://i.ytimg.com/vi/" + videoID + "/mqdefault.jpg"
}
