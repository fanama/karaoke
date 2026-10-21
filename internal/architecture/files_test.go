package architecture

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const maxLines = 300

// repoRoot walks up from this test file to the repository root.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate the test file")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

// sourceFiles lists every architecture-relevant source file.
func sourceFiles(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if skipDir(entry.Name()) {
				return fs.SkipDir
			}
			return nil
		}
		if isSource(entry.Name()) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func skipDir(name string) bool {
	switch name {
	case ".git", "node_modules", "dist", "testdata":
		return true
	}
	return false
}

func isSource(name string) bool {
	switch filepath.Ext(name) {
	case ".go", ".ts", ".svelte", ".css":
		return true
	}
	return false
}

func filesWithExt(t *testing.T, root, ext string) []string {
	t.Helper()
	var out []string
	for _, path := range sourceFiles(t, root) {
		if filepath.Ext(path) == ext {
			out = append(out, path)
		}
	}
	return out
}

func readLines(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
}

func relPath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}

// sloc counts lines that are neither blank nor comment-only.
func sloc(lines []string) int {
	count := 0
	inBlock := false
	for _, raw := range lines {
		if !isCommentLine(raw, &inBlock) {
			count++
		}
	}
	return count
}

func isCommentLine(raw string, inBlock *bool) bool {
	line := strings.TrimSpace(raw)
	if *inBlock {
		if strings.Contains(line, "*/") {
			*inBlock = false
		}
		return true
	}
	if line == "" {
		return true
	}
	if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
		return true
	}
	if strings.HasPrefix(line, "/*") {
		if !strings.Contains(line, "*/") {
			*inBlock = true
		}
		return true
	}
	return false
}

func TestFileLineLimits(t *testing.T) {
	root := repoRoot(t)
	for _, path := range sourceFiles(t, root) {
		checkFileSize(t, root, path)
	}
}

func checkFileSize(t *testing.T, root, path string) {
	t.Helper()
	lines := readLines(t, path)
	rel := relPath(root, path)
	if len(lines) > maxLines {
		t.Errorf("%s: %d raw lines (max %d)", rel, len(lines), maxLines)
	}
	if got := sloc(lines); got > maxLines {
		t.Errorf("%s: %d SLOC (max %d)", rel, got, maxLines)
	}
}
