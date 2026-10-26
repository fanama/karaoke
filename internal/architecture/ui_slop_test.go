package architecture

import (
	"regexp"
	"strings"
	"testing"
)

// uiSlopAllow marks a reviewed exception. It must name a reason and sit on the
// offending line or the line directly above it; the count is reported on every run.
var uiSlopAllow = regexp.MustCompile(`ui-slop-allow:\s*\S`)

var (
	reGradient = regexp.MustCompile(`(?:linear|radial|conic)-gradient`)
	reGlass    = regexp.MustCompile(`(?:backdrop-filter|filter:\s*blur)`)
	reGlyph    = regexp.MustCompile(`[\x{2190}-\x{2BFF}\x{2700}-\x{27BF}\x{1F000}-\x{1FAFF}]`)
	reRadius   = regexp.MustCompile(`border-radius:\s*(?:0|50%|var\(--radius-[a-z0-9-]+\))\s*;?`)
	reShadow   = regexp.MustCompile(`box-shadow:\s*var\(--shadow-[a-z0-9-]+\)\s*;?`)
)

// uiSlopRule is one ban applied to a set of files.
type uiSlopRule struct {
	name  string
	ext   string
	scope string
	check func(rel, line string) bool
}

func uiSlopRules() []uiSlopRule {
	line := func(re *regexp.Regexp) func(string, string) bool {
		return func(_, text string) bool { return re.MatchString(text) }
	}
	return []uiSlopRule{
		{"gradient in svelte", ".svelte", "web/src/", line(reGradient)},
		{"gradient in css", ".css", "web/src/", line(reGradient)},
		{"glass in svelte", ".svelte", "web/src/", line(reGlass)},
		{"glass in css", ".css", "web/src/", line(reGlass)},
		{"glow in svelte", ".svelte", "web/src/", glowCheck},
		{"glow in css", ".css", "web/src/", glowCheck},
		{"off-scale radius", ".svelte", "web/src/ui/", radiusCheck},
		{"glyph used as icon", ".svelte", "web/src/ui/", glyphCheck},
		{"filler copy", ".ts", "web/src/application/i18n/", copyCheck},
	}
}

// TestUISlopBans fails on the patterns DESIGN.md forbids.
func TestUISlopBans(t *testing.T) {
	root := repoRoot(t)
	exempt := 0
	for _, rule := range uiSlopRules() {
		for _, path := range filesWithExt(t, root, rule.ext) {
			rel := relPath(root, path)
			if !strings.HasPrefix(rel, rule.scope) {
				continue
			}
			exempt += checkUISlopFile(t, rule, rel, readLines(t, path))
		}
	}
	t.Logf("ui-slop: %d documented exemption(s)", exempt)
}

// checkUISlopFile reports every unsuppressed violation and returns the number of
// lines that were exempted by a `ui-slop-allow` comment.
func checkUISlopFile(t *testing.T, rule uiSlopRule, rel string, lines []string) int {
	t.Helper()
	exempt := 0
	for i, line := range lines {
		if !rule.check(rel, line) {
			continue
		}
		if allowedAt(lines, i) {
			exempt++
			continue
		}
		t.Errorf("%s:%d: %s: %s", rel, i+1, rule.name, strings.TrimSpace(line))
	}
	return exempt
}

func allowedAt(lines []string, i int) bool {
	if uiSlopAllow.MatchString(lines[i]) {
		return true
	}
	return i > 0 && uiSlopAllow.MatchString(lines[i-1])
}

// glowCheck bans text shadows and any box-shadow that is not a --shadow-* token.
func glowCheck(_, line string) bool {
	if strings.Contains(line, "text-shadow") {
		return true
	}
	if !strings.Contains(line, "box-shadow") {
		return false
	}
	return !reShadow.MatchString(line)
}

func radiusCheck(_, line string) bool {
	if !strings.Contains(line, "border-radius") {
		return false
	}
	return !reRadius.MatchString(line)
}

func glyphCheck(rel, line string) bool {
	if strings.HasSuffix(rel, "Icon.svelte") {
		return false
	}
	return reGlyph.MatchString(line)
}

// fillerPhrases are marketing tics, not product words. Bare verbs like "unlock"
// and "harness" are left out on purpose: this app really does unlock a queue.
var fillerPhrases = []string{
	"lorem ipsum", "seamless", "effortless", "elevate your", "unlock your",
	"unlock the power", "supercharge", "revolutionary", "cutting-edge",
	"game-changing", "next-level", "empower your", "leverage the power",
	"robust", "intuitive", "state-of-the-art", "best-in-class", "dive in",
	"unleash", "in today's", "whether you're", "take it to the next level",
	"get started", "learn more", "discover more",
}

func copyCheck(_, line string) bool {
	lower := strings.ToLower(line)
	for _, phrase := range fillerPhrases {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}
