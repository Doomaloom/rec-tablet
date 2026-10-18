package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"rectablet/src/shared/components/actionbutton"
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	lesson "rectablet/src/workflows/swimming"
	store "rectablet/src/workflows/swimming/repository"
	"rectablet/src/workflows/swimming/transfer"
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

func TestImportThroughFilePicker(t *testing.T) {
	sources := map[string]string{"sample": "../../../../examples/classes.json"}
	if path := os.Getenv("REC_TABLET_IMPORT_FILE"); path != "" {
		sources["supplied_export"] = path
	}
	for name, source := range sources {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			data := must(os.ReadFile(source))
			check(t, os.WriteFile(filepath.Join(home, "classes.json"), data, 0600))
			expected := must(transfer.DecodePackage(data))
			a := test.NewApp()
			a.Settings().SetTheme(ui.AppTheme{})
			s := must(store.Open(filepath.Join(t.TempDir(), "lessons.sqlite3")))
			c := lesson.New(s)
			w := a.NewWindow("Import test")
			v := shell.New(c, w)
			w.Show()
			t.Cleanup(func() { v.Close(); w.Close(); a.Quit(); s.Close() })
			f := &uiFixture{v, c, s, w}
			check(t, c.Maintenance())
			// Before the fix this tap panics inside FileDialog.Resize.
			f.tap(t, "row:import")
			overlay := w.Canvas().Overlays().Top()
			grid := dialogWidget(t, overlay, func(*widget.GridWrap) bool { return true })
			// The isolated home contains only classes.json plus the parent item.
			grid.Select(1)
			open := dialogWidget(t, overlay, func(b *widget.Button) bool { return b.Text == "Open" })
			if open.Disabled() {
				t.Fatal("cannot select the class package")
			}
			test.Tap(open)
			if c.Snapshot().Error {
				t.Fatal(c.Snapshot().Notice)
			}
			if w.Canvas().Overlays().Top() != nil {
				t.Fatal("picker did not close")
			}
			var count int
			check(t, s.DB.QueryRow("SELECT count(*) FROM courses").Scan(&count))
			if count != len(expected.Courses) {
				t.Fatalf("imported %d courses; expected %d", count, len(expected.Courses))
			}
			f.tap(t, "home")
			if c.Snapshot().RowCount == 0 {
				t.Fatal("successful import left the class list empty")
			}
			t.Logf("File picker imported %d courses; selected lesson date %s", count, c.Date())
		})
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

func TestSessionCodeTouchKeypadPreservesLeadingZeroes(t *testing.T) {
	f := setupUI(t)
	check(t, f.c.Maintenance())
	var received string
	f.c.ConfigureDeviceSessions(func(code string, done func(error)) { received = code; done(nil) }, nil, nil, "https://share.example", nil, nil)
	f.tap(t, "row:share")
	overlay := f.w.Canvas().Overlays().Top()
	f.shot(t, "share-code")
	for _, digit := range []string{"0", "0", "1", "2", "3", "4"} {
		button := dialogWidget(t, overlay, func(b *actionbutton.Button) bool { return b.AccessibilityLabel() == digit })
		test.Tap(button)
	}
	download := dialogWidget(t, overlay, func(b *actionbutton.Button) bool { return b.AccessibilityLabel() == "Download" })
	// Confirmation must remain reachable on the 800x480 display even when the
	// keypad body needs scrolling.
	position := fyne.CurrentApp().Driver().AbsolutePositionForObject(download)
	if position.Y < 0 || position.Y+download.Size().Height > f.w.Canvas().Size().Height {
		t.Fatalf("download action is outside the viewport: %v size %v", position, download.Size())
	}
	test.Tap(download)
	if received != "001234" {
		t.Fatalf("received instructor code %q", received)
	}
}
