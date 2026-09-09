package swimming_test

import (
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"rectablet/internal/shared/ui"
	"rectablet/internal/shell"
	lesson "rectablet/internal/workflows/swimming"
	store "rectablet/internal/workflows/swimming/repository"
	"rectablet/internal/workflows/swimming/transfer"
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
	sources := map[string]string{"sample": "../../../examples/classes.json"}
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
			if c.State.Error {
				t.Fatal(c.State.Notice)
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
			if c.State.RowCount == 0 {
				t.Fatal("successful import left the class list empty")
			}
			t.Logf("File picker imported %d courses; selected lesson date %s", count, c.Date)
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
	save := dialogWidget(t, form, func(b *widget.Button) bool { return b.Text == "Save" })
	test.Tap(save)
	if f.c.State.Error {
		t.Fatal(f.c.State.Notice)
	}
	if _, err := os.Stat(filepath.Join(home, "rec-tablet-results.json")); err != nil {
		t.Fatal("export was not saved in the selected folder:", err)
	}
}
