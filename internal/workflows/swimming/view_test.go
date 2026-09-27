package swimming_test

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"image/png"
	"os"
	"path/filepath"
	"rectablet/internal/shared/ui"
	"rectablet/internal/shell"
	lesson "rectablet/internal/workflows/swimming"
	"rectablet/internal/workflows/swimming/curriculum"
	store "rectablet/internal/workflows/swimming/repository"
	"rectablet/internal/workflows/swimming/transfer"
	"strings"
	"testing"
	"time"
)

type uiFixture struct {
	view *shell.View
	c    *lesson.Controller
	s    *store.Store
	w    fyne.Window
}

func must[T any](v T, e error) T {
	if e != nil {
		panic(e)
	}
	return v
}
func check(t *testing.T, e error) {
	t.Helper()
	if e != nil {
		t.Fatal(e)
	}
}
func setupUI(t *testing.T) *uiFixture {
	t.Helper()
	a := test.NewApp()
	a.Settings().SetTheme(ui.AppTheme{})
	s := must(store.Open(filepath.Join(t.TempDir(), "test.sqlite3")))
	_ = must(s.ImportFile("../../../examples/classes.json"))
	c := lesson.New(s)
	c.Date = "2026-09-04"
	check(t, c.Refresh())
	w := a.NewWindow("Rec Tablet")
	v := shell.New(c, w)
	w.Show()
	f := &uiFixture{v, c, s, w}
	t.Cleanup(func() { v.Close(); w.Close(); a.Quit(); s.Close() })
	return f
}
func (f *uiFixture) tap(t *testing.T, key string) {
	t.Helper()
	b := f.view.Buttons[key]
	if b == nil {
		t.Fatal("missing button", key)
	}
	for _, b := range f.view.Buttons {
		b.CanActivate = func() bool { return true }
	}
	test.Tap(b)
}
func (f *uiFixture) row(t *testing.T, index int) {
	t.Helper()
	f.tap(t, "row:"+f.c.State.Rows[index].ID)
}
func (f *uiFixture) rowLabel(t *testing.T, label string) {
	t.Helper()
	for i, row := range f.c.State.Rows {
		if row.Label == label {
			f.row(t, i)
			return
		}
	}
	t.Fatalf("missing row %q", label)
}
func (f *uiFixture) skill(t *testing.T) {
	f.row(t, 0)
	f.tap(t, "row:skills")
	f.row(t, 0)
	f.rowLabel(t, "Harrold")
}
func (f *uiFixture) shot(t *testing.T, name string) {
	t.Helper()
	dir := os.Getenv("REC_TABLET_SCREENSHOTS")
	if dir == "" {
		return
	}
	check(t, os.MkdirAll(dir, 0755))
	file := must(os.Create(filepath.Join(dir, name+".png")))
	check(t, png.Encode(file, f.w.Canvas().Capture()))
	check(t, file.Close())
}
func TestPointerNavigationAttendanceAndMaintenance(t *testing.T) {
	f := setupUI(t)
	f.shot(t, "classes")
	for _, b := range f.view.Buttons {
		b.CanActivate = func() bool { return true }
	}
	test.TapCanvas(f.w.Canvas(), fyne.NewPos(380, 158))
	if f.c.State.Screen != "class" {
		t.Fatal(f.c.State.Screen)
	}
	for _, b := range f.view.Buttons {
		b.CanActivate = func() bool { return true }
	}
	test.TapCanvas(f.w.Canvas(), fyne.NewPos(380, 158))
	if f.c.State.Screen != "attendance" {
		t.Fatal(f.c.State.Screen)
	}
	f.shot(t, "attendance")
	f.row(t, 0)
	if f.c.State.Rows[0].Status != "present" {
		t.Fatal("attendance did not save")
	}
	f.row(t, 0)
	if f.c.State.Rows[0].Status != "absent" {
		t.Fatal("attendance toggle did not switch to absent")
	}
	f.c.Dismiss()
	f.tap(t, "down")
	if f.c.State.Page != 1 {
		t.Fatal("paging failed")
	}
	f.tap(t, "home")
	if f.c.State.Screen != "classes" {
		t.Fatal("home failed")
	}
	hold := f.view.Buttons["title"]
	original := hold.OnHold
	done := make(chan struct{})
	hold.OnHold = func() { original(); close(done) }
	for _, b := range f.view.Buttons {
		b.CanActivate = func() bool { return true }
	}
	hold.MouseDown(&desktop.MouseEvent{Button: desktop.MouseButtonPrimary})
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("three-second maintenance hold did not fire")
	}
	hold.MouseUp(&desktop.MouseEvent{})
	hold.Tapped(nil)
	if f.c.State.Screen != "maintenance" {
		t.Fatal(f.c.State.Screen)
	}
	f.shot(t, "maintenance")
}
func TestSkillChecklistSaveDiscardAndHold(t *testing.T) {
	f := setupUI(t)
	f.skill(t)
	if f.c.State.Screen != "assessment" || f.c.State.RowCount < 2 {
		t.Fatal("skill checklist did not open")
	}
	firstID := f.c.State.Rows[0].ID
	secondID := f.c.State.Rows[1].ID
	person := f.c.Frame().Person
	version := f.c.Frame().Version
	f.tap(t, "row:"+firstID)
	if f.c.State.Rows[0].Status != "successful" {
		t.Fatal("subskill check not saved")
	}
	f.shot(t, "assessment")
	f.tap(t, "row:"+secondID)
	if f.c.State.Screen != "assessment" {
		t.Fatal("checklist exited before Save was pressed")
	}
	if f.c.State.Notice != "" {
		t.Fatalf("skill completion displayed an unexpected notice: %q", f.c.State.Notice)
	}
	if !f.view.Buttons["home"].IconOnly || f.view.Buttons["home"].Icon == nil || !f.view.Buttons["back"].IconOnly || f.view.Buttons["back"].Icon == nil {
		t.Fatal("checklist did not replace Home and Back with Save and Discard")
	}
	f.tap(t, "home")
	if f.c.State.Screen != "skill_swimmers" {
		t.Fatal("Save attempt did not return to the student list")
	}
	if p := must(f.s.Progress(person, "Splash1:1", version)); p.Successes != 1 {
		t.Fatal("successful check did not increment progress")
	}
	if f.view.Buttons["row:"+person] == nil {
		t.Fatal("student button was not rebuilt after completing the checklist")
	}
	var count int
	check(t, f.s.DB.QueryRow("SELECT COUNT(*) FROM attempts").Scan(&count))
	if count != 1 {
		t.Fatalf("saved attempt count = %d, want 1", count)
	}
	f.rowLabel(t, "Harrold")
	f.tap(t, "row:"+f.c.State.Rows[0].ID)
	f.tap(t, "home")
	if f.c.State.Screen != "skill_swimmers" {
		t.Fatal("Save incomplete attempt did not return to the student list")
	}
	check(t, f.s.DB.QueryRow("SELECT COUNT(*) FROM attempts").Scan(&count))
	if count != 2 {
		t.Fatalf("saved incomplete attempt count = %d, want 2", count)
	}
	if p := must(f.s.Progress(person, "Splash1:1", version)); p.Successes != 1 {
		t.Fatal("incomplete attempt incorrectly changed successful progress")
	}
	f.rowLabel(t, "Harrold")
	f.tap(t, "row:"+f.c.State.Rows[0].ID)
	f.tap(t, "back")
	if f.c.State.Screen != "skill_swimmers" {
		t.Fatal("Discard attempt did not return to the student list")
	}
	check(t, f.s.DB.QueryRow("SELECT COUNT(*) FROM attempts").Scan(&count))
	if count != 2 {
		t.Fatalf("discarded unfinished attempt changed saved attempt count to %d", count)
	}
	for _, row := range f.c.State.Rows {
		if row.ID == person {
			if row.ProgressSegments != 1 {
				t.Fatalf("student button did not show one segment: %d", row.ProgressSegments)
			}
		}
	}

	skillButton := f.view.Buttons["row:"+person]
	if skillButton == nil || skillButton.OnHold == nil || skillButton.HoldDuration != 2*time.Second || skillButton.ProgressSegments != 1 {
		t.Fatal("skill hold action is not configured for two seconds")
	}
	for _, b := range f.view.Buttons {
		b.CanActivate = func() bool { return true }
	}
	originalHold := skillButton.OnHold
	holdComplete := make(chan struct{})
	skillButton.OnHold = func() {
		originalHold()
		close(holdComplete)
	}
	skillButton.MouseDown(&desktop.MouseEvent{Button: desktop.MouseButtonPrimary})
	select {
	case <-holdComplete:
	case <-time.After(5 * time.Second):
		t.Fatal("two-second skill hold did not fire")
	}
	skillButton.MouseUp(&desktop.MouseEvent{})
	if p := must(f.s.Progress(person, "Splash1:1", version)); p.Successes != 0 {
		t.Fatal("holding a skill did not remove its latest check")
	}
	check(t, f.s.DB.QueryRow("SELECT COUNT(*) FROM attempts").Scan(&count))
	if count != 1 {
		t.Fatal("removed check left more than the saved incomplete attempt")
	}
}

func TestSkillsViewUsesCompactSkillTitles(t *testing.T) {
	f := setupUI(t)
	f.row(t, 0)
	f.tap(t, "row:skills")
	const skillID = "Splash1:10-front-crawl-wearing-pfd-5-m"
	for {
		for _, row := range f.c.State.Rows {
			if row.ID != skillID {
				continue
			}
			skill := must(curriculum.Bundled().Skill(skillID))
			want := skill.Number + ". " + curriculum.CompactTitle(skill)
			if row.Label != want {
				t.Fatalf("Skills view label = %q, want compact title %q", row.Label, want)
			}
			if row.Label == skill.Number+". "+skill.Name {
				t.Fatal("Skills view still uses the full curriculum title")
			}
			return
		}
		if f.view.Buttons["down"].Disabled {
			t.Fatalf("skill %s was not found", skillID)
		}
		f.tap(t, "down")
	}
}

func TestTimedHeatSelectionTimerAndSave(t *testing.T) {
	f := setupUI(t)
	p := must(transfer.ReadPackage("../../../examples/classes.json"))
	level := "Splash7"
	for i := range p.Courses[0].Swimmers[:2] {
		p.Courses[0].Swimmers[i].Level = &level
	}
	_ = must(f.s.Import(p))
	check(t, f.c.Refresh())
	f.row(t, 0)
	f.tap(t, "row:skills")
	for {
		found := false
		for _, row := range f.c.State.Rows {
			if row.ID == "Splash7:10" {
				found = true
				break
			}
		}
		if found {
			break
		}
		f.tap(t, "down")
	}
	f.tap(t, "row:Splash7:10")
	if f.c.State.Screen != "skill_swimmers" || len(f.c.State.Rows) != 2 {
		t.Fatalf("timed skill did not show its selected roster: screen=%s rows=%d", f.c.State.Screen, len(f.c.State.Rows))
	}
	if f.view.Buttons["header-switch-0"] != nil || f.view.Buttons["header-switch-1"] != nil {
		t.Fatal("timed benchmark exposed the normal Single/Multi switch")
	}
	people := []string{f.c.State.Rows[0].ID, f.c.State.Rows[1].ID}
	f.tap(t, "row:"+people[0])
	f.tap(t, "row:"+people[1])
	if !f.view.Buttons["home"].IconOnly || f.view.Buttons["home"].Icon == nil || !f.view.Buttons["back"].IconOnly || f.view.Buttons["back"].Icon == nil {
		t.Fatal("timed swimmer selection did not show heat actions")
	}
	f.tap(t, "home")
	if f.c.State.Screen != "timed_laps" {
		t.Fatal("timed selection did not open lap setup")
	}
	f.tap(t, "timed-digit:3")
	f.compactLayout(t, "timed-laps")
	f.tap(t, "home")
	if f.c.State.Screen != "timed_heat" || f.view.Buttons["header-action-0"] == nil || f.view.Buttons["header-action-1"] == nil {
		t.Fatal("lap setup did not open the timed heat")
	}
	f.compactLayout(t, "timed-heat")
	f.tap(t, "header-action-0")
	for i := 0; i < 3; i++ {
		f.tap(t, "row:timed-lap:"+people[0])
	}
	f.compactLayout(t, "timed-heat-completed")
	if f.c.State.Screen != "timed_heat" || f.c.State.Rows[0].ProgressValue != 3 || f.c.State.Rows[0].ProgressTotal != 3 {
		t.Fatal("lap taps did not complete the swimmer")
	}
	f.tap(t, "home")
	if f.c.State.Screen != "skill_swimmers" {
		t.Fatal("saving timed heat did not return to swimmers")
	}
	var attempts, timed int
	check(t, f.s.DB.QueryRow("SELECT COUNT(*) FROM attempts WHERE skill_id='Splash7:10'").Scan(&attempts))
	check(t, f.s.DB.QueryRow("SELECT COUNT(*) FROM timed_attempts").Scan(&timed))
	if attempts != 2 || timed != 2 {
		t.Fatalf("timed heat was not persisted: attempts=%d timed=%d", attempts, timed)
	}
	if progress := must(f.s.Progress(people[0], "Splash7:10", f.c.Frame().Version)); progress.Successes != 1 {
		t.Fatalf("completed timed swimmer did not pass: %+v", progress)
	}
}

func TestMissingCurriculumAndEmptyClass(t *testing.T) {
	f := setupUI(t)
	p := must(transfer.ReadPackage("../../../examples/classes.json"))
	level := "Unknown Level"
	for i := range p.Courses[0].Swimmers {
		p.Courses[0].Swimmers[i].Level = &level
	}
	p.Courses[1].Swimmers = []transfer.SwimmerInput{}
	_ = must(f.s.Import(p))
	check(t, f.c.Refresh())
	f.row(t, 0)
	f.tap(t, "row:skills")
	if f.c.State.RowCount != 1 || !f.c.State.Rows[0].Disabled || f.c.State.Rows[0].Label != "Unknown Level · Missing curriculum · update swimmer level" {
		t.Fatalf("missing curriculum state absent: screen=%s rows=%d empty=%q", f.c.State.Screen, f.c.State.RowCount, f.c.State.Empty)
	}
	f.tap(t, "home")
	f.row(t, 1)
	f.tap(t, "row:attendance")
	if f.c.State.RowCount != 0 {
		t.Fatal("empty roster not empty")
	}
}
func TestPrivateSwimmersWithoutLevelsStayOutOfSkillAssessment(t *testing.T) {
	f := setupUI(t)
	p := must(transfer.ReadPackage("../../../examples/classes.json"))
	for i := range p.Courses[2].Swimmers {
		p.Courses[2].Swimmers[i].Level = nil
	}
	_ = must(f.s.Import(p))
	check(t, f.c.Refresh())
	f.row(t, 2)
	f.tap(t, "row:skills")
	if f.c.State.RowCount != 0 || f.c.State.Empty != "No enrolled swimmers with an assigned level." {
		t.Fatalf("unassigned private swimmers appeared in skills: rows=%d empty=%q", f.c.State.RowCount, f.c.State.Empty)
	}
	f.tap(t, "back")
	if f.c.State.Screen != "class" {
		t.Fatalf("back from skills opened %q, want class", f.c.State.Screen)
	}
	f.tap(t, "row:attendance")
	if f.c.State.RowCount != 2 {
		t.Fatalf("private attendance roster has %d swimmers, want 2", f.c.State.RowCount)
	}
}
func TestSkillActionFailuresAreQuiet(t *testing.T) {
	f := setupUI(t)
	f.skill(t)
	_, e := f.s.DB.Exec(`CREATE TEMP TRIGGER reject_draft BEFORE UPDATE ON drafts BEGIN SELECT RAISE(ABORT,'Disk full'); END`)
	check(t, e)
	f.tap(t, "row:"+f.c.State.Rows[0].ID)
	if f.c.State.Error || f.c.State.Rows[0].Status != "unmarked" {
		t.Fatal("failed save was not handled quietly")
	}
	f.tap(t, "down")
	if f.c.State.Error {
		t.Fatal("quiet skill failure set an error state")
	}
	f.shot(t, "save-error")
}
func TestLongCriteriaAndLargerLandscape(t *testing.T) {
	f := setupUI(t)
	f.row(t, 2)
	f.tap(t, "row:skills")
	for {
		found := false
		for _, r := range f.c.State.Rows {
			if r.ID == "Splash9:7" {
				found = true
			}
		}
		if found {
			break
		}
		f.tap(t, "down")
	}
	f.tap(t, "row:Splash9:7")
	f.rowLabel(t, "Jordan")
	f.shot(t, "long-skill")
	f.w.Resize(fyne.NewSize(1280, 800))
	f.shot(t, "assessment-1280")
	if f.c.State.RowCount == 0 {
		t.Fatal("long skill has no subskills")
	}
}
func TestNavigationGuard(t *testing.T) {
	f := setupUI(t)
	f.row(t, 0)
	test.Tap(f.view.Buttons["row:attendance"])
	if f.c.State.Screen != "class" {
		t.Fatal("rapid tap crossed navigation boundary")
	}
	f.tap(t, "row:attendance")
	if f.c.State.Screen != "attendance" {
		t.Fatal("navigation did not unlock")
	}
}
func TestLongestCriterionFitsAndGroupsRemainVisible(t *testing.T) {
	f := setupUI(t)
	p := must(transfer.ReadPackage("../../../examples/classes.json"))
	level := "SplashFitness"
	p.Courses[0].Swimmers[0].Level = &level
	_ = must(f.s.Import(p))
	check(t, f.c.Refresh())
	f.row(t, 0)
	f.tap(t, "row:skills")
	for {
		found := false
		for _, r := range f.c.State.Rows {
			if r.ID == "SplashFitness:3" {
				found = true
			}
		}
		if found {
			break
		}
		f.tap(t, "down")
	}
	f.tap(t, "row:SplashFitness:3")
	f.rowLabel(t, "Harrold")
	_ = f.w.Canvas().Capture()
	if f.c.State.RowCount == 0 {
		t.Fatal("longest skill has no subskills")
	}
	f.shot(t, "longest-criterion")
}
func TestFailedImportAndExportSurfaceErrors(t *testing.T) {
	f := setupUI(t)
	before := f.c.State.RowCount
	path := filepath.Join(t.TempDir(), "bad.json")
	check(t, os.WriteFile(path, []byte("broken"), 0600))
	if f.c.Import(path) == nil || !f.c.State.Error {
		t.Fatal("invalid import hidden")
	}
	if f.c.State.RowCount != before {
		t.Fatal("failed import changed view")
	}
	f.c.Dismiss()
	if f.c.Export(filepath.Join(t.TempDir(), "missing", "out.json")) == nil {
		t.Fatal("export failure hidden")
	}
	f.c.Fail(errors.New("test"))
	if !f.c.State.Error {
		t.Fatal("error state missing")
	}
}

// Check the actual rendered objects as well as the feature placement contracts.
func (f *uiFixture) compactLayout(t *testing.T, name string) {
	t.Helper()
	original := f.w.Canvas().Size()
	defer f.w.Resize(original)
	for _, size := range []fyne.Size{fyne.NewSize(800, 480), fyne.NewSize(1200, 720), fyne.NewSize(1280, 800)} {
		f.w.Resize(size)
		f.w.Canvas().Capture()
		scale := min(size.Width/800, size.Height/480)
		composition := shell.ComposeScreen(f.view, f.c.Snapshot())
		for _, component := range composition.Components {
			x, y, width, height := component.Position(size.Width/scale, size.Height/scale)
			if x < 0 || y < 0 || width <= 0 || height <= 0 || x+width > size.Width/scale+0.01 || y+height > size.Height/scale+0.01 {
				t.Fatalf("%s outside %v: %s at %v", name, size, component.Name, []float32{x, y, width, height})
			}
			if strings.HasPrefix(component.Name, "feature-") && (x < ui.ContentInset || x+width > size.Width/scale-ui.ContentInset+0.01) {
				t.Fatalf("%s overlaps a rail: %s", name, component.Name)
			}
		}
		for id, button := range f.view.Buttons {
			for _, object := range test.WidgetRenderer(button).Objects() {
				if !object.Visible() {
					continue
				}
				if _, ok := object.(*canvas.Image); ok {
					position, bounds := object.Position(), object.Size()
					if position.X < 0 || position.Y < 0 || position.X+bounds.Width > button.Size().Width+0.01 || position.Y+bounds.Height > button.Size().Height+0.01 {
						t.Fatalf("%s icon escapes %s at %v", name, id, size)
					}
				}
			}
		}
		f.shot(t, fmt.Sprintf("%s-%dx%d", name, int(size.Width), int(size.Height)))
	}
}

func TestCompactFeedbackCorrectionAndDateLayouts(t *testing.T) {
	f := setupUI(t)
	f.compactLayout(t, "classes")
	f.tap(t, "header")
	selected := f.view.Buttons["row:2026-09-04"]
	if selected == nil || !selected.Selected {
		t.Fatal("selected date should remain visible")
	}
	f.compactLayout(t, "dates")
	f.tap(t, "row:2026-09-04")
	f.row(t, 0)
	f.compactLayout(t, "lesson-menu")
	f.tap(t, "row:skills")
	f.row(t, 0)
	f.rowLabel(t, "Harrold")
	f.row(t, 0)
	f.tap(t, "home") // Save an incomplete attempt for feedback and correction.
	f.tap(t, "back")
	f.tap(t, "back")
	f.tap(t, "row:feedback_swimmers")
	f.compactLayout(t, "feedback-swimmers")
	f.rowLabel(t, "Harrold")
	f.compactLayout(t, "feedback-skills")
	f.row(t, 0)
	f.compactLayout(t, "feedback-detail")
	f.tap(t, "header")
	f.compactLayout(t, "attempts")
	f.row(t, 0)
	if f.view.Buttons["submit"] == nil {
		t.Fatal("correction action missing")
	}
	f.compactLayout(t, "correction")
	f.tap(t, "successful")
	f.tap(t, "down")
	f.tap(t, "not_yet")
	f.tap(t, "submit")
	if f.c.State.Screen != "attempts" {
		t.Fatal("correction did not return to attempts")
	}
}
