package architecture

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

var tsFunctionStart = regexp.MustCompile(`(?m)^[ \t]*(?:export\s+)?(?:async\s+)?function\s+[A-Za-z0-9_$]+\s*\(`)

var tsArrowStart = regexp.MustCompile(`(?m)^[ \t]*(?:export\s+)?(?:const|let)\s+[A-Za-z0-9_$]+\s*(?::[^=\n]+)?=\s*(?:async\s+)?(?:\([^)\n]*\)|[A-Za-z0-9_$]+)\s*=>\s*\{`)

var tsMethodStart = regexp.MustCompile(`(?m)^[ \t]{2,}(?:(?:private|public|protected|static|readonly|async)\s+)*[A-Za-z0-9_$]+\s*\([^)\n]*\)\s*(?::\s*[^{;\n]+)?\{`)

var tsKeywords = map[string]bool{
	"if": true, "for": true, "while": true, "switch": true,
	"catch": true, "function": true, "return": true,
}

// TestFrontendFunctionLineLimits checks every TypeScript function and Svelte
// script function stays within the 40 line budget.
func TestFrontendFunctionLineLimits(t *testing.T) {
	root := repoRoot(t)
	for _, path := range filesWithExt(t, root, ".ts") {
		checkTSFunctions(t, root, path, tsCode(t, path))
	}
	for _, path := range filesWithExt(t, root, ".svelte") {
		checkTSFunctions(t, root, path, svelteScript(t, path))
	}
}

func checkTSFunctions(t *testing.T, root, path, code string) {
	t.Helper()
	rel := relPath(root, path)
	starts := functionStarts(code)
	for _, start := range starts {
		end, ok := blockEnd(code, start)
		if !ok {
			continue
		}
		if end-start+1 > maxFuncLines {
			t.Errorf("%s:%d: function spans %d lines (max %d)", rel, start, end-start+1, maxFuncLines)
		}
	}
}

func functionStarts(code string) []int {
	lines := strings.Split(code, "\n")
	var starts []int
	for _, re := range []*regexp.Regexp{tsFunctionStart, tsArrowStart, tsMethodStart} {
		for _, loc := range re.FindAllStringIndex(code, -1) {
			line := 1 + strings.Count(code[:loc[0]], "\n")
			if keywordLine(lines[line-1]) {
				continue
			}
			starts = append(starts, line)
		}
	}
	return starts
}

func keywordLine(line string) bool {
	fields := strings.Fields(strings.TrimSpace(line))
	if len(fields) == 0 {
		return true
	}
	return tsKeywords[fields[0]]
}

// blockEnd returns the last line of the brace block starting after line start.
func blockEnd(code string, start int) (int, bool) {
	offset := lineOffset(code, start)
	open := strings.Index(code[offset:], "{")
	if open < 0 {
		return 0, false
	}
	depth := 0
	for i := offset + open; i < len(code); {
		if next, ok := skipTrivia(code, i); ok {
			i = next
			continue
		}
		switch code[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return 1 + strings.Count(code[:i], "\n"), true
			}
		}
		i++
	}
	return 0, false
}

// skipTrivia skips quoted strings and comments, returning the next index.
func skipTrivia(code string, i int) (int, bool) {
	if code[i] == '/' && i+1 < len(code) {
		if code[i+1] == '/' {
			return skipLine(code, i), true
		}
		if code[i+1] == '*' {
			return skipBlock(code, i), true
		}
	}
	if strings.ContainsRune("'\"`", rune(code[i])) {
		return skipString(code, i), true
	}
	return i, false
}

func skipLine(code string, i int) int {
	if next := strings.IndexByte(code[i:], '\n'); next >= 0 {
		return i + next
	}
	return len(code)
}

func skipBlock(code string, i int) int {
	if next := strings.Index(code[i+2:], "*/"); next >= 0 {
		return i + next + 4
	}
	return len(code)
}

func skipString(code string, i int) int {
	quote := code[i]
	for j := i + 1; j < len(code); j++ {
		switch code[j] {
		case '\\':
			j++
		case quote:
			return j + 1
		}
	}
	return len(code)
}

func lineOffset(code string, line int) int {
	count := 1
	for i := 0; i < len(code); i++ {
		if count == line {
			return i
		}
		if code[i] == '\n' {
			count++
		}
	}
	return 0
}

// tsCode returns the file content with Svelte templates blanked out.
func svelteScript(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	lines := strings.Split(source, "\n")
	inside := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "<script") {
			inside = true
			lines[i] = ""
			continue
		}
		if strings.HasPrefix(trimmed, "</script") {
			inside = false
			lines[i] = ""
			continue
		}
		if !inside {
			lines[i] = ""
		}
	}
	return strings.Join(lines, "\n")
}

func tsCode(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
