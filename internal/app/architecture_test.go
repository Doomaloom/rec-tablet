package app

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestPackageBoundaries(t *testing.T) {
	err := filepath.WalkDir("..", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		rel := filepath.ToSlash(strings.TrimPrefix(path, "../"))
		for _, imp := range file.Imports {
			dep, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return err
			}
			if strings.HasPrefix(rel, "shell/") || strings.HasPrefix(rel, "shared/") || strings.HasPrefix(rel, "platform/") {
				if strings.Contains(dep, "/workflows/") {
					t.Errorf("%s depends on workflow %s", rel, dep)
				}
			}
			if strings.HasPrefix(rel, "workflows/swimming/domain/") && (strings.Contains(dep, "fyne.io/") || strings.HasSuffix(dep, "/repository") || dep == "database/sql") {
				t.Errorf("domain depends on infrastructure: %s", dep)
			}
			for _, feature := range []string{"classes", "attendance", "assessment", "feedback", "setup"} {
				if !strings.HasPrefix(rel, "workflows/swimming/"+feature+"/") {
					continue
				}
				if strings.HasSuffix(dep, "/repository") || dep == "database/sql" {
					t.Errorf("feature %s bypasses its repository interface", rel)
				}
				for _, sibling := range []string{"classes", "attendance", "assessment", "feedback", "setup"} {
					if sibling != feature && strings.HasSuffix(dep, "/swimming/"+sibling) {
						t.Errorf("feature %s imports sibling %s", rel, dep)
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
