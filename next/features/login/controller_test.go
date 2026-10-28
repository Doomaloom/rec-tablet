package login

import (
	"errors"
	"testing"
)

func TestCodeEntryAndLoginLifecycle(t *testing.T) {
	var received string
	var finish func(error)
	calls, cancelled := 0, 0
	c := New(Options{Login: func(code string, done func(error)) { calls++; received, finish = code, done }, Cancel: func() { cancelled++ }})
	_ = c.Submit()
	c.Digit("x")
	c.Digit("12")
	if c.code != "" || calls != 0 {
		t.Fatal("invalid code accepted")
	}
	c.Digit("9")
	c.Delete()
	c.Delete()
	for _, digit := range []string{"0", "0", "1", "2", "3", "4", "5"} {
		c.Digit(digit)
	}
	if c.code != "001234" {
		t.Fatal(c.code)
	}
	_ = c.Submit()
	_ = c.Submit()
	c.Delete()
	c.Clear()
	c.Digit("8")
	_ = c.OpenSetup()
	_ = c.Back()
	if calls != 1 || received != "001234" || !c.busy || c.settings || c.code != "001234" {
		t.Fatalf("in-flight state: %+v, calls %d", c, calls)
	}
	finish(errors.New("Sharing has ended. Ask for a new code."))
	if c.busy || !c.Snapshot().Error || c.Snapshot().Notice == "" {
		t.Fatal("failure not presented")
	}
	c.Delete()
	if c.error || c.code != "00123" {
		t.Fatal("cannot edit code after failure")
	}
	c.Digit("4")
	_ = c.Submit()
	if calls != 2 {
		t.Fatal("cannot retry")
	}
	c.Deactivate()
	finish(errors.New("late failure"))
	_ = c.Submit()
	c.Digit("1")
	if c.code != "" || calls != 2 || cancelled != 1 {
		t.Fatal("detached login remained active")
	}
	c.Deactivate()
	if cancelled != 1 {
		t.Fatal("close was not idempotent")
	}
}

func TestSettingsCannotExposeLessonNavigation(t *testing.T) {
	c := New(Options{})
	_ = c.Home()
	_ = c.Back()
	if c.Snapshot().Screen != "login" {
		t.Fatal("back escaped login")
	}
	_ = c.OpenSetup()
	if c.Snapshot().Screen != "login_setup" {
		t.Fatal("settings unavailable")
	}
	_ = c.Back()
	if c.Snapshot().Screen != "login" {
		t.Fatal("back did not return to login")
	}
	c.Digit("1")
	c.Clear()
	if c.code != "" {
		t.Fatal("clear did not reset code")
	}
}

func TestResumeLifecycle(t *testing.T) {
	calls := 0
	var finish func(error)
	c := New(Options{Resume: func(done func(error)) { calls++; finish = done }})
	c.Digit("9")
	c.Delete()
	if calls != 0 {
		t.Fatal("normal backspace resumed session")
	}
	c.Resume()
	c.Resume()
	if calls != 1 || !c.busy {
		t.Fatal("resume must guard duplicate requests")
	}
	finish(errors.New("No previous session"))
	if c.busy || !c.error {
		t.Fatal("restore error not shown")
	}
	_ = c.OpenSetup()
	c.Resume()
	if calls != 1 {
		t.Fatal("settings resumed session")
	}
	_ = c.Back()
	c.Resume()
	c.Deactivate()
	finish(errors.New("stale failure"))
	c.Resume()
	if calls != 2 {
		t.Fatal("detached login resumed session")
	}
}
