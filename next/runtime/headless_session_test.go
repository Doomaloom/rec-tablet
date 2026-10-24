//go:build ci

package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStartupShowsLoginWithUnavailablePreviousProfile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(dir, "cache"))
	path := filepath.Join(dir, "lessons.sqlite3")
	profiles, err := newProfileManager(path)
	checkSession(t, err)
	profile, err := profiles.importPackage(profilePackage("Alex"))
	checkSession(t, err)
	checkSession(t, profiles.activate(profile.Key))
	checkSession(t, os.Remove(profile.Path))
	// A persisted selection is not authorization or a startup dependency. Even
	// an unavailable previous instructor database must allow the keypad to open.
	capture := filepath.Join(dir, "login.png")
	checkSession(t, Run(Options{Database: path, SmokeTest: true, Screenshot: capture}))
	if _, err := os.Stat(capture); err != nil {
		t.Fatal("login was not rendered:", err)
	}
}
