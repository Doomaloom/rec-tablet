package setup

import (
	"errors"
	"rectablet/src/workflows/swimming/navigation"
	"testing"
)

type transferStore struct {
	err        error
	exportPath string
}

func (s *transferStore) Meta(string) (string, error)    { return "", nil }
func (s *transferStore) DatabasePath() string           { return "/device/lessons.sqlite3" }
func (s *transferStore) ImportFile(string) (int, error) { return 2, s.err }
func (s *transferStore) Dates() ([]string, error)       { return []string{"2026-09-04"}, nil }
func (s *transferStore) Export(path string) error       { s.exportPath = path; return s.err }

func TestTransferActionsPreserveErrorsAndNormalizeExportName(t *testing.T) {
	nav := navigation.New()
	s := &transferStore{}
	c := New(nav, s)
	nav.Features[navigation.Maintenance] = c
	nav.Push(navigation.Maintenance, nil)
	nav.Date = "2026-09-11"
	s.err = errors.New("invalid package")
	if c.Import("bad.json") == nil || nav.Date != "2026-09-11" || !nav.State.Error {
		t.Fatal("failed import lost error or date")
	}
	s.err = nil
	if err := c.Import("valid.json"); err != nil {
		t.Fatal(err)
	}
	if nav.Date != "2026-09-04" || nav.State.Error {
		t.Fatal("import did not choose loaded date")
	}
	if err := c.Export("/tmp/results.txt"); err != nil {
		t.Fatal(err)
	}
	if s.exportPath != "/tmp/results.json" {
		t.Fatal("export extension changed")
	}
}
