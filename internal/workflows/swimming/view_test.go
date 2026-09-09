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
func (f *uiFixture) skill(t *testing.T) {
	f.row(t, 0)
	f.tap(t, "row:skills")
	f.row(t, 0)
	f.row(t, 0)
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
func TestAssessmentDraftResumeFeedbackAndDoubleSubmit(t *testing.T) {
	f := setupUI(t)
	f.skill(t)
	f.tap(t, "header")
	a := f.c.Frame().Attempt
	if f.c.State.CanSubmit {
		t.Fatal("incomplete submit enabled")
	}
	f.tap(t, "successful")
	f.c.Dismiss()
	if f.c.State.Rows[0].Status != "successful" {
		t.Fatal("mark not saved")
	}
	f.shot(t, "assessment")
	f.tap(t, "home")
	f.skill(t)
	f.tap(t, "header")
	if f.c.Frame().Attempt != a || f.c.State.Rows[0].Status != "successful" {
		t.Fatal("draft not resumed")
	}
	f.tap(t, "down")
	f.tap(t, "not_yet")
	f.c.Dismiss()
	if !f.c.State.CanSubmit {
		t.Fatal("submit not enabled")
	}
	f.tap(t, "submit")
	if f.c.State.Screen != "attempts" {
		t.Fatal("submit did not navigate")
	}
	check(t, f.c.Submit())
	var count int
	check(t, f.s.DB.QueryRow("SELECT COUNT(*) FROM attempts").Scan(&count))
	if count != 1 {
		t.Fatal("double submit duplicated")
	}
	person := f.c.Frame().Person
	f.c.Dismiss()
	f.tap(t, "home")
	f.row(t, 0)
	f.tap(t, "row:feedback_swimmers")
	f.tap(t, "row:"+person)
	f.row(t, 0)
	if f.c.State.RowCount != 1 {
		t.Fatal("feedback missing failed subskill")
	}
	f.shot(t, "feedback")
}
func TestMissingCurriculumAndEmptyClass(t *testing.T) {
	f := setupUI(t)
	p := must(transfer.ReadPackage("../../../examples/classes.json"))
	level := "Unknown Level"
	p.Courses[0].Swimmers[0].Level = &level
	p.Courses[1].Swimmers = []transfer.SwimmerInput{}
	_ = must(f.s.Import(p))
	check(t, f.c.Refresh())
	f.row(t, 0)
	f.tap(t, "row:skills")
	for f.c.State.Page+1 < f.c.State.PageCount {
		f.tap(t, "down")
	}
	found := false
	for _, r := range f.c.State.Rows {
		if r.Disabled {
			found = true
		}
	}
	if !found {
		t.Fatal("missing curriculum state absent")
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
	f.tap(t, "header")
	_, e := f.s.DB.Exec(`CREATE TEMP TRIGGER reject_draft BEFORE UPDATE ON drafts BEGIN SELECT RAISE(ABORT,'Disk full'); END`)
	check(t, e)
	f.tap(t, "successful")
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
	f.row(t, 0)
	f.tap(t, "header")
	f.shot(t, "long-skill")
	f.w.Resize(fyne.NewSize(1280, 800))
	f.shot(t, "assessment-1280")
	if f.view.Criterion.Truncated {
		t.Fatal("criterion clipped")
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
	f.row(t, 0)
	f.tap(t, "header")
	f.tap(t, "down")
	_ = f.w.Canvas().Capture()
	if f.view.Criterion.Truncated {
		t.Fatal("longest criterion clipped")
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
