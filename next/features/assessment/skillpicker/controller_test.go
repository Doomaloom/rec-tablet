package skillpicker

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"rectablet/next/contracts"
	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/navigation"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
)

func paginate(state *navigation.State, rows []navigation.Row) {
	state.Rows = rows
	state.RowCount = len(rows)
}

func TestSkillsFiltersLevelsAndPreservesSelection(t *testing.T) {
	skill := curriculum.Skill{ID: "level:1", Number: "1", Name: "Float"}
	catalog := curriculum.Catalog{Levels: []curriculum.Level{{ID: "level", Skills: []curriculum.Skill{skill, {ID: "level:2", Number: "2", Name: "Glide"}}}}}
	for _, selected := range []string{"level:1", "level:2", "missing", "removed"} {
		t.Run(selected, func(t *testing.T) {
			c := New(Options{
				Storage: Storage{
					CatalogFor: func(string) (curriculum.Catalog, error) { return catalog, nil },
					Skill:      func(string, string) (curriculum.Skill, error) { return skill, nil },
					Progress:   func(string, string, string) (domain.Progress, error) { return domain.Progress{}, nil },
				},
				Paginate: paginate,
			})
			c.Layout.Panels = true
			c.State.SelectedID = selected
			state, err := c.Build(common.Input{
				Frame:      navigation.Frame{Screen: navigation.Skills},
				Course:     domain.Course{CurriculumVersion: "course-version"},
				Attendance: map[string]string{"a": "present", "b": "present", "c": "present"},
				Roster: []domain.Enrollment{
					{ID: "inactive", Level: "ignored", Active: false},
					{ID: "unassigned", Active: true},
					{ID: "missing", Level: "missing", Active: true},
					{ID: "a", Level: "level", Active: true},
					{ID: "b", Level: "level", Active: true},
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(state.Rows) != 3 || !state.Rows[0].Disabled || state.Rows[1].ID != skill.ID {
				t.Fatalf("skill rows = %+v", state.Rows)
			}
			wantSelected := skill.ID
			if selected == "level:2" {
				wantSelected = selected
			}
			if c.State.SelectedID != wantSelected || state.ComposeBody == nil || state.PageSize != 3 {
				t.Fatalf("selection = %q, state = %+v", c.State.SelectedID, state)
			}
			if got := navigation.Destination(state.Rows[1]); got.Level != "level" || got.Version != "course-version" {
				t.Fatalf("skill destination = %+v", got)
			}
		})
	}
}

func TestSwimmersUseRouteVersionAndBindCorrectionsToEachPerson(t *testing.T) {
	for _, panels := range []bool{false, true} {
		t.Run(map[bool]string{false: "list", true: "panels"}[panels], func(t *testing.T) {
			var removed [][3]string
			c := New(Options{
				Navigation: common.Navigation{Run: func(fn func() error) error { return fn() }, Saved: func(string) {}},
				Storage: Storage{
					Skill: func(version, id string) (curriculum.Skill, error) {
						if version != "route-version" || id != "skill" {
							t.Fatalf("Skill(%q, %q)", version, id)
						}
						return curriculum.Skill{ID: id}, nil
					},
					Progress: func(person, skill, version string) (domain.Progress, error) {
						if skill != "skill" || version != "route-version" {
							t.Fatalf("Progress version or skill: %q, %q", skill, version)
						}
						return domain.Progress{Successes: map[string]int{"a": 5, "b": 1, "c": -1}[person]}, nil
					},
					RemoveLatestSuccess: func(person, skill, version string) error {
						removed = append(removed, [3]string{person, skill, version})
						return nil
					},
				},
				Actions:  Actions{ModeSwitch: func(curriculum.Skill, bool) []contracts.HeaderSwitchOption { return nil }},
				Paginate: paginate,
			})
			c.Layout.Panels = panels
			state, err := c.Build(common.Input{
				Frame:      navigation.Frame{Screen: navigation.SkillSwimmers, Parameters: navigation.Parameters{Skill: "skill", Level: "level", Version: "route-version"}},
				Course:     domain.Course{CurriculumVersion: "course-version"},
				Attendance: map[string]string{"a": "present", "b": "present", "c": "present"},
				Roster: []domain.Enrollment{
					{ID: "a", Name: "Alex", Level: "level", Active: true},
					{ID: "b", Name: "Blair", Level: "level", Active: true},
					{ID: "c", Name: "Casey", Level: "level", Active: true},
					{ID: "inactive", Level: "level"},
					{ID: "other", Level: "other", Active: true},
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(state.Rows) != 3 || state.Rows[0].ProgressSegments != 3 || state.Rows[2].ProgressSegments != 0 {
				t.Fatalf("swimmer rows = %+v", state.Rows)
			}
			for _, row := range state.Rows[:2] {
				if row.HoldDuration != 2*time.Second {
					t.Fatalf("hold duration = %v", row.HoldDuration)
				}
				row.OnHold()
			}
			if state.Rows[2].OnHold != nil {
				t.Fatal("correction enabled without a success")
			}
			want := [][3]string{{"a", "skill", "route-version"}, {"b", "skill", "route-version"}}
			if !reflect.DeepEqual(removed, want) {
				t.Fatalf("corrections = %v, want %v", removed, want)
			}
			if got := navigation.Destination(state.Rows[1]); got.Skill != "skill" || got.Version != "route-version" || got.Level != "level" {
				t.Fatalf("swimmer destination = %+v", got)
			}
			if state.ComposeBody == nil {
				t.Fatal("missing swimmer template")
			}
		})
	}
}

func TestSkillsPropagateCatalogFailureBeforePagination(t *testing.T) {
	want := errors.New("catalog unavailable")
	c := New(Options{
		Storage:  Storage{CatalogFor: func(string) (curriculum.Catalog, error) { return curriculum.Catalog{}, want }},
		Paginate: func(*navigation.State, []navigation.Row) { t.Fatal("paginated after a storage error") },
	})
	_, err := c.Build(common.Input{Frame: navigation.Frame{Screen: navigation.Skills}})
	if !errors.Is(err, want) {
		t.Fatalf("Build error = %v, want %v", err, want)
	}
}
