// Package lessonplan presents read-only lesson snapshots and launches assessments.
package lessonplan

import (
	"fmt"
	"strconv"
	"strings"

	"rectablet/next/contracts"
	"rectablet/next/navigation"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/transfer"
)

type Repository interface {
	navigation.LessonReader
	LessonPlan(string) (*transfer.LessonPlan, error)
	CatalogFor(string) (curriculum.Catalog, error)
}
type Component struct {
	nav   *navigation.Controller
	store Repository
	start func(navigation.Row) error
}

func New(nav *navigation.Controller, store Repository, start func(navigation.Row) error) *Component {
	return &Component{nav: nav, store: store, start: start}
}
func (c *Component) Build() (contracts.State, error) {
	state := c.nav.InitialState()
	state.Title = "Lesson Plan"
	state.PageSize = 6
	f := c.nav.Frame()
	course, roster, err := navigation.LessonContext(c.store, f, &state)
	if err != nil {
		return state, err
	}
	state.SplitHeader = &contracts.SplitHeader{Left: "Lesson Plan · " + course.Level, Right: course.Time}
	plan, err := c.store.LessonPlan(f.Lesson)
	if err != nil {
		return state, err
	}
	rows := []navigation.Row{}
	if plan != nil {
		for i, r := range plan.Rows {
			label := strings.TrimSpace(r.Skill)
			if label == "" {
				label = "Planned activity"
			}
			rows = append(rows, navigation.Row{ID: strconv.Itoa(i), Label: fmt.Sprintf("%d. %s", i+1, label), Selected: f.Screen == navigation.LessonPlanDetail && f.PlanRow == strconv.Itoa(i)})
		}
	}
	state.Empty = "No lesson plan available for this lesson"
	c.nav.Paginate(&state, rows)
	var selected *transfer.PlannedRow
	var marking navigation.Row
	var choices []navigation.Row
	reason := "This planned activity has no matching marking skill."
	if plan != nil && f.Screen == navigation.LessonPlanDetail {
		index, e := strconv.Atoi(f.PlanRow)
		if e == nil && index >= 0 && index < len(plan.Rows) {
			selected = &plan.Rows[index]
			catalog, e := c.store.CatalogFor(course.CurriculumVersion)
			if e != nil {
				return state, e
			}
			level, ok := catalog.Level(plan.CurriculumLevel)
			if ok && selected.SkillID != "" {
				for _, skill := range level.PlannedSkills(selected.SkillID) {
					choice := navigation.Row{ID: skill.ID, Label: curriculum.CompactTitle(skill), Data: navigation.Parameters{Level: level.ID, Version: course.CurriculumVersion}}
					choices = append(choices, choice)
					if marking.ID == "" || f.Skill == skill.ID {
						marking = choice
					}
				}
				if marking.ID != "" {
					reason = "No active swimmers are assigned to this skill’s level."
					for _, person := range roster {
						if person.Active && person.Level == level.ID {
							reason = ""
							break
						}
					}
				}
			}
		}
	}
	c.compose(&state, selected, marking, choices, reason)
	return state, nil
}
func (c *Component) Select(id string) error {
	if c.nav.Frame().Screen == navigation.LessonPlan {
		page := c.nav.Frame().Page
		c.nav.Push(navigation.LessonPlanDetail, func(f *navigation.Frame) { f.PlanRow = id; f.Page = page })
	} else {
		c.nav.History.Frames[len(c.nav.History.Frames)-1].PlanRow = id
		c.nav.History.Frames[len(c.nav.History.Frames)-1].Skill = ""
	}
	return nil
}
func (c *Component) Header() error     { return nil }
func (c *Component) Mark(string) error { return nil }
func (c *Component) Submit() error     { return nil }

func (c *Component) chooseComponent(id string) error {
	return c.nav.Run(func() error {
		c.nav.History.Frames[len(c.nav.History.Frames)-1].Skill = id
		return nil
	})
}
