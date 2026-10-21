package architecture

import (
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"
)

const modulePrefix = "fana-karaoke/"

// TestGoHexagonalDependencies enforces the DDD import matrix for Go code.
func TestGoHexagonalDependencies(t *testing.T) {
	root := repoRoot(t)
	for _, path := range filesWithExt(t, root, ".go") {
		checkGoImports(t, root, path)
	}
}

func checkGoImports(t *testing.T, root, path string) {
	t.Helper()
	rel := relPath(root, path)
	if strings.HasSuffix(rel, "_test.go") {
		return
	}
	imports, err := parseImports(path)
	if err != nil {
		t.Fatalf("%s: %v", rel, err)
	}
	layer := goLayer(rel)
	for _, imp := range imports {
		if err := importAllowed(layer, imp); err != nil {
			t.Errorf("%s imports %q: %v", rel, imp, err)
		}
	}
}

func parseImports(path string) ([]string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(file.Imports))
	for _, spec := range file.Imports {
		value, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		out = append(out, value)
	}
	return out, nil
}

// goLayer maps a repository-relative path to its hexagonal layer.
func goLayer(rel string) string {
	for _, layer := range []string{"internal/domain", "internal/ports", "internal/usecase", "internal/adapters", "cmd/"} {
		if strings.HasPrefix(rel, layer) {
			return strings.TrimSuffix(layer, "/")
		}
	}
	return "other"
}

func importAllowed(layer, imp string) error {
	internal := strings.HasPrefix(imp, modulePrefix)
	target := strings.TrimPrefix(imp, modulePrefix)
	switch layer {
	case "internal/domain":
		return domainImportError(imp, internal)
	case "internal/ports":
		return portsImportError(internal, target)
	case "internal/usecase":
		return usecaseImportError(internal, target)
	case "internal/adapters":
		return adapterImportError(internal, target)
	}
	return nil
}

var impureStdlib = map[string]bool{
	"net/http": true, "sync": true, "crypto/rand": true,
	"encoding/json": true, "database/sql": true, "os": true, "log/slog": true,
}

func domainImportError(imp string, internal bool) error {
	if internal {
		return errDomain("the domain core must not import application packages")
	}
	if impureStdlib[imp] {
		return errDomain("the domain core must stay pure (no " + imp + ")")
	}
	if strings.Contains(strings.Split(imp, "/")[0], ".") {
		return errDomain("the domain core must not import third-party packages")
	}
	return nil
}

func portsImportError(internal bool, target string) error {
	if !internal {
		return nil
	}
	if target == "internal/domain" {
		return nil
	}
	return errDomain("ports may only depend on the domain")
}

func usecaseImportError(internal bool, target string) error {
	if !internal {
		return nil
	}
	if target == "internal/domain" || target == "internal/ports" {
		return nil
	}
	return errDomain("use cases may only depend on the domain and the ports")
}

func adapterImportError(internal bool, target string) error {
	if !internal {
		return nil
	}
	switch {
	case target == "internal/domain", target == "internal/ports":
		return nil
	case strings.HasPrefix(target, "internal/adapters"):
		return errDomain("adapters must not import one another")
	default:
		return errDomain("adapters must reach the core through ports only (imported " + target + ")")
	}
}

type archError string

func (e archError) Error() string { return string(e) }

func errDomain(msg string) error { return archError(msg) }
