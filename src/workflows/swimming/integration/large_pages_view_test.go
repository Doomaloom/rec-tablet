package integration_test

import (
	"fmt"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/transfer"
)

func TestLargeClassesCalendarMenuAndAttendancePages(t *testing.T) {
	f := setupUI(t)
	f.c.ConfigureDisplay(ui.DisplayLarge, nil)
	if got := f.w.Canvas().Size(); got != fyne.NewSize(1024, 600) {
		t.Fatalf("large display size = %v, want 1024×600", got)
	}

	pkg := must(transfer.ReadPackage("../../../../examples/classes.json"))
	for i := 0; i < 9; i++ {
		pkg.Courses[0].Swimmers = append(pkg.Courses[0].Swimmers, transfer.SwimmerInput{
			SourceID: fmt.Sprintf("large-roster-%02d", i), Name: fmt.Sprintf("Large Display Swimmer %02d", i),
		})
	}
	for i := 0; i < 4; i++ {
		course := pkg.Courses[0]
		course.SourceID = fmt.Sprintf("large-course-%02d", i)
		course.Name = fmt.Sprintf("Large Display Course %02d", i)
		course.Swimmers = []transfer.SwimmerInput{}
		pkg.Courses = append(pkg.Courses, course)
	}
	_ = must(f.s.Import(pkg))
	check(t, f.c.Refresh())
	if f.c.State.RowCount != 7 || f.c.State.PageCount != 2 {
		t.Fatalf("classes rows/pages = %d/%d, want 7/2", f.c.State.RowCount, f.c.State.PageCount)
	}
	assertLargeCards(t, f, 2, 3, 6)
	f.shot(t, "large-classes")

	f.tap(t, "header")
	assertLargeCalendar(t, f)
	f.shot(t, "large-calendar")
	if button := f.view.Buttons["row:2026-09-04"]; button == nil || !button.Selected || button.Disabled {
		t.Fatal("selected scheduled lesson date is not highlighted and selectable")
	}
	if button := f.view.Buttons["row:unscheduled-2026-09-05"]; button == nil || !button.Disabled {
		t.Fatal("unscheduled calendar date should be disabled")
	}
	f.tap(t, "calendar-next")
	if f.c.State.Screen != "dates" || f.view.Buttons["row:2026-09-04"] != nil {
		t.Fatal("next month did not replace the visible calendar month")
	}
	f.tap(t, "calendar-previous")
	if f.view.Buttons["row:2026-09-04"] == nil {
		t.Fatal("previous month did not return to the selected lesson month")
	}
	f.tap(t, "row:2026-09-11")
	if f.c.Date != "2026-09-11" || f.c.State.Screen != "classes" {
		t.Fatal("scheduled calendar date did not return to classes")
	}
	selectClassStartingWith(t, f, "Splash 1 ·")
	assertLargeCards(t, f, 3, 1, 3)
	f.shot(t, "large-lesson-menu")
	f.tap(t, "row:attendance")
	if f.c.State.RowCount != 13 || f.c.State.PageCount != 2 || len(f.c.State.Rows) != 12 {
		t.Fatalf("attendance rows/pages/current = %d/%d/%d, want 13/2/12", f.c.State.RowCount, f.c.State.PageCount, len(f.c.State.Rows))
	}
	assertLargeCards(t, f, 2, 6, 12)
	if !containsRosterName(f.c.State.Rows, "Alexandria Montgomery-Sutherland") {
		t.Fatalf("large attendance page omitted the long name: %+v", f.c.State.Rows)
	}
	f.shot(t, "large-attendance")
	f.tap(t, "down")
	if f.c.State.Page != 1 || len(f.c.State.Rows) != 1 {
		t.Fatalf("partial attendance page = %d, %d rows", f.c.State.Page, len(f.c.State.Rows))
	}
	assertLargeCards(t, f, 2, 6, 1)
	f.shot(t, "large-attendance-partial")
}

func selectClassStartingWith(t *testing.T, f *uiFixture, prefix string) {
	t.Helper()
	for {
		for i, row := range f.c.State.Rows {
			if strings.HasPrefix(row.Label, prefix) {
				f.row(t, i)
				return
			}
		}
		if f.c.State.Page >= f.c.State.PageCount-1 {
			t.Fatalf("missing class with label prefix %q", prefix)
		}
		f.tap(t, "down")
	}
}

func assertLargeCards(t *testing.T, f *uiFixture, columns, visualRows, expected int) {
	t.Helper()
	if expected > columns*visualRows {
		t.Fatalf("test expects %d cards in a %d×%d grid", expected, columns, visualRows)
	}
	composed := shell.ComposeScreen(f.view, f.c.Snapshot())
	count := 0
	var previousX, previousY float32
	bodyTop := ui.OuterMargin + ui.HeaderHeight + ui.Gap
	if f.c.State.SplitHeader != nil {
		bodyTop = ui.OuterMargin + shell.SplitHeaderHeight + ui.Gap
	}
	for _, component := range composed.Components {
		if !strings.HasPrefix(component.Name, "feature-list-row-") {
			continue
		}
		x, y, width, height := component.Position(1024, 600)
		if x < ui.ContentInset || y < bodyTop || width <= 0 || height < 48 || x+width > 1024-ui.ContentInset+0.01 || y+height > 600-ui.OuterMargin+0.01 {
			t.Fatalf("large card target too small or outside content: %s at %.1f,%.1f %.1fx%.1f", component.Name, x, y, width, height)
		}
		if count%columns == 0 && count > 0 && y <= previousY {
			t.Fatalf("card row %d did not move down: previous y %.1f, current y %.1f", count/columns, previousY, y)
		}
		if count%columns != 0 && (x <= previousX || y != previousY) {
			t.Fatalf("cards are not laid out row-major: previous (%.1f,%.1f), current (%.1f,%.1f)", previousX, previousY, x, y)
		}
		previousX, previousY = x, y
		count++
	}
	if count != expected {
		t.Fatalf("large card targets = %d, want %d", count, expected)
	}
}

func assertLargeCalendar(t *testing.T, f *uiFixture) {
	t.Helper()
	for _, id := range []string{"calendar-previous", "calendar-next", "row:weekday-0", "row:2026-09-04"} {
		if f.view.Buttons[id] == nil {
			t.Fatalf("large calendar missing %s", id)
		}
	}
	for _, id := range []string{"calendar-previous", "calendar-next"} {
		if size := f.view.Buttons[id].Size(); size.Width < 48 || size.Height < 48 {
			t.Fatalf("calendar control %s target is smaller than 48 px: %v", id, size)
		}
	}
	composed := shell.ComposeScreen(f.view, f.c.Snapshot())
	count := 0
	for _, component := range composed.Components {
		if !strings.HasPrefix(component.Name, "feature-list-row-") {
			continue
		}
		x, y, width, height := component.Position(1024, 600)
		bodyTop := ui.OuterMargin + ui.HeaderHeight + ui.Gap
		if x < ui.ContentInset || y < bodyTop || width < 48 || height < 48 || x+width > 1024-ui.ContentInset+0.01 || y+height > 600-ui.OuterMargin+0.01 {
			t.Fatalf("large calendar cell target outside body or below 48 px: %s at %.1f,%.1f %.1fx%.1f", component.Name, x, y, width, height)
		}
		count++
	}
	if count != 49 {
		t.Fatalf("calendar cells = %d, want 49 including weekdays", count)
	}
}

func containsRosterName(rows []ui.Item, name string) bool {
	for _, row := range rows {
		if strings.HasPrefix(row.Label, name+" · ") {
			return true
		}
	}
	return false
}
