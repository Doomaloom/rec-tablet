package app

import (
	"testing"

	"fyne.io/fyne/v2/test"

	"rectablet/next/contracts"
)

func TestFeatureOverridesAreScopedAndDisposed(t *testing.T) {
	a := test.NewApp()
	t.Cleanup(a.Quit)
	w := a.NewWindow("overrides")
	t.Cleanup(w.Close)
	calls := 0
	action := func() error { calls++; return nil }
	p := &compositionPresenter{state: contracts.State{
		Screen: "custom", CanBack: true, PageCount: 2,
		LeftRailActions:  []contracts.RailAction{{ID: "custom-left", Label: "Save", Run: action}},
		RightRailActions: []contracts.RailAction{{ID: "custom-right", Label: "Help", Run: action}},
		HeaderActions:    []contracts.HeaderAction{{Label: "Custom", Run: action}},
	}}
	v := New(p, w)
	t.Cleanup(v.Close)
	if v.Buttons["home"] != nil || v.Buttons["back"] != nil || v.Buttons["up"] != nil || v.Buttons["down"] != nil {
		t.Fatal("default rails leaked into overridden rails")
	}
	old := v.Buttons["custom-left"]
	for _, id := range []string{"custom-left", "custom-right", "header-action-0"} {
		b := v.Buttons[id]
		if b == nil {
			t.Fatalf("missing %s", id)
		}
		b.CanActivate = func() bool { return true }
		test.Tap(b)
	}
	if calls != 3 {
		t.Fatalf("custom callbacks: %d", calls)
	}
	p.state = contracts.State{Screen: "standard", PageCount: 1}
	p.changed()
	for _, id := range []string{"custom-left", "custom-right", "header-action-0"} {
		if v.Buttons[id] != nil {
			t.Errorf("override %s survived navigation", id)
		}
	}
	if v.Buttons["home"] == nil || v.Buttons["up"] == nil || v.Buttons["down"] == nil {
		t.Fatal("default rails did not return")
	}
	test.Tap(old)
	if calls != 3 {
		t.Fatal("detached override accepted a tap")
	}
	p.state.LeftRailActions = []contracts.RailAction{}
	p.state.RightRailActions = []contracts.RailAction{}
	p.changed()
	if v.Buttons["home"] != nil || v.Buttons["up"] != nil || v.Buttons["down"] != nil {
		t.Fatal("explicit empty rails did not hide controls")
	}
}
