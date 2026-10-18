package assessment

import (
	"errors"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"rectablet/src/workflows/swimming/assessment/view/large"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
	"rectablet/src/workflows/swimming/navigation"
)

type largeAssessmentStore struct {
	Repository
	catalog                  curriculum.Catalog
	roster                   []domain.Enrollment
	failProgress, failSubmit error
	attempts                 map[string]string
	answers                  map[string]map[string]string
	submitted                [][]string
	courseLevel              string
}

func (s *largeAssessmentStore) GetCourse(string) (domain.Course, error) {
	return domain.Course{Name: "Swimming", Time: "9:00", Level: s.courseLevel, CurriculumVersion: s.catalog.Version}, nil
}
func (s *largeAssessmentStore) ListRoster(string) ([]domain.Enrollment, error) { return s.roster, nil }
func (s *largeAssessmentStore) CatalogFor(string) (curriculum.Catalog, error)  { return s.catalog, nil }
func (s *largeAssessmentStore) Skill(_, id string) (curriculum.Skill, error) {
	return s.catalog.Skill(id)
}
func (s *largeAssessmentStore) Progress(string, string, string) (domain.Progress, error) {
	return domain.Progress{}, s.failProgress
}
func (s *largeAssessmentStore) Draft(attempt string) (map[string]string, error) {
	return s.answers[attempt], nil
}
func (s *largeAssessmentStore) SubmitAttempts(ids []string) error {
	if s.failSubmit != nil {
		return s.failSubmit
	}
	s.submitted = append(s.submitted, append([]string(nil), ids...))
	return nil
}
func (s *largeAssessmentStore) GetEnrollment(id string) (domain.Enrollment, error) {
	for _, student := range s.roster {
		if student.ID == id {
			return student, nil
		}
	}
	return domain.Enrollment{}, errors.New("missing student")
}
func (s *largeAssessmentStore) GetAttempt(id string) (domain.Attempt, error) {
	for person, attempt := range s.attempts {
		if attempt == id {
			return domain.Attempt{ID: id, EnrollmentID: person, SkillID: s.catalog.Levels[0].Skills[0].ID, CurriculumVersion: s.catalog.Version}, nil
		}
	}
	return domain.Attempt{}, errors.New("missing attempt")
}
func (s *largeAssessmentStore) NewAttempt(person, _, _ string) (string, error) {
	if s.attempts == nil {
		s.attempts = make(map[string]string)
	}
	if s.answers == nil {
		s.answers = make(map[string]map[string]string)
	}
	if s.attempts[person] == "" {
		s.attempts[person] = "draft-" + person
		s.answers[s.attempts[person]] = make(map[string]string)
	}
	return s.attempts[person], nil
}
func (s *largeAssessmentStore) Mark(attempt, criterion, status string) error {
	if s.answers[attempt] == nil {
		s.answers[attempt] = make(map[string]string)
	}
	s.answers[attempt][criterion] = status
	return nil
}

type largePresenter struct{ state shell.State }

func TestLargeReturnToSwimmersPagesWholeRoster(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("returned swimmer list")
	defer window.Close()
	c := &Controller{}
	state := shell.State{DisplayMode: ui.DisplayLarge, Screen: "skill_swimmers", PageCount: 1}
	for i := 0; i < 13; i++ {
		state.Rows = append(state.Rows, ui.Item{ID: fmt.Sprint(i), Label: fmt.Sprintf("Swimmer %d", i)})
	}
	state.RowCount = len(state.Rows)
	state.ComposeBody = c.composeSwimmerListBody(state)
	p := &largePresenter{state: state}
	view := shell.New(p, window)
	defer view.Close()
	if view.Buttons["row:11"] == nil || view.Buttons["row:12"] != nil || view.Buttons["large-return-swimmers-page-1"] == nil {
		t.Fatal("returned roster must show twelve swimmers and offer local paging")
	}
	c.largeRosterPage = 1
	view.Render()
	if view.Buttons["row:12"] == nil || view.Buttons["row:0"] != nil {
		t.Fatal("last swimmer is not reachable on the second page")
	}
}

func (p *largePresenter) Snapshot() shell.State { return p.state }
func (p *largePresenter) Subscribe(func())      {}
func (p *largePresenter) Home() error           { return nil }
func (p *largePresenter) Back() error           { return nil }
func (p *largePresenter) Page(int) error        { return nil }
func (p *largePresenter) Header() error         { return nil }
func (p *largePresenter) Submit() error         { return nil }
func (p *largePresenter) Dismiss()              {}
func (p *largePresenter) OpenSetup() error      { return nil }

func TestLargeChecklistFitsEveryCatalogSkillWithSidebar(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	app.Settings().SetTheme(ui.AppTheme{})
	window := app.NewWindow("large checklist")
	defer window.Close()
	controller := &Controller{}
	presenter := &largePresenter{state: shell.State{DisplayMode: ui.DisplayLarge}}
	view := shell.New(presenter, window)
	defer view.Close()
	for _, level := range curriculum.Bundled().Levels {
		for _, skill := range level.Skills {
			rows := make([]navigation.Row, 0, len(skill.Subskills))
			for _, subskill := range skill.Subskills {
				rows = append(rows, navigation.Row{ID: subskill.ID, Label: subskill.Text})
			}
			state := shell.State{
				DisplayMode: ui.DisplayLarge, Screen: "large-checklist", HideFooter: true, PageCount: 4,
				SplitHeader: &shell.SplitHeader{Left: skill.Number + ". " + skill.Name, Right: "A Long Swimmer Name"},
				PrimaryInfo: "Saturday morning swimming group · 9:00 AM",
			}
			configureLargeClassHeader(&state, skill, domain.Course{Name: "Saturday morning swimming group", Time: "9:00 AM"})
			state.ComposeBody = func(surface *ui.Surface) {
				_, right := large.Split(surface, .28)
				controller.largeView().RenderLargeChecklist(right, rows, nil)
			}
			presenter.state = state
			composition := shell.ComposeScreen(view, state)
			seen := 0
			for _, component := range composition.Components {
				if component.Name == "header-half-0" {
					header := component.Object.(*ui.TextBlock)
					_, _, width, height := component.Position(1024, 600)
					header.Resize(fyne.NewSize(width, height))
					header.Refresh()
					if header.Truncated {
						t.Fatalf("%s: skill header clips", skill.ID)
					}
				}
				if component.Name == "primary-info" {
					info := component.Object.(*ui.TextBlock)
					_, _, width, height := component.Position(1024, 600)
					info.Resize(fyne.NewSize(width, height))
					info.Refresh()
					if info.Truncated {
						t.Fatalf("%s: full skill information clips", skill.ID)
					}
				}
				if !strings.HasPrefix(component.Name, "feature-large-criterion-") {
					continue
				}
				button, ok := component.Object.(*ui.TouchButton)
				if !ok {
					t.Fatalf("%s: criterion is not a button", skill.ID)
				}
				_, _, width, height := component.Position(1024, 600)
				if height < 48 {
					t.Fatalf("%s: target height %.1f < 48", skill.ID, height)
				}
				if button.FontSize < 18 {
					t.Fatalf("%s: font %.1f < 18", skill.ID, button.FontSize)
				}
				needed := ui.WrappedTextHeight(button.Text, width-2*ui.ButtonPadding, button.FontSize, true)
				if needed+4 > height {
					t.Fatalf("%s: criterion %q clips: %.1f > %.1f", skill.ID, button.Text, needed+4, height)
				}
				seen++
			}
			if seen != len(rows) {
				t.Fatalf("%s: rendered %d of %d criteria", skill.ID, seen, len(rows))
			}
			if skill.ID == "Splash8:14" {
				if dir := os.Getenv("REC_TABLET_SCREENSHOTS"); dir != "" {
					if err := os.MkdirAll(dir, 0755); err != nil {
						t.Fatal(err)
					}
					view.Render()
					window.Show()
					window.Resize(fyne.NewSize(1024, 600))
					file, err := os.Create(filepath.Join(dir, "large-class-checklist-12.png"))
					if err != nil {
						t.Fatal(err)
					}
					if err := png.Encode(file, window.Canvas().Capture()); err != nil {
						t.Fatal(err)
					}
					if err := file.Close(); err != nil {
						t.Fatal(err)
					}
				}
			}
		}
	}
}

func TestLargeSingleAndCorrectionShowFullTitlesAndAllCriteria(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	app.Settings().SetTheme(ui.AppTheme{})
	window := app.NewWindow("large single")
	defer window.Close()
	presenter := &largePresenter{state: shell.State{DisplayMode: ui.DisplayLarge}}
	view := shell.New(presenter, window)
	defer view.Close()
	c := &Controller{}
	for _, level := range curriculum.Bundled().Levels {
		for _, skill := range level.Skills {
			rows := make([]navigation.Row, 0, len(skill.Subskills))
			for _, criterion := range skill.Subskills {
				rows = append(rows, navigation.Row{ID: criterion.ID, Label: criterion.Text})
			}
			for _, correction := range []bool{false, true} {
				state := shell.State{DisplayMode: ui.DisplayLarge, Screen: "assessment", SplitHeader: &shell.SplitHeader{Left: "A Long Swimmer Name", Right: skill.Name}}
				configureLargeAssessmentHeader(&state, skill)
				if correction {
					state.FooterAction = "Save correction"
				}
				state.ComposeBody = func(surface *ui.Surface) {
					if correction {
						c.largeView().RenderLargeChecklist(surface, rows, nil)
						return
					}
					_, right := large.Split(surface, .28)
					c.largeView().RenderLargeChecklist(right, rows, nil)
				}
				presenter.state = state
				composition := shell.ComposeScreen(view, state)
				if state.SplitHeader.Right != skill.Name && !strings.Contains(state.PrimaryInfo, skill.Name) {
					t.Fatalf("%s: complete skill name is missing", skill.ID)
				}
				seen := 0
				for _, component := range composition.Components {
					if component.Name == "header-half-1" || component.Name == "primary-info" {
						text := component.Object.(*ui.TextBlock)
						_, _, width, height := component.Position(1024, 600)
						text.Resize(fyne.NewSize(width, height))
						text.Refresh()
						if text.Truncated {
							t.Fatalf("%s correction=%v: %s clips", skill.ID, correction, component.Name)
						}
					}
					if !strings.HasPrefix(component.Name, "feature-large-criterion-") {
						continue
					}
					button := component.Object.(*ui.TouchButton)
					_, _, width, height := component.Position(1024, 600)
					if height < 48 || button.FontSize < 18 {
						t.Fatalf("%s correction=%v: criterion size %.1f font %.1f", skill.ID, correction, height, button.FontSize)
					}
					if ui.WrappedTextHeight(button.Text, width-2*ui.ButtonPadding, button.FontSize, true)+4 > height {
						t.Fatalf("%s correction=%v: criterion clips", skill.ID, correction)
					}
					seen++
				}
				if seen != len(rows) {
					t.Fatalf("%s correction=%v: rendered %d/%d criteria", skill.ID, correction, seen, len(rows))
				}
				if skill.ID == "Splash6:12" && !correction {
					if dir := os.Getenv("REC_TABLET_SCREENSHOTS"); dir != "" {
						if err := os.MkdirAll(dir, 0755); err != nil {
							t.Fatal(err)
						}
						view.Render()
						window.Show()
						window.Resize(fyne.NewSize(1024, 600))
						file, err := os.Create(filepath.Join(dir, "large-single-long-title.png"))
						if err != nil {
							t.Fatal(err)
						}
						if err := png.Encode(file, window.Canvas().Capture()); err != nil {
							t.Fatal(err)
						}
						if err := file.Close(); err != nil {
							t.Fatal(err)
						}
					}
				}
			}
		}
	}
}

func TestLargeClassRosterShowsSavedDiscardedAndPendingSessions(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("large class")
	defer window.Close()
	people := []domain.MultiAttempt{
		{PersonID: "a", Name: "Alex"}, {PersonID: "b", Name: "Blair"}, {PersonID: "c", Name: "Casey"},
	}
	nav := navigation.New()
	nav.Push(navigation.ClassAssessment, nil)
	c := &Controller{Controller: nav, classRoster: people, classAttempts: people[1:2], classStatus: map[string]string{"a": "Saved", "c": "Discarded"}}
	rows := []navigation.Row{{ID: "criterion", Label: "Can swim"}}
	state := shell.State{DisplayMode: ui.DisplayLarge, Screen: "class", HideFooter: true, ComposeBody: c.largeView().ComposeLargeClassBody(rows)}
	presenter := &largePresenter{state: state}
	view := shell.New(presenter, window)
	defer view.Close()
	composition := shell.ComposeScreen(view, state)
	want := map[string]struct {
		label    string
		disabled bool
	}{
		"feature-list-row-a": {"Alex · Saved", true},
		"feature-list-row-b": {"Blair · Pending", false},
		"feature-list-row-c": {"Casey · Discarded", true},
	}
	for _, component := range composition.Components {
		expected, ok := want[component.Name]
		if !ok {
			continue
		}
		button, ok := component.Object.(*ui.TouchButton)
		if !ok || button.Text != expected.label || button.Disabled != expected.disabled {
			t.Fatalf("%s: wrong status button: %#v", component.Name, button)
		}
		delete(want, component.Name)
	}
	if len(want) > 0 {
		t.Fatalf("missing class roster entries: %v", want)
	}
	if len(c.classAttempts) != 1 || c.classAttempts[0].PersonID != "b" {
		t.Fatal("rendering changed the pending batch")
	}
}

func TestLargeSkillsReturnsRosterProgressFailure(t *testing.T) {
	catalog := curriculum.Bundled()
	level := catalog.Levels[0]
	for len(level.Skills) == 0 {
		level = catalog.Levels[1]
	}
	want := errors.New("progress unavailable")
	store := &largeAssessmentStore{catalog: catalog, roster: []domain.Enrollment{{ID: "a", Name: "Alex", Level: level.ID, Active: true}}, failProgress: want}
	nav := navigation.New()
	nav.DisplayMode = ui.DisplayLarge
	nav.Push(navigation.Skills, func(f *navigation.Frame) { f.Course = "course" })
	c := New(nav, store)
	_, err := c.Build()
	if !errors.Is(err, want) {
		t.Fatalf("Build error = %v, want %v", err, want)
	}
}

func TestLargeClassSaveKeepsSessionStatusAndPendingBatch(t *testing.T) {
	catalog := curriculum.Bundled()
	var skill curriculum.Skill
	for _, level := range catalog.Levels {
		if len(level.Skills) > 0 {
			skill = level.Skills[0]
			break
		}
	}
	people := []domain.MultiAttempt{{AttemptID: "attempt-a", PersonID: "a", Name: "Alex"}, {AttemptID: "attempt-b", PersonID: "b", Name: "Blair"}}
	store := &largeAssessmentStore{catalog: catalog}
	nav := navigation.New()
	nav.DisplayMode = ui.DisplayLarge
	nav.Push(navigation.ClassAssessment, func(f *navigation.Frame) { f.Course, f.Skill, f.Version = "course", skill.ID, catalog.Version })
	c := New(nav, store)
	nav.Features[navigation.ClassAssessment] = c
	c.classAttempts = append([]domain.MultiAttempt(nil), people...)
	c.classRoster = append([]domain.MultiAttempt(nil), people...)
	c.classStatus = make(map[string]string)
	store.failSubmit = errors.New("write failed")
	if err := c.SaveClassStudent(); !errors.Is(err, store.failSubmit) {
		t.Fatalf("failed save error = %v", err)
	}
	if len(c.classAttempts) != 2 || len(c.classStatus) != 0 {
		t.Fatal("failed save changed the pending batch")
	}
	store.failSubmit = nil
	if err := c.SaveClassStudent(); err != nil {
		t.Fatal(err)
	}
	if len(c.classAttempts) != 1 || c.classAttempts[0].PersonID != "b" {
		t.Fatalf("pending batch = %+v", c.classAttempts)
	}
	if c.classStatus["a"] != "Saved" || c.classStatus["b"] != "" {
		t.Fatalf("session statuses = %+v", c.classStatus)
	}
	if nav.Frame().Page != 0 {
		t.Fatalf("current pending page = %d", nav.Frame().Page)
	}
}

func TestLargeSingleSwitchPreservesEachSwimmerDraft(t *testing.T) {
	catalog := curriculum.Bundled()
	level, skill := catalog.Levels[0], catalog.Levels[0].Skills[0]
	store := &largeAssessmentStore{catalog: catalog, roster: []domain.Enrollment{
		{ID: "a", Name: "Alex", Level: level.ID, Active: true},
		{ID: "b", Name: "Blair", Level: level.ID, Active: true},
	}}
	first, err := store.NewAttempt("a", "lesson", skill.ID)
	if err != nil {
		t.Fatal(err)
	}
	nav := navigation.New()
	nav.DisplayMode = ui.DisplayLarge
	nav.Push(navigation.SkillSwimmers, func(f *navigation.Frame) {
		f.Course, f.Lesson, f.Level, f.Skill, f.Version = "course", "lesson", level.ID, skill.ID, catalog.Version
	})
	nav.Push(navigation.Assessment, func(f *navigation.Frame) { f.Person, f.Attempt = "a", first })
	c := New(nav, store)
	nav.Features[navigation.Assessment] = c
	if err := nav.Refresh(); err != nil {
		t.Fatal(err)
	}
	criterion := skill.Subskills[0].ID
	if err := nav.Select(criterion); err != nil {
		t.Fatal(err)
	}
	if store.answers[first][criterion] != "successful" {
		t.Fatal("first swimmer answer was not saved in draft")
	}
	c.switchLargeAssessmentPerson("b")
	if nav.Frame().Person != "b" || nav.Frame().Attempt != "draft-b" {
		t.Fatalf("switch to Blair failed: %+v", nav.Frame())
	}
	if len(store.answers["draft-b"]) != 0 {
		t.Fatal("second swimmer inherited first swimmer answers")
	}
	c.switchLargeAssessmentPerson("a")
	if nav.Frame().Person != "a" || nav.Frame().Attempt != first {
		t.Fatalf("switch back did not reuse draft: %+v", nav.Frame())
	}
	if nav.State.Rows[0].Status != "successful" {
		t.Fatal("first swimmer draft did not reappear in checklist")
	}
}

func TestLargeMultiTablePageDoesNotChangeSavedGroupScope(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("large group")
	defer window.Close()
	catalog := curriculum.Bundled()
	skill := catalog.Levels[0].Skills[0]
	store := &largeAssessmentStore{catalog: catalog}
	nav := navigation.New()
	nav.DisplayMode = ui.DisplayLarge
	nav.Push(navigation.Skills, func(f *navigation.Frame) { f.Course = "course" })
	nav.Push(navigation.MultiAssessment, func(f *navigation.Frame) { f.Skill, f.Version = skill.ID, catalog.Version })
	c := New(nav, store)
	nav.Features[navigation.Skills], nav.Features[navigation.MultiAssessment] = c, c
	group := multiGroup{Number: 1, Capacity: 11}
	for i := 0; i < 11; i++ {
		group.Attempts = append(group.Attempts, domain.MultiAttempt{AttemptID: fmt.Sprintf("attempt-%d", i), PersonID: fmt.Sprintf("p%d", i), Name: fmt.Sprintf("Swimmer %d", i)})
	}
	c.multi = &multiSession{Skill: skill, GroupCount: 1, Groups: []multiGroup{group}}
	if err := nav.Refresh(); err != nil {
		t.Fatal(err)
	}
	view := shell.New(nav, window)
	defer view.Close()
	view.Render()
	next := view.Buttons["multi-table-page-1"]
	if next == nil {
		t.Fatal("missing local table paging button")
	}
	next.OnTap()
	if c.largeTablePage != 1 || nav.Frame().Page != 0 || len(c.multi.Groups[0].Attempts) != 11 {
		t.Fatalf("table page changed group scope: local=%d route=%d attempts=%d", c.largeTablePage, nav.Frame().Page, len(c.multi.Groups[0].Attempts))
	}
	if err := c.SaveMultiGroup(); err != nil {
		t.Fatal(err)
	}
	if len(store.submitted) != 1 || len(store.submitted[0]) != 11 {
		t.Fatalf("saved IDs = %+v", store.submitted)
	}
}

func TestLargePrivateLessonKeepsUnassignedSwimmersOutsideAssessment(t *testing.T) {
	catalog := curriculum.Bundled()
	store := &largeAssessmentStore{catalog: catalog, courseLevel: "SplashPrivate", roster: []domain.Enrollment{{ID: "a", Name: "Alex", Active: true}}}
	for _, mode := range []ui.DisplayMode{ui.DisplayCompact, ui.DisplayLarge} {
		nav := navigation.New()
		nav.DisplayMode = mode
		nav.Push(navigation.Skills, func(f *navigation.Frame) { f.Course = "private" })
		state, err := New(nav, store).Build()
		if err != nil {
			t.Fatal(err)
		}
		if state.RowCount != 0 || len(state.Rows) != 0 {
			t.Fatalf("%s: private swimmer without a level appeared in assessment", mode)
		}
	}
}
