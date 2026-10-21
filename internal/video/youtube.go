package video

import (
	"net/url"
	"regexp"
	"strings"

	"fana-karaoke/internal/domain"
)

var idPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

var allowedHosts = map[string]bool{
	"youtube.com":       true,
	"www.youtube.com":   true,
	"m.youtube.com":     true,
	"music.youtube.com": true,
	"youtu.be":          true,
}

// ParseURL extracts a YouTube video ID from a user-supplied URL. It never
// fetches the URL, so arbitrary hosts cannot be used for SSRF.
func ParseURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 2048 {
		return "", domain.ErrInvalidURL
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || !allowedHosts[strings.ToLower(u.Hostname())] {
		return "", domain.ErrInvalidURL
	}
	if strings.ToLower(u.Hostname()) == "youtu.be" {
		id := strings.Trim(strings.Split(strings.TrimPrefix(u.Path, "/"), "/")[0], " ")
		if idPattern.MatchString(id) {
			return id, nil
		}
		return "", domain.ErrInvalidURL
	}
	if id := u.Query().Get("v"); idPattern.MatchString(id) {
		return id, nil
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) >= 2 && (parts[0] == "shorts" || parts[0] == "embed" || parts[0] == "live") {
		if idPattern.MatchString(parts[1]) {
			return parts[1], nil
		}
	}
	return "", domain.ErrInvalidURL
}

func ValidID(id string) bool { return idPattern.MatchString(id) }
