package setup

import (
	"rectablet/internal/shared/ui"
	"rectablet/internal/workflows/swimming/navigation"
	"testing"
)

func TestDisplayChoiceRefreshesSetupAndPersists(t *testing.T) {
	nav := navigation.New()
	c := New(nav, &transferStore{})
	nav.Features[navigation.Maintenance] = c
	nav.Push(navigation.Maintenance, nil)
	var saved ui.DisplayMode
	c.SaveDisplayMode = func(mode ui.DisplayMode) { saved = mode }
	for _, mode := range []ui.DisplayMode{ui.DisplayLarge, ui.DisplayCompact} {
		if err := c.SetDisplayMode(mode); err != nil {
			t.Fatal(err)
		}
		if nav.DisplayMode != mode || nav.State.DisplayMode != mode || saved != mode {
			t.Fatalf("mode not applied and persisted: %s / %s / %s", nav.DisplayMode, nav.State.DisplayMode, saved)
		}
		if nav.Frame().Screen != navigation.Maintenance {
			t.Fatal("display change navigated away")
		}
	}
	if err := c.SetDisplayMode("future-mode"); err != nil {
		t.Fatal(err)
	}
	if saved != ui.DisplayCompact {
		t.Fatal("invalid setting did not normalize to compact")
	}
}
