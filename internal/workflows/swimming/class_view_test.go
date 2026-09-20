package swimming_test

import (
	"fmt"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"rectablet/internal/shared/ui"
	"rectablet/internal/shell"
	"rectablet/internal/workflows/swimming/curriculum"
	"rectablet/internal/workflows/swimming/transfer"
)

func (f *uiFixture) openClassSkill(t *testing.T, level, skill string) {
	t.Helper()
	if level != "" {
		roster := must(transfer.ReadPackage("../../../examples/classes.json"))
		for i := range roster.Courses[0].Swimmers {
			roster.Courses[0].Swimmers[i].Level = &level
		}
		_ = must(f.s.Import(roster))
		check(t, f.c.Refresh())
	}
	f.row(t, 0)
	f.tap(t, "row:skills")
	for f.view.Buttons["row:"+skill] == nil && !f.view.Buttons["down"].Disabled {
		f.tap(t, "down")
	}
	f.tap(t, "row:"+skill)
}

func (f *uiFixture) enterClass(t *testing.T) {
	t.Helper()
	for i, option := range f.c.State.HeaderSwitch {
		if option.Label == "Class" {
			f.tap(t, fmt.Sprintf("header-switch-%d", i))
			if f.c.State.Screen != "class_assessment" {
				t.Fatalf("did not enter Class: %+v", f.c.State)
			}
			return
		}
	}
	t.Fatal("missing Class option")
}

func TestClassAssessmentNavigationAndSaveActions(t *testing.T) {
	f := setupUI(t)
	f.openClassSkill(t, "", "Splash1:1")
	if len(f.c.State.HeaderSwitch) != 3 {
		t.Fatal("ordinary skills should offer three modes")
	}
	f.shot(t, "class-mode-options")
	frame := f.c.Frame()
	people := must(f.s.ListRoster(frame.Course))
	f.enterClass(t)
	if f.c.State.PageCount != len(people) || f.c.State.Page != 0 || !f.view.Buttons["up"].Disabled {
		t.Fatal("Class pagination must be by student")
	}
	if f.view.Buttons["home"] != nil || f.view.Buttons["back"] != nil {
		t.Fatal("Class must replace the Home/Back rail")
	}
	wantIDs := []string{"save-student", "discard-student", "discard-all", "save-all"}
	for i, action := range f.c.State.LeftRailActions {
		if action.ID != wantIDs[i] || !f.view.Buttons[action.ID].IconOnly {
			t.Fatalf("wrong rail action at %d: %+v", i, action)
		}
	}
	for _, row := range f.c.State.Rows {
		f.tap(t, "row:"+row.ID)
	}
	firstHeader := f.c.State.Title
	firstSkill, firstStudent := f.c.State.SplitHeader.Left, f.c.State.SplitHeader.Right
	f.tap(t, "down")
	if f.c.State.Title == firstHeader || f.c.State.Rows[0].Status == "successful" {
		t.Fatal("student paging did not switch drafts/header")
	}
	if f.c.State.SplitHeader.Right == firstStudent || f.c.State.SplitHeader.Left != firstSkill {
		t.Fatal("paging should change only the student half of the header")
	}
	f.tap(t, "up")
	if f.c.State.Rows[0].Status != "successful" {
		t.Fatal("paging lost draft marks")
	}
	f.tap(t, "save-student")
	if f.c.State.PageCount != len(people)-1 || f.c.State.Title == firstHeader {
		t.Fatal("save current did not advance")
	}
	if progress := must(f.s.Progress(people[0].ID, frame.Skill, frame.Version)); progress.Successes != 1 {
		t.Fatalf("first student's assessment not saved: %+v", progress)
	}
	f.row(t, 0)
	f.tap(t, "discard-student")
	if attempts := must(f.s.ListAttempts(people[1].ID, frame.Skill, frame.Version)); len(attempts) != 0 {
		t.Fatal("discard current retained its draft")
	}
	f.row(t, 0)
	f.tap(t, "down")
	if !f.view.Buttons["down"].Disabled {
		t.Fatal("last student's next arrow is not disabled")
	}
	f.row(t, 1)
	f.tap(t, "save-all")
	if f.c.State.Screen != "skill_swimmers" {
		t.Fatal("save all did not return to skill roster")
	}
	for _, person := range people[2:] {
		attempts := must(f.s.ListAttempts(person.ID, frame.Skill, frame.Version))
		if len(attempts) != 1 || attempts[0].SubmittedAt == nil || attempts[0].Success {
			t.Fatalf("partial draft not saved correctly for %s: %+v", person.Name, attempts)
		}
	}
	var successes int
	check(t, f.s.DB.QueryRow("SELECT COUNT(*) FROM results WHERE status='successful'").Scan(&successes))
	if successes != 4 {
		t.Fatalf("saved marks = %d, want 4", successes)
	}
}

func TestClassDiscardAllKeepsSavedStudents(t *testing.T) {
	f := setupUI(t)
	f.openClassSkill(t, "", "Splash1:1")
	f.enterClass(t)
	for _, row := range f.c.State.Rows {
		f.tap(t, "row:"+row.ID)
	}
	f.tap(t, "save-student")
	f.row(t, 0)
	f.tap(t, "discard-all")
	var attempts, submitted int
	check(t, f.s.DB.QueryRow("SELECT COUNT(*), COUNT(submitted_at) FROM attempts").Scan(&attempts, &submitted))
	if f.c.State.Screen != "skill_swimmers" || attempts != 1 || submitted != 1 {
		t.Fatalf("discard all affected a saved attempt or left drafts: %s, %d/%d", f.c.State.Screen, submitted, attempts)
	}
}

func TestClassNamePrefixesStayStableAsQueueShrinks(t *testing.T) {
	f := setupUI(t)
	f.openClassSkill(t, "", "Splash1:1")
	people := must(f.s.ListRoster(f.c.Frame().Course))
	names := []string{"Sam Taylor", "Sam Turner", "Sam Tayler", "Solo"}
	for i, person := range people {
		_, err := f.s.DB.Exec("UPDATE swimmers SET name=? WHERE id=(SELECT swimmer_id FROM enrollments WHERE id=?)", names[i], person.ID)
		check(t, err)
	}
	check(t, f.c.Refresh())
	people = must(f.s.ListRoster(f.c.Frame().Course))
	want := map[string]string{"Sam Taylor": "Sam Taylo.", "Sam Turner": "Sam Tu.", "Sam Tayler": "Sam Tayle.", "Solo": "Solo"}
	f.enterClass(t)
	for i, person := range people {
		if got := f.view.Buttons["header-half-1"].Text; got != want[person.Name] {
			t.Fatalf("name prefix changed as queue shrank: %s displayed as %s", person.Name, got)
		}
		action := "save-student"
		if i%2 == 1 {
			action = "discard-student"
		}
		f.tap(t, action)
		if stored := must(f.s.GetEnrollment(person.ID)); stored.Name != person.Name {
			t.Fatal("header abbreviation changed the stored student name")
		}
	}
	if f.c.State.Screen != "skill_swimmers" || f.c.State.SplitHeader != nil {
		t.Fatal("compact header should be limited to Class assessment")
	}
}

func TestClassResumesDraftsAndFiltersRosterWhenSwitchingFromMulti(t *testing.T) {
	f := setupUI(t)
	f.openClassSkill(t, "", "Splash1:1")
	frame := f.c.Frame()
	people := must(f.s.ListRoster(frame.Course))
	_, err := f.s.DB.Exec("UPDATE enrollments SET active=0 WHERE id=?", people[1].ID)
	check(t, err)
	_, err = f.s.DB.Exec("UPDATE enrollments SET level='Splash2B' WHERE id=?", people[2].ID)
	check(t, err)
	existing := must(f.s.NewAttempt(people[0].ID, frame.Lesson, frame.Skill))
	check(t, f.s.Mark(existing, "Splash1:1:1", "successful"))
	check(t, f.c.Refresh())
	f.tap(t, "header-switch-1")
	f.row(t, 0) // Class should include every eligible student, not only this selection.
	f.enterClass(t)
	if f.c.State.PageCount != 2 || f.c.State.Rows[0].Status != "successful" {
		t.Fatal("Class lost an existing draft or included inactive/other-level students")
	}
	f.row(t, 0)
	if f.c.State.Rows[0].Status == "successful" {
		t.Fatal("could not uncheck a resumed mark")
	}
	f.row(t, 0)
	check(t, f.c.Back())
	f.enterClass(t)
	if f.c.State.Rows[0].Status != "successful" {
		t.Fatal("leaving and reopening Class lost a draft")
	}
	f.tap(t, "down")
	f.tap(t, "discard-student")
	if f.c.State.PageCount != 1 || f.c.State.Page != 0 || !f.view.Buttons["up"].Disabled || !f.view.Buttons["down"].Disabled {
		t.Fatal("discarding the last page did not return to the remaining student")
	}
	f.tap(t, "save-student")
	if f.c.State.Screen != "skill_swimmers" {
		t.Fatal("saving the final student did not finish Class mode")
	}
	attempts := must(f.s.ListAttempts(people[0].ID, frame.Skill, frame.Version))
	if len(attempts) != 1 || attempts[0].ID != existing || attempts[0].SubmittedAt == nil {
		t.Fatal("Class did not reuse the existing attempt")
	}
	for _, person := range people[1:] {
		if attempts := must(f.s.ListAttempts(person.ID, frame.Skill, frame.Version)); len(attempts) != 0 {
			t.Fatalf("unexpected attempt for excluded or discarded student %s", person.Name)
		}
	}
}

func TestClassActionsKeepQueueWhenStorageFails(t *testing.T) {
	for _, action := range []string{"save-student", "discard-student", "save-all", "discard-all"} {
		t.Run(action, func(t *testing.T) {
			f := setupUI(t)
			f.openClassSkill(t, "", "Splash1:1")
			f.enterClass(t)
			f.row(t, 0)
			before := f.c.State
			operation := "UPDATE"
			if strings.HasPrefix(action, "discard") {
				operation = "DELETE"
			}
			_, err := f.s.DB.Exec("CREATE TRIGGER class_write_failure BEFORE " + operation + " ON attempts BEGIN SELECT RAISE(ABORT, 'disk full'); END")
			check(t, err)
			f.tap(t, action)
			if f.c.State.Screen != before.Screen || f.c.State.PageCount != before.PageCount || f.c.State.Title != before.Title || f.c.State.Rows[0].Status != "successful" {
				t.Fatal("failed action changed queue or marks")
			}
			_, err = f.s.DB.Exec("DROP TRIGGER class_write_failure")
			check(t, err)
			f.tap(t, action)
			if f.c.State.Screen == before.Screen && f.c.State.PageCount == before.PageCount {
				t.Fatal("could not retry failed action")
			}
		})
	}
}

func TestClassChecklistLayoutAndModeRestrictions(t *testing.T) {
	for _, scenario := range []struct {
		level, skill string
		count        int
	}{
		{"Splash1", "Splash1:10-front-crawl-wearing-pfd-5-m", 4},
		{"Splash5", "Splash5:8-front-crawl", 7},
		{"Splash8", "Splash8:14", 12},
		{"Splash9", "Splash9:9-breaststroke", 6},
	} {
		t.Run(scenario.skill, func(t *testing.T) {
			f := setupUI(t)
			f.openClassSkill(t, scenario.level, scenario.skill)
			if len(f.c.State.HeaderSwitch) != 2 || f.c.State.HeaderSwitch[1].Label != "Class" {
				t.Fatal("stroke/dense skills should only offer Single and Class")
			}
			f.enterClass(t)
			skill := must(f.s.Skill(f.c.Frame().Version, scenario.skill))
			people := must(f.s.ListRoster(f.c.Frame().Course))
			if f.view.Buttons["header-half-0"].Text != skill.Number+". "+curriculum.CompactTitle(skill) || f.view.Buttons["header-half-1"].Text != "Alexandria M." || f.view.Buttons["context"] != nil || f.view.Buttons["title"] != nil {
				t.Fatal("Class should use the compact student/skill header only")
			}
			if !strings.Contains(f.c.State.Context, people[0].Name) || !strings.Contains(f.c.State.Context, skill.Name) || !strings.Contains(f.c.State.Context, "Student 1 of 4") {
				t.Fatal("full details should retain the original student, skill and position")
			}
			for _, id := range []string{"header-half-0", "header-half-1"} {
				f.tap(t, id)
				details := f.w.Canvas().Overlays().Top()
				if details == nil {
					t.Fatalf("%s did not open full details", id)
				}
				details.Hide()
			}
			if len(f.c.State.Rows) != scenario.count {
				t.Fatalf("checklist paginated: got %d subskills, want %d", len(f.c.State.Rows), scenario.count)
			}
			for i, row := range f.c.State.Rows {
				want := ui.Blue
				if i%2 == 1 {
					want = ui.Navy
				}
				if f.view.Buttons["row:"+row.ID].Fill != want {
					t.Fatalf("subskill %d does not alternate blue shades", i)
				}
				f.tap(t, "row:"+row.ID)
				if f.view.Buttons["row:"+row.ID].Fill != ui.Green {
					t.Fatal("marked subskill must stay green")
				}
				f.tap(t, "row:"+row.ID)
				if f.view.Buttons["row:"+row.ID].Fill != want {
					t.Fatal("unchecking did not restore the alternating blue shade")
				}
			}
			f.row(t, 0)
			f.w.Resize(fyne.NewSize(800, 480))
			f.shot(t, fmt.Sprintf("class-%d-subskills", scenario.count))
			for _, bounds := range []fyne.Size{fyne.NewSize(800, 480), fyne.NewSize(1200, 720)} {
				f.w.Resize(bounds)
				f.w.Canvas().Capture()
				composition := shell.ComposeScreen(f.view, f.c.Snapshot())
				var rowSize fyne.Size
				for _, component := range composition.Components {
					if !strings.HasPrefix(component.Name, "feature-class-subskill-") {
						continue
					}
					x, y, width, height := component.Position(800, 480)
					size := fyne.NewSize(width, height)
					if rowSize.IsZero() {
						rowSize = size
					} else if size != rowSize {
						t.Fatalf("subskill buttons differ in size: %v and %v", rowSize, size)
					}
					if x < 100 || y < 56 || x+width > 700 || y+height > 466.01 {
						t.Fatalf("subskill outside body: %s at %v", component.Name, []float32{x, y, width, height})
					}
					if component.Name == "feature-class-subskill-0" && y != 56 {
						t.Fatal("checklist did not reclaim the removed header row")
					}
				}
				var subskillFont float32
				for id, button := range f.view.Buttons {
					if !strings.HasPrefix(id, "row:") && !strings.HasPrefix(id, "header-half-") {
						continue
					}
					if strings.HasPrefix(id, "row:") {
						if subskillFont == 0 {
							subskillFont = button.FontSize
						} else if button.FontSize != subskillFont {
							t.Fatalf("subskill font sizes differ: %v and %v", subskillFont, button.FontSize)
						}
					}
					button.Refresh()
					for _, object := range test.WidgetRenderer(button).Objects() {
						if label, ok := object.(*ui.TextBlock); ok && label.Visible() {
							if label.Truncated {
								t.Fatalf("text clipped in %s at %v: %q", id, bounds, label.Text)
							}
							if strings.HasPrefix(id, "header-half-") && (label.Wrap || label.MaxLines != 1 || label.FontSize != 18) {
								t.Fatal("compact header must stay one line at 18pt")
							}
						}
					}
				}
			}
		})
	}
}
