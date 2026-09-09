package main

import (

	"errors"
	"flag"

	"os"
	"path/filepath"
	"rectablet/internal/workflows/swimming/transfer"
	"testing"

	store "rectablet/internal/workflows/swimming/repository"

)

func TestCLIValidation(t *testing.T) {
	for _, args := range [][]string{
		{"--screenshot", "unused.png"},
		{"unexpected"},
		{"adapt-rosters"},
		{"extract-curriculum"},
	} {
		if err := run(args); err == nil {
			t.Fatalf("accepted invalid arguments: %v", args)
		}
	}
	if err := run([]string{"--help"}); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("help: %v", err)
	}
}

func TestAdapterCLIProducesImportableFile(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "classes.json")
	args := []string{"adapt-rosters", "--dataset-id", "demo-autumn-2026", "--instructor", "Alex", "../../examples/sister-roster.json", output}
	if err := run(args); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	p, err := transfer.DecodePackage(body)
	if err != nil {
		t.Fatal(err)
	}
	if p.DatasetID != "demo-autumn-2026" || p.Courses[0].Swimmers[0].SourceID != "demo-swimmer-001" {
		t.Fatalf("identity changed: %+v", p)
	}
	db, err := store.Open(filepath.Join(dir, "lessons.sqlite3"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.ImportFile(output); err != nil {
		t.Fatal(err)
	}
	// A failed conversion must leave an existing export intact.
	args[len(args)-2] = filepath.Join(dir, "missing.json")
	if err = run(args); err == nil {
		t.Fatal("missing source accepted")
	}
	after, err := os.ReadFile(output)
	if err != nil || string(after) != string(body) {
		t.Fatal("failed conversion replaced output")
	}
}
