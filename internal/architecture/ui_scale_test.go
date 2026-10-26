package architecture

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var (
	reFontSize  = regexp.MustCompile(`font-size:\s*([^;]+)`)
	reSpacing   = regexp.MustCompile(`(?:gap|padding|margin):\s*([^;]+)`)
	reOffScale  = regexp.MustCompile(`\d*\.?\d+(?:px|rem)`)
	reTypeToken = regexp.MustCompile(`var\(--text-[a-z0-9-]+\)`)
)

// TestUIScaleDrift reports type and spacing values that bypass the scales in
// DESIGN.md. It is advisory until SLOP_STRICT=1, so the list can be burned down.
func TestUIScaleDrift(t *testing.T) {
	root := repoRoot(t)
	strict := os.Getenv("SLOP_STRICT") == "1"
	total := 0
	for _, path := range filesWithExt(t, root, ".svelte") {
		rel := relPath(root, path)
		if !strings.HasPrefix(rel, "web/src/ui/") {
			continue
		}
		findings := scaleFindings(readLines(t, path))
		total += len(findings)
		reportScaleFindings(t, rel, findings, strict)
	}
	t.Logf("ui-scale: %d value(s) outside the scale", total)
}

func reportScaleFindings(t *testing.T, rel string, findings []string, strict bool) {
	t.Helper()
	if len(findings) == 0 {
		return
	}
	message := rel + ": " + strings.Join(head(findings, 3), " | ")
	if len(findings) > 3 {
		message += " (+" + strconv.Itoa(len(findings)-3) + " more)"
	}
	if strict {
		t.Error(message)
		return
	}
	t.Log(message)
}

func scaleFindings(lines []string) []string {
	var findings []string
	for i, line := range lines {
		if isCommentOrBlank(line) {
			continue
		}
		if literal := offScaleType(line); literal != "" {
			findings = append(findings, strconv.Itoa(i+1)+": font-size "+literal)
		}
		if literal := offScaleSpacing(line); literal != "" {
			findings = append(findings, strconv.Itoa(i+1)+": space "+literal)
		}
	}
	return findings
}

func offScaleType(line string) string {
	match := reFontSize.FindStringSubmatch(line)
	if match == nil {
		return ""
	}
	value := strings.TrimSpace(match[1])
	if reTypeToken.MatchString(value) || strings.Contains(value, "clamp(") {
		return ""
	}
	return value
}

// offScaleSpacing flags numeric spacing that ignores var(--space-*) and calc().
func offScaleSpacing(line string) string {
	match := reSpacing.FindStringSubmatch(line)
	if match == nil {
		return ""
	}
	value := strings.TrimSpace(match[1])
	if strings.Contains(value, "var(") || strings.Contains(value, "calc(") {
		return ""
	}
	if !reOffScale.MatchString(value) {
		return ""
	}
	return value
}

func isCommentOrBlank(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*")
}

func head(values []string, n int) []string {
	if len(values) <= n {
		return values
	}
	return values[:n]
}
