package architecture

import (
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

type colorPair struct {
	fg, bg string
}

// tokenPairs are the foreground/background combinations DESIGN.md documents.
var tokenPairs = []colorPair{
	{"text", "bg"},
	{"text", "surface"},
	{"muted", "bg"},
	{"muted", "surface-2"},
	{"dim", "bg"},
	{"accent", "bg"},
	{"accent-ink", "accent"},
	{"ink", "paper"},
	{"ink-muted", "paper"},
	{"danger", "bg"},
}

var hexToken = regexp.MustCompile(`--([a-z0-9-]+):\s*(#[0-9a-fA-F]{6})\s*;`)

// TestTokenContrast reports WCAG ratios for every documented token pair and
// fails under UI_CONTRAST_STRICT=1 while the palette settles.
func TestTokenContrast(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "web", "src", "styles", "tokens.css")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read tokens: %v", err)
	}
	tokens := parseHexTokens(string(data))
	strict := os.Getenv("UI_CONTRAST_STRICT") == "1"
	for _, pair := range tokenPairs {
		ratio, ok := ratioFor(tokens, pair)
		if !ok {
			t.Errorf("tokens.css: missing colour for %s on %s", pair.fg, pair.bg)
			continue
		}
		t.Logf("%s on %s: %.2f:1", pair.fg, pair.bg, ratio)
		if strict && ratio < 4.5 {
			t.Errorf("%s on %s is %.2f:1, below 4.5:1", pair.fg, pair.bg, ratio)
		}
	}
}

func parseHexTokens(css string) map[string]string {
	tokens := map[string]string{}
	for _, match := range hexToken.FindAllStringSubmatch(css, -1) {
		tokens[match[1]] = match[2]
	}
	return tokens
}

func ratioFor(tokens map[string]string, pair colorPair) (float64, bool) {
	fg, okFg := tokens[pair.fg]
	bg, okBg := tokens[pair.bg]
	if !okFg || !okBg {
		return 0, false
	}
	return contrastRatio(fg, bg), true
}

func contrastRatio(a, b string) float64 {
	la, lb := relativeLuminance(a), relativeLuminance(b)
	hi, lo := math.Max(la, lb), math.Min(la, lb)
	return (hi + 0.05) / (lo + 0.05)
}

func relativeLuminance(hex string) float64 {
	r, g, b := hexChannels(hex)
	return 0.2126*linearChannel(r) + 0.7152*linearChannel(g) + 0.0722*linearChannel(b)
}

func hexChannels(hex string) (float64, float64, float64) {
	value := strings.TrimPrefix(hex, "#")
	number, err := strconv.ParseUint(value, 16, 32)
	if err != nil {
		return 0, 0, 0
	}
	r := float64((number>>16)&0xff) / 255
	g := float64((number>>8)&0xff) / 255
	b := float64(number&0xff) / 255
	return r, g, b
}

func linearChannel(value float64) float64 {
	if value <= 0.04045 {
		return value / 12.92
	}
	return math.Pow((value+0.055)/1.055, 2.4)
}
