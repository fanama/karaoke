package sqlite

import "strings"

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

// searchText is the indexed form of a song title and channel.
func searchText(title, channel string) string {
	return normalize(title + " " + channel)
}

// searchTokens turns a user query into the words matched against search_text.
func searchTokens(query string) []string {
	fields := strings.Fields(normalize(query))
	if len(fields) > maxTokens {
		fields = fields[:maxTokens]
	}
	return fields
}

// normalize lowercases, folds accents and replaces punctuation with spaces.
func normalize(raw string) string {
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
