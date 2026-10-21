//go:build ci

package main

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestUISelectionAndHeadlessCapture(t *testing.T) {
	for _, implementation := range []string{"default", "next", "legacy"} {
		t.Run(implementation, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
			t.Setenv("XDG_CACHE_HOME", filepath.Join(dir, "cache"))
			capture := filepath.Join(dir, "screen.png")
			args := []string{"--database", filepath.Join(dir, "lessons.sqlite3"), "--smoke-test", "--screenshot", capture}
			if implementation != "default" {
				args = append(args, "--ui", implementation)
			}
			if err := run(args); err != nil {
				t.Fatal(err)
			}
			file, err := os.Open(capture)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			config, err := png.DecodeConfig(file)
			if err != nil {
				t.Fatal(err)
			}
			if config.Width < 800 || config.Height < 480 {
				t.Fatalf("unexpected capture size: %+v", config)
			}
		})
	}
}
