package shell_test

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
)

// Equipment is deliberately unrelated to lessons, enrollments, and curricula.
// It exists only in this test; the production app still opens My Classes.
type equipment struct {
	*shell.History[string]
	changed func()
	left    int
}

func (e *equipment) Snapshot() shell.State {
	route := e.Frames[len(e.Frames)-1]
	return shell.State{Screen: route, Title: "Equipment", CanBack: len(e.Frames) > 1,
		PageCount: 1, OnLeave: func() { e.left++ }, ComposeBody: func(s *ui.Surface) {
			ui.List(s, []ui.Item{{ID: "check", Label: "Check equipment"}}, func(string) {
				e.Append("equipment/check")
				e.changed()
			})
		}}
}
func (e *equipment) Subscribe(fn func()) { e.changed = fn }
func (e *equipment) Home() error         { e.Reset(); e.changed(); return nil }
func (e *equipment) Back() error         { e.Pop(); e.changed(); return nil }
func (e *equipment) Page(int) error      { return nil }
func (e *equipment) Header() error       { return nil }
func (e *equipment) Submit() error       { return nil }
func (e *equipment) Dismiss()            {}
func (e *equipment) OpenSetup() error    { return nil }

func TestIndependentWorkflowNavigationAndDisposal(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Independent workflow")
	defer w.Close()
	e := &equipment{History: shell.NewHistory("equipment")}
	v := shell.New(e, w)
	defer v.Close()
	w.Show()
	old := v.Buttons["row:check"]
	old.CanActivate = func() bool { return true }
	test.Tap(old)
	if e.Snapshot().Screen != "equipment/check" || e.left != 1 {
		t.Fatalf("navigation/disposal: %+v", e)
	}
	// Detached controls must not fire actions after a screen transition.
	test.Tap(old)
	if len(e.Frames) != 2 {
		t.Fatal("a disposed screen accepted a tap")
	}
	v.Buttons["back"].CanActivate = func() bool { return true }
	test.Tap(v.Buttons["back"])
	if e.Snapshot().Screen != "equipment" || e.left != 2 {
		t.Fatal("Back failed")
	}
	if err := e.Home(); err != nil {
		t.Fatal(err)
	}
	if e.left != 2 {
		t.Fatal("redrawing the same route disposed its resources")
	}
	v.Close()
	v.Close()
	if e.left != 3 || e.changed != nil {
		t.Fatal("Close did not dispose/unsubscribe exactly once")
	}
}
