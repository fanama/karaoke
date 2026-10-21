package architecture

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

const maxFuncLines = 40

func TestGoFunctionLineLimits(t *testing.T) {
	root := repoRoot(t)
	for _, path := range filesWithExt(t, root, ".go") {
		checkGoFunctions(t, root, path)
	}
}

func checkGoFunctions(t *testing.T, root, path string) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("%s: %v", relPath(root, path), err)
	}
	lines := readLines(t, path)
	ast.Inspect(file, func(node ast.Node) bool {
		start, end, ok := funcRange(fset, node)
		if ok {
			checkFuncSize(t, relPath(root, path), fset.Position(node.Pos()).Line, start, end, lines)
		}
		return true
	})
}

func funcRange(fset *token.FileSet, node ast.Node) (int, int, bool) {
	switch fn := node.(type) {
	case *ast.FuncDecl:
		return fset.Position(fn.Pos()).Line, fset.Position(fn.End()).Line, true
	case *ast.FuncLit:
		return fset.Position(fn.Pos()).Line, fset.Position(fn.End()).Line, true
	}
	return 0, 0, false
}

func checkFuncSize(t *testing.T, rel string, line, start, end int, fileLines []string) {
	t.Helper()
	if endl := end - start + 1; endl > maxFuncLines {
		t.Errorf("%s:%d: function spans %d raw lines (max %d)", rel, line, endl, maxFuncLines)
	}
	if start < 1 || end > len(fileLines) {
		return
	}
	if got := sloc(fileLines[start-1 : end]); got > maxFuncLines {
		t.Errorf("%s:%d: function has %d SLOC (max %d)", rel, line, got, maxFuncLines)
	}
}
