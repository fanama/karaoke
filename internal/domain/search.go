package domain

import (
	"sort"
	"strings"
)

// MaxSearchTokens caps how many words a query contributes to a match.
const MaxSearchTokens = 6

var ligatures = strings.NewReplacer(
	"œ", "oe", "Œ", "oe",
	"æ", "ae", "Æ", "ae",
	"ß", "ss",
)

var accents = map[rune]rune{
	'à': 'a', 'á': 'a', 'â': 'a', 'ã': 'a', 'ä': 'a', 'å': 'a',
	'ç': 'c', 'è': 'e', 'é': 'e', 'ê': 'e', 'ë': 'e',
	'ì': 'i', 'í': 'i', 'î': 'i', 'ï': 'i',
	'ñ': 'n', 'ò': 'o', 'ó': 'o', 'ô': 'o', 'õ': 'o', 'ö': 'o',
	'ù': 'u', 'ú': 'u', 'û': 'u', 'ü': 'u', 'ý': 'y', 'ÿ': 'y',
	'ø': 'o', 'ł': 'l', 'ś': 's', 'š': 's', 'ž': 'z', 'ź': 'z', 'ż': 'z',
	'ā': 'a', 'ē': 'e', 'ī': 'i', 'ō': 'o', 'ū': 'u',
	'ğ': 'g', 'ı': 'i', 'ş': 's',
}

// SearchText is the normalized title and channel used by library search.
func SearchText(title, channel string) string {
	return NormalizeSearch(title + " " + channel)
}

// SearchTokens turns a query into the words matched against SearchText.
func SearchTokens(query string) []string {
	fields := strings.Fields(NormalizeSearch(query))
	if len(fields) > MaxSearchTokens {
		fields = fields[:MaxSearchTokens]
	}
	return fields
}

// NormalizeSearch lowercases, folds accents and replaces punctuation with spaces.
func NormalizeSearch(raw string) string {
	raw = ligatures.Replace(strings.ToLower(strings.TrimSpace(raw)))
	var b strings.Builder
	b.Grow(len(raw))
	for _, r := range raw {
		if folded, ok := accents[r]; ok {
			r = folded
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' {
			b.WriteRune(r)
			continue
		}
		b.WriteRune(' ')
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// RankSearch filters songs containing every query token and orders matches by
// exact phrase, then title length and title. It never returns more than limit.
func RankSearch(songs []SearchResult, query string, limit int) []SearchResult {
	tokens := SearchTokens(query)
	if len(tokens) == 0 || limit <= 0 {
		return nil
	}
	phrase := strings.Join(tokens, " ")
	matches := make([]rankedSong, 0, len(songs))
	for _, song := range songs {
		text := SearchText(song.Title, song.Channel)
		if containsAll(text, tokens) {
			matches = append(matches, rankedSong{song: song, exact: strings.Contains(text, phrase)})
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		return rankedLess(matches[i], matches[j])
	})
	if len(matches) > limit {
		matches = matches[:limit]
	}
	out := make([]SearchResult, len(matches))
	for i, match := range matches {
		out[i] = match.song
	}
	return out
}

type rankedSong struct {
	song  SearchResult
	exact bool
}

func containsAll(text string, tokens []string) bool {
	for _, token := range tokens {
		if !strings.Contains(text, token) {
			return false
		}
	}
	return true
}

func rankedLess(a, b rankedSong) bool {
	if a.exact != b.exact {
		return a.exact
	}
	if len(a.song.Title) != len(b.song.Title) {
		return len(a.song.Title) < len(b.song.Title)
	}
	return a.song.Title < b.song.Title
}
