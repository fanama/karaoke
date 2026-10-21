package architecture

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var importSpecifier = regexp.MustCompile(`(?m)(?:from\s+|import\s+)['"]([^'"]+)['"]`)

// TestFrontendHexagonalDependencies enforces the UI/application/adapters split.
func TestFrontendHexagonalDependencies(t *testing.T) {
	root := repoRoot(t)
	files := filesWithExt(t, root, ".ts")
	files = append(files, filesWithExt(t, root, ".svelte")...)
	for _, path := range files {
		rel := relPath(root, path)
		if !strings.HasPrefix(rel, "web/src/") {
			continue
		}
		checkFrontendImports(t, rel, path)
	}
}

func checkFrontendImports(t *testing.T, rel, path string) {
	t.Helper()
	layer := frontendLayer(rel)
	if layer == "" {
		return
	}
	for _, spec := range importSpecs(path) {
		if strings.HasPrefix(spec, "svelte") {
			checkSvelteImport(t, rel, layer, spec)
			continue
		}
		if !strings.HasPrefix(spec, ".") {
			continue
		}
		target := frontendLayer(resolveSourcePath(rel, spec))
		if err := frontendImportAllowed(layer, target); err != nil {
			t.Errorf("%s imports %q: %v", rel, spec, err)
		}
	}
}

func importSpecs(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	matches := importSpecifier.FindAllStringSubmatch(string(data), -1)
	out := make([]string, 0, len(matches))
	for _, match := range matches {
		out = append(out, match[1])
	}
	return out
}

func checkSvelteImport(t *testing.T, rel, layer, spec string) {
	t.Helper()
	if spec == "svelte" || strings.HasPrefix(spec, "svelte/") || strings.HasPrefix(spec, "svelte-") {
		return
	}
	if layer == "domain" || layer == "ports" {
		t.Errorf("%s imports %q: %s must stay framework free", rel, spec, layer)
	}
}

// frontendLayer maps a web/src relative path to its layer name.
func frontendLayer(rel string) string {
	rest := strings.TrimPrefix(rel, "web/src/")
	if rest == rel {
		return ""
	}
	switch strings.Split(rest, "/")[0] {
	case "domain", "ports", "application", "adapters", "ui":
		return strings.Split(rest, "/")[0]
	case "main.ts", "App.svelte":
		return "root"
	}
	return ""
}

func frontendImportAllowed(layer, target string) error {
	if target == "" || target == "root" {
		return nil
	}
	switch layer {
	case "domain":
		if target != "domain" {
			return archError("the domain layer must stay pure")
		}
	case "ports":
		if target != "domain" && target != "ports" {
			return archError("ports may only depend on the domain")
		}
	case "application":
		if target != "domain" && target != "ports" && target != "application" {
			return archError("application must not import adapters or ui")
		}
	case "adapters":
		if target == "ui" || target == "application" {
			return archError("adapters must not import ui or application")
		}
	case "ui":
		if target == "adapters" {
			return archError("ui must receive its ports through context, never adapters")
		}
	}
	return nil
}

func resolveSourcePath(fromRel, spec string) string {
	dir := filepath.Dir(fromRel)
	joined := filepath.ToSlash(filepath.Clean(filepath.Join(dir, spec)))
	return joined
}
