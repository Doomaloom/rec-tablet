package swimming_test

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"image/png"
	"os"
	"path/filepath"
	"rectablet/internal/shared/ui"
	"rectablet/internal/shell"
	lesson "rectablet/internal/workflows/swimming"
	store "rectablet/internal/workflows/swimming/repository"
	"rectablet/internal/workflows/swimming/transfer"
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
	f.row(t, 0)
	if f.c.State.Rows[0].Status != "present" {
		t.Fatal("attendance did not save")
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
func TestSkillChecklistAutoCompletesAndHoldRemovesCheck(t *testing.T) {
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
	if f.c.State.Screen != "skill_swimmers" {
		t.Fatal("complete checklist did not close")
	}
	if f.c.State.Notice != "" {
		t.Fatalf("skill completion displayed an unexpected notice: %q", f.c.State.Notice)
	}
	if p := must(f.s.Progress(person, "Splash1:1", version)); p.Successes != 1 {
		t.Fatal("successful check did not increment progress")
	}
	if f.view.Buttons["row:"+person] == nil {
		t.Fatal("student button was not rebuilt after completing the checklist")
	}
	for _, row := range f.c.State.Rows {
		if row.ID == person {
			if row.Detail != "" {
				t.Fatalf("student progress subtext was not removed: %q", row.Detail)
			}
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
	var count int
	check(t, f.s.DB.QueryRow("SELECT COUNT(*) FROM attempts").Scan(&count))
	if count != 0 {
		t.Fatal("removed check left its attempt behind")
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
	if f.c.State.RowCount != 1 || !f.c.State.Rows[0].Disabled || f.c.State.Rows[0].Detail != "Missing curriculum · update swimmer level" {
		t.Fatalf("missing curriculum state absent: screen=%s rows=%d empty=%q", f.c.State.Screen, f.c.State.RowCount, f.c.State.Empty)
	}
	f.tap(t, "home")
	f.row(t, 1)
	f.tap(t, "row:attendance")
	if f.c.State.RowCount != 0 {
		t.Fatal("empty roster not empty")
	}
}
func TestSaveFailureKeepsDraftAndVisibleError(t *testing.T) {
	f := setupUI(t)
	f.skill(t)
	_, e := f.s.DB.Exec(`CREATE TEMP TRIGGER reject_draft BEFORE UPDATE ON drafts BEGIN SELECT RAISE(ABORT,'Disk full'); END`)
	check(t, e)
	f.tap(t, "row:"+f.c.State.Rows[0].ID)
	if !f.c.State.Error || f.c.State.Rows[0].Status != "unmarked" {
		t.Fatal("failed save falsely acknowledged")
	}
	f.tap(t, "down")
	if !f.c.State.Error {
		t.Fatal("error dismissed by navigation")
	}
	f.shot(t, "save-error")
	f.tap(t, "dismiss")
	if f.c.State.Error {
		t.Fatal("dismiss did not clear error")
	}
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
