package next_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestIndependentUIAndComponentOwnership(t *testing.T) {
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		for _, imp := range file.Imports {
			dep, _ := strconv.Unquote(imp.Path.Value)
			if dep == "rectablet/src/app" || dep == "rectablet/src/shell" {
				t.Errorf("%s imports legacy UI %s", path, dep)
			}
			for _, feature := range []string{"classes", "attendance", "assessment", "feedback", "setup", "navigation"} {
				prefix := "rectablet/src/workflows/swimming/" + feature
				if dep == prefix || strings.HasPrefix(dep, prefix+"/") {
					t.Errorf("%s imports legacy feature %s", path, dep)
				}
			}
		}
		if strings.HasPrefix(path, "features/") {
			ast.Inspect(file, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				if sel.Sel.Name == "IsLarge" {
					dir := strings.Split(filepath.ToSlash(path), "/")
					if len(dir) != 3 || filepath.Base(path) != "template.go" {
						t.Errorf("display composition must be selected in the feature root template: %s", path)
					}
				}
				return true
			})
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, feature := range []string{"classes", "attendance", "assessment", "feedback", "setup"} {
		err := filepath.WalkDir(filepath.Join("features", feature), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() || strings.Contains(filepath.ToSlash(path), "/internal") {
				return nil
			}
			for _, name := range []string{"controller.go", "template.go"} {
				if _, err := os.Stat(filepath.Join(path, name)); err != nil {
					t.Errorf("component %s lacks %s", path, name)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
