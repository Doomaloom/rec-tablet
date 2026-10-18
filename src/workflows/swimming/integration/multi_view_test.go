package integration_test

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/domain"
	"rectablet/src/workflows/swimming/transfer"
)

func (f *uiFixture) multiPresent(t *testing.T) []string {
	t.Helper()
	frame := f.c.Frame()
	var people []string
	for _, student := range must(f.s.ListRoster(frame.Course)) {
		if student.Active && student.Level == frame.Level {
			check(t, f.s.SetAttendance(student.ID, frame.Lesson, "present"))
			people = append(people, student.ID)
		}
	}
	return people
}

func (f *uiFixture) pagedRow(t *testing.T, id string) {
	t.Helper()
	check(t, f.c.Page(-10000))
	for {
		if f.view.Buttons["row:"+id] != nil {
			f.tap(t, "row:"+id)
			return
		}
		if f.view.Buttons["down"].Disabled {
			t.Fatalf("missing roster row %s", id)
		}
		f.tap(t, "down")
	}
}

func (f *uiFixture) multiCount(t *testing.T) int {
	t.Helper()
	for _, component := range shell.ComposeScreen(f.view, f.c.State).Components {
		if component.Name == "feature-multi-count-value" {
			value, err := strconv.Atoi(component.Object.(*ui.TextBlock).Text)
			check(t, err)
			return value
		}
	}
	t.Fatal("missing group count display")
	return 0
}

func (f *uiFixture) chooseGroups(t *testing.T, count int) {
	t.Helper()
	current := f.multiCount(t)
	for current < count {
		f.tap(t, "multi-count-plus")
		current++
	}
	for current > count {
		f.tap(t, "multi-count-minus")
		current--
	}
	if f.multiCount(t) != count {
		t.Fatal("group count did not change")
	}
	f.tap(t, "submit")
}

func (f *uiFixture) assignMulti(t *testing.T, groups [][]string) {
	t.Helper()
	f.tap(t, "header-switch-1")
	if f.c.State.Screen != "multi_group_count" {
		t.Fatal("missing group count setup")
	}
	f.chooseGroups(t, len(groups))
	for index, group := range groups {
		if index > 0 {
			f.tap(t, "multi-group-next")
		}
		for _, person := range group {
			f.pagedRow(t, person)
		}
	}
	if !f.c.State.CanSubmit {
		t.Fatal("complete group assignment cannot start")
	}
	f.tap(t, "submit")
	if f.c.State.Screen != "multi_assessment" {
		t.Fatalf("could not start marking: %+v", f.c.State)
	}
}

func (f *uiFixture) multiAttempt(t *testing.T, person string) domain.Attempt {
	t.Helper()
	frame := f.c.Frame()
	attempts := must(f.s.ListAttempts(person, frame.Skill, frame.Version))
	if len(attempts) != 1 {
		t.Fatalf("got %d attempts for student", len(attempts))
	}
	return attempts[0]
}

func TestMultiGroupPassChecklistAndBatchSave(t *testing.T) {
	f := setupUI(t)
	f.openClassSkill(t, "", "Splash1:1")
	frame := f.c.Frame()
	people := f.multiPresent(t)
	f.assignMulti(t, [][]string{people[:2], people[2:]})
	if f.c.State.PageCount != 2 || len(f.c.State.Rows) != 2 {
		t.Fatal("pages are not groups")
	}
	for i, id := range []string{"save-group", "discard-group", "discard-all", "save-all"} {
		if f.c.State.LeftRailActions[i].ID != id {
			t.Fatal("incorrect rail actions")
		}
	}
	f.tap(t, "multi-pass:"+people[0])
	f.tap(t, "multi-pass:"+people[0])
	if a := f.multiAttempt(t, people[0]); a.SubmittedAt != nil {
		t.Fatal("pass submitted the student early")
	}
	f.tap(t, "multi-fail:"+people[0])
	if f.c.State.Rows[0].Status != "successful" || f.c.State.Rows[1].Status != "successful" {
		t.Fatal("existing marks not preselected")
	}
	f.row(t, 0)
	f.tap(t, "back")
	if !f.view.Buttons["multi-pass:"+people[0]].Selected {
		t.Fatal("discard did not restore pass")
	}
	f.tap(t, "multi-fail:"+people[1])
	f.row(t, 0)
	f.tap(t, "home")
	if !f.view.Buttons["multi-fail:"+people[1]].Selected {
		t.Fatal("partial result not shown")
	}
	partial := f.multiAttempt(t, people[1])
	if partial.SubmittedAt != nil {
		t.Fatal("checklist saved an attempt instead of a draft")
	}
	answers := must(f.s.Draft(partial.ID))
	if answers["Splash1:1:1"] != "successful" || answers["Splash1:1:2"] != "not_yet" {
		t.Fatal("wrong partial draft", answers)
	}
	f.tap(t, "down")
	f.tap(t, "multi-fail:"+people[2])
	f.tap(t, "home") // Explicit all-failed result.
	if !f.view.Buttons["multi-fail:"+people[2]].Selected {
		t.Fatal("empty checklist did not mark incomplete")
	}
	f.tap(t, "up")
	if !f.view.Buttons["multi-pass:"+people[0]].Selected || !f.view.Buttons["multi-fail:"+people[1]].Selected {
		t.Fatal("paging lost marks")
	}
	f.tap(t, "save-group")
	if f.c.State.PageCount != 1 || f.c.State.SplitHeader.Right != "Group 2 of 2" {
		t.Fatal("saving group did not advance with stable numbering")
	}
	f.tap(t, "save-all")
	if f.c.State.Screen != "skill_swimmers" {
		t.Fatal("finished session did not return to roster")
	}
	for i, person := range people {
		attempts := must(f.s.ListAttempts(person, frame.Skill, frame.Version))
		if len(attempts) != 1 || attempts[0].SubmittedAt == nil || attempts[0].Success != (i == 0) {
			t.Fatalf("wrong saved result for student %d: %+v", i, attempts)
		}
	}
	if progress := must(f.s.Progress(people[0], frame.Skill, frame.Version)); progress.Successes != 1 {
		t.Fatal("pass did not count")
	}
}

func TestMultiSetupFiltersAndBalancesStudents(t *testing.T) {
	f := setupUI(t)
	f.openClassSkill(t, "", "Splash1:1")
	people := f.multiPresent(t)
	frame := f.c.Frame()
	check(t, f.s.SetAttendance(people[3], frame.Lesson, "absent"))
	f.tap(t, "header-switch-1")
	if f.multiCount(t) != 1 || !f.view.Buttons["multi-count-minus"].Disabled {
		t.Fatal("group count should start at one with minus disabled")
	}
	f.tap(t, "multi-count-minus")
	if f.multiCount(t) != 1 {
		t.Fatal("count went below one")
	}
	f.tap(t, "multi-count-plus")
	if f.c.State.Progress != "1 group of 2, 1 group of 1" || f.multiCount(t) != 2 || f.c.State.Screen != "multi_group_count" {
		t.Fatal("live split preview did not update")
	}
	f.tap(t, "multi-count-plus")
	if !f.view.Buttons["multi-count-plus"].Disabled || f.c.State.Progress != "3 groups of 1" {
		t.Fatal("maximum count not enforced")
	}
	f.tap(t, "multi-count-plus")
	if f.multiCount(t) != 3 {
		t.Fatal("count exceeded roster")
	}
	f.tap(t, "multi-count-minus")
	f.compactLayout(t, "multi-group-count")
	f.chooseGroups(t, 2)
	f.compactLayout(t, "multi-group-assignment")
	if f.c.State.RowCount != 3 || f.c.State.CanSubmit {
		t.Fatal("wrong assignment roster")
	}
	f.pagedRow(t, people[0])
	f.pagedRow(t, people[1])
	f.pagedRow(t, people[2])
	if !strings.Contains(f.c.State.PrimaryInfo, "2 of 2 selected") || f.c.State.CanSubmit {
		t.Fatal("capacity not enforced")
	}
	f.tap(t, "multi-group-next")
	f.pagedRow(t, people[0])
	if !strings.Contains(f.c.State.PrimaryInfo, "0 of 1 selected") {
		t.Fatal("student assigned twice")
	}
	f.pagedRow(t, people[2])
	if !f.c.State.CanSubmit {
		t.Fatal("complete assignment blocked")
	}
	// Release a student in each group, then swap them without losing other selections.
	f.pagedRow(t, people[2])
	f.tap(t, "multi-group-previous")
	f.pagedRow(t, people[1])
	f.pagedRow(t, people[2])
	f.tap(t, "multi-group-next")
	f.pagedRow(t, people[1])
	if !f.c.State.CanSubmit {
		t.Fatal("reassignment did not complete")
	}
	f.tap(t, "back")
	f.chooseGroups(t, 2)
	if !f.c.State.CanSubmit {
		t.Fatal("reselecting the same count lost assignments")
	}
	f.tap(t, "back")
	f.chooseGroups(t, 1)
	if f.c.State.CanSubmit || f.c.State.Progress != "0/3 assigned" {
		t.Fatal("changed group count kept assignments")
	}
	f.tap(t, "home")
	if f.c.State.Screen != "skill_swimmers" {
		t.Fatal("cancel did not return to roster")
	}
	var attempts int
	check(t, f.s.DB.QueryRow("SELECT COUNT(*) FROM attempts").Scan(&attempts))
	if attempts != 0 {
		t.Fatal("setup created assessment attempts")
	}
}

func TestMultiEligibilityEmptyAndSingleStudent(t *testing.T) {
	f := setupUI(t)
	f.openClassSkill(t, "", "Splash1:1")
	frame := f.c.Frame()
	people := must(f.s.ListRoster(frame.Course))
	f.tap(t, "header-switch-1")
	if f.c.State.RowCount != 0 || !strings.Contains(f.c.State.Empty, "Take attendance") {
		t.Fatal("unmarked attendance included")
	}
	f.tap(t, "header-switch-0")
	for _, p := range people {
		check(t, f.s.SetAttendance(p.ID, frame.Lesson, "present"))
	}
	_, err := f.s.DB.Exec("UPDATE enrollments SET active=0 WHERE id=?", people[1].ID)
	check(t, err)
	_, err = f.s.DB.Exec("UPDATE enrollments SET level='Splash2B' WHERE id=?", people[2].ID)
	check(t, err)
	check(t, f.s.SetAttendance(people[3].ID, frame.Lesson, "unmarked"))
	f.assignMulti(t, [][]string{{people[0].ID}})
	if len(f.c.State.Rows) != 1 || !f.view.Buttons["up"].Disabled || !f.view.Buttons["down"].Disabled {
		t.Fatal("single-student group not isolated")
	}
	f.tap(t, "discard-all")
}

func TestMultiDiscardRestoresResumedDraftsAndKeepsSavedGroups(t *testing.T) {
	f := setupUI(t)
	f.openClassSkill(t, "", "Splash1:1")
	frame := f.c.Frame()
	people := f.multiPresent(t)
	existing := must(f.s.NewAttempt(people[2], frame.Lesson, frame.Skill))
	original := map[string]string{"Splash1:1:1": "successful"}
	check(t, f.s.ReplaceDraftAnswers(existing, original))
	f.assignMulti(t, [][]string{people[:2], people[2:]})
	f.tap(t, "multi-pass:"+people[0])
	f.tap(t, "save-group")
	f.tap(t, "multi-pass:"+people[2])
	f.tap(t, "multi-pass:"+people[3])
	f.tap(t, "discard-all")
	if got := must(f.s.Draft(existing)); !reflect.DeepEqual(got, original) {
		t.Fatal("resumed draft not restored", got)
	}
	if got := must(f.s.ListAttempts(people[3], frame.Skill, frame.Version)); len(got) != 0 {
		t.Fatal("new discarded draft remains")
	}
	for _, person := range people[:2] {
		got := must(f.s.ListAttempts(person, frame.Skill, frame.Version))
		if len(got) != 1 || got[0].SubmittedAt == nil {
			t.Fatal("discard changed saved group")
		}
	}
	// Reopening starts at group-count setup, while the existing student draft survives.
	f.tap(t, "header-switch-1")
	if f.c.State.Screen != "multi_group_count" {
		t.Fatal("group assignments survived session")
	}
	if f.multiCount(t) != 1 {
		t.Fatal("previous group count reused")
	}

}

func TestMultiChecklistAllPassedAndEscape(t *testing.T) {
	f := setupUI(t)
	f.openClassSkill(t, "", "Splash1:1")
	people := f.multiPresent(t)
	f.assignMulti(t, [][]string{people[:2], people[2:]})
	f.tap(t, "down")
	f.tap(t, "multi-fail:"+people[2])
	f.row(t, 0)
	f.w.Canvas().OnTypedKey()(&fyne.KeyEvent{Name: fyne.KeyEscape})
	if f.c.State.Page != 1 || f.view.Buttons["multi-pass:"+people[2]].Selected || f.view.Buttons["multi-fail:"+people[2]].Selected {
		t.Fatal("Escape committed changes or changed groups")
	}
	f.tap(t, "multi-fail:"+people[2])
	if f.c.State.Rows[0].Status == "successful" {
		t.Fatal("abandoned local edit leaked")
	}
	for _, row := range f.c.State.Rows {
		f.tap(t, "row:"+row.ID)
	}
	f.tap(t, "home")
	if f.c.State.Page != 1 || !f.view.Buttons["multi-pass:"+people[2]].Selected {
		t.Fatal("all-passed checklist did not return a pass to the same group")
	}
	f.tap(t, "discard-group")
	if f.c.State.PageCount != 1 || f.c.State.SplitHeader.Right != "Group 1 of 2" {
		t.Fatal("discard last group did not clamp navigation")
	}
	f.tap(t, "discard-group")
	if f.c.State.Screen != "skill_swimmers" {
		t.Fatal("last group did not exit")
	}
}

func TestMultiFailedActionsPreserveDraftsAndGroups(t *testing.T) {
	for _, action := range []string{"save-group", "discard-group", "save-all", "discard-all", "pass", "checklist"} {
		t.Run(action, func(t *testing.T) {
			f := setupUI(t)
			f.openClassSkill(t, "", "Splash1:1")
			people := f.multiPresent(t)
			f.assignMulti(t, [][]string{people[:2], people[2:]})
			f.tap(t, "multi-pass:"+people[0])
			before := f.c.State
			table, operation := "attempts", "UPDATE"
			if strings.HasPrefix(action, "discard") {
				operation = "DELETE"
			}
			if action == "pass" || action == "checklist" {
				table = "drafts"
			}
			_, err := f.s.DB.Exec("CREATE TRIGGER multi_write_failure BEFORE " + operation + " ON " + table + " BEGIN SELECT RAISE(ABORT, 'disk full'); END")
			check(t, err)
			button := action
			if action == "pass" {
				button = "multi-pass:" + people[1]
			}
			if action == "checklist" {
				f.tap(t, "multi-fail:"+people[1])
				f.row(t, 0)
				button = "home"
			}
			f.tap(t, button)
			if !f.c.State.Error {
				t.Fatal("write failure not reported")
			}
			if action == "checklist" {
				if f.c.State.Screen != "multi_student" || f.c.State.Rows[0].Status != "successful" {
					t.Fatal("failed checklist save lost edit")
				}
			} else if f.c.State.Screen != before.Screen || f.c.State.PageCount != before.PageCount || f.c.State.Rows[0].Status != "successful" {
				t.Fatal("failed action changed groups or marks")
			}
			for _, person := range people {
				a := f.multiAttempt(t, person)
				if a.SubmittedAt != nil {
					t.Fatal("batch failure partially submitted")
				}
			}
			_, err = f.s.DB.Exec("DROP TRIGGER multi_write_failure")
			check(t, err)
			f.c.Dismiss()
			f.tap(t, button)
			if f.c.State.Error {
				t.Fatal("retry failed")
			}
		})
	}
}

func TestMultiGroupAndChecklistLayouts(t *testing.T) {
	for _, count := range []int{1, 4, 8, 9, 12, 16, 20} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			f := setupUI(t)
			pkg := must(transfer.ReadPackage("../../../../examples/classes.json"))
			pkg.Courses[0].Swimmers = nil
			level := "Splash9"
			for i := 0; i < count; i++ {
				pkg.Courses[0].Swimmers = append(pkg.Courses[0].Swimmers, transfer.SwimmerInput{SourceID: fmt.Sprintf("layout-%d", i), Name: fmt.Sprintf("Alexandria Longsurname %02d", i), Level: &level})
			}
			_ = must(f.s.Import(pkg))
			check(t, f.c.Refresh())
			f.openClassSkill(t, "", "Splash9:12")
			people := f.multiPresent(t)
			f.assignMulti(t, [][]string{people})
			if len(f.c.State.Rows) != count || f.c.State.PageCount != 1 {
				t.Fatal("group split across pages")
			}
			f.tap(t, "multi-pass:"+people[0])
			f.compactLayout(t, fmt.Sprintf("multi-group-%d", count))
			for _, person := range people {
				if f.view.Buttons["multi-pass:"+person] == nil || f.view.Buttons["multi-fail:"+person] == nil {
					t.Fatal("missing student result controls")
				}
			}
			composition := shell.ComposeScreen(f.view, f.c.Snapshot())
			var nameLabel *ui.TextBlock
			for _, component := range composition.Components {
				if component.Name == "feature-multi-name:"+people[0] {
					nameLabel, _ = component.Object.(*ui.TextBlock)
					break
				}
			}
			if nameLabel == nil || !strings.Contains(nameLabel.Text, "Alexandria Longsurname 00") {
				t.Fatal("full student name should appear in the main row text")
			}
			f.tap(t, "multi-fail:"+people[0])
			if len(f.c.State.Rows) != 6 || f.c.State.PageCount != 1 {
				t.Fatal("checklist did not fit full skill")
			}
			f.compactLayout(t, fmt.Sprintf("multi-checklist-%d", count))
			for id, button := range f.view.Buttons {
				if !strings.HasPrefix(id, "row:") {
					continue
				}
				for _, object := range test.WidgetRenderer(button).Objects() {
					if text, ok := object.(*ui.TextBlock); ok && text.Visible() && text.Truncated {
						t.Fatalf("criterion clipped: %s", text.Text)
					}
				}
			}
		})
	}
}

func TestMultiStartFailureKeepsAssignmentsAndCreatesNoPartialAttempts(t *testing.T) {
	f := setupUI(t)
	f.openClassSkill(t, "", "Splash1:1")
	people := f.multiPresent(t)
	f.tap(t, "header-switch-1")
	f.chooseGroups(t, 1)
	for _, person := range people {
		f.pagedRow(t, person)
	}
	_, err := f.s.DB.Exec("CREATE TRIGGER multi_start_failure BEFORE INSERT ON attempts WHEN NEW.enrollment_id = '" + people[1] + "' BEGIN SELECT RAISE(ABORT, 'disk full'); END")
	check(t, err)
	f.tap(t, "submit")
	if f.c.State.Screen != "multi_selection" || !f.c.State.CanSubmit || !f.c.State.Error {
		t.Fatal("failed start lost assignment screen")
	}
	var count int
	check(t, f.s.DB.QueryRow("SELECT COUNT(*) FROM attempts").Scan(&count))
	if count != 0 {
		t.Fatal("failed start retained partial attempts")
	}
	_, err = f.s.DB.Exec("DROP TRIGGER multi_start_failure")
	check(t, err)
	f.c.Dismiss()
	f.tap(t, "submit")
	if f.c.State.Screen != "multi_assessment" || len(f.c.State.Rows) != len(people) {
		t.Fatal("could not retry completed assignment")
	}
	f.tap(t, "multi-pass:"+people[0])
	check(t, f.c.Back())
	if f.c.State.Screen != "skill_swimmers" {
		t.Fatal("setup frames remained in history")
	}
	f.assignMulti(t, [][]string{people[:2], people[2:]})
	if !f.view.Buttons["multi-pass:"+people[0]].Selected {
		t.Fatal("leaving and reassigning lost persisted draft")
	}
	f.tap(t, "discard-all")
}

func TestMultiUnevenGroupPagesSaveAllPendingGroups(t *testing.T) {
	f := setupUI(t)
	f.openClassSkill(t, "", "Splash1:1")
	frame := f.c.Frame()
	people := f.multiPresent(t)
	f.assignMulti(t, [][]string{people[:2], people[2:3], people[3:]})
	for index, expected := range []int{2, 1, 1} {
		if f.c.State.PageCount != 3 || len(f.c.State.Rows) != expected || f.c.State.SplitHeader.Right != fmt.Sprintf("Group %d of 3", index+1) {
			t.Fatal("uneven group lost its page boundary")
		}
		f.compactLayout(t, fmt.Sprintf("multi-uneven-group-%d", index+1))
		for _, row := range f.c.State.Rows {
			f.tap(t, "multi-pass:"+row.ID)
		}
		if index < 2 {
			f.tap(t, "down")
		}
	}
	f.tap(t, "save-all")
	if f.c.State.Screen != "skill_swimmers" {
		t.Fatal("batch save did not finish all groups")
	}
	for _, person := range people {
		attempts := must(f.s.ListAttempts(person, frame.Skill, frame.Version))
		if len(attempts) != 1 || attempts[0].SubmittedAt == nil || !attempts[0].Success {
			t.Fatal("batch save omitted a group member")
		}
	}
}
