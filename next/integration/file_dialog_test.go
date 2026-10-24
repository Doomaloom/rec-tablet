package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"rectablet/src/shared/components/actionbutton"
)

// Traverse public canvas/widget APIs so the tests exercise the actual picker
// and callback without depending on Fyne's private dialog layout structs.
func dialogWidget[T fyne.CanvasObject](t *testing.T, root fyne.CanvasObject, matches func(T) bool) T {
	t.Helper()
	seen := map[fyne.CanvasObject]bool{}
	var visit func(fyne.CanvasObject) (T, bool)
	visit = func(obj fyne.CanvasObject) (T, bool) {
		var zero T
		if obj == nil || seen[obj] {
			return zero, false
		}
		seen[obj] = true
		if candidate, ok := obj.(T); ok && matches(candidate) {
			return candidate, true
		}
		var children []fyne.CanvasObject
		if c, ok := obj.(*fyne.Container); ok {
			children = c.Objects
		} else if w, ok := obj.(fyne.Widget); ok {
			children = test.TempWidgetRenderer(t, w).Objects()
		}
		for _, child := range children {
			if result, ok := visit(child); ok {
				return result, true
			}
		}
		return zero, false
	}
	result, ok := visit(root)
	if !ok {
		t.Fatalf("dialog widget not found: %T", result)
	}
	return result
}

func TestSetupDoesNotExposeLoginBypasses(t *testing.T) {
	f := setupUI(t)
	check(t, f.c.Maintenance())
	for _, id := range []string{"row:import", "row:share", "row:profile:saved-session"} {
		if f.view.Buttons[id] != nil {
			t.Fatalf("setup exposes %s", id)
		}
	}
	if f.c.Import("../../examples/classes.json") == nil {
		t.Fatal("manual import bypassed login")
	}
}

func TestExportThroughFolderPicker(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	f := setupUI(t)
	check(t, f.c.Maintenance())
	// The export chooser had the same Resize-before-Show crash as import.
	f.tap(t, "row:export")
	overlay := f.w.Canvas().Overlays().Top()
	open := dialogWidget(t, overlay, func(b *widget.Button) bool { return b.Text == "Open" })
	test.Tap(open)
	form := f.w.Canvas().Overlays().Top()
	save := dialogWidget(t, form, func(b *actionbutton.Button) bool { return b.AccessibilityLabel() == "Save" })
	test.Tap(save)
	if f.c.Snapshot().Error {
		t.Fatal(f.c.Snapshot().Notice)
	}
	if _, err := os.Stat(filepath.Join(home, "rec-tablet-results.json")); err != nil {
		t.Fatal("export was not saved in the selected folder:", err)
	}
}
