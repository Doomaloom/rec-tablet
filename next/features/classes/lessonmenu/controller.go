// Package lessonmenu renders the lesson action menu.
package lessonmenu

import (
	"rectablet/next/contracts"
	"rectablet/src/workflows/swimming/domain"
)

// Component has a selection callback and receives a snapshot to render.
type Component struct {
	Layout   Layout
	onSelect func(string)
}

func New(onSelect func(string)) *Component {
	return &Component{onSelect: onSelect}
}

// Configure presents the selected lesson and its available actions.
func (c *Component) Configure(state *contracts.State, course domain.Course) []contracts.Row {
	if c.Layout.Columns > 0 && c.Layout.Rows > 0 {
		state.PageSize = c.Layout.Columns * c.Layout.Rows
	}
	level := course.Level
	if level == "" {
		level = course.Name
	}
	state.SplitHeader = &contracts.SplitHeader{Left: level, Right: course.Time}
	state.PrimaryInfo = ""
	rows := []contracts.Row{{ID: "attendance", Label: "Attendance"}}
	if c.Layout.Plans {
		rows = append(rows, contracts.Row{ID: "lesson_plan", Label: "Lesson Plan"})
	}
	rows = append(rows, contracts.Row{ID: "skills", Label: "Skills"}, contracts.Row{ID: "feedback_swimmers", Label: "Feedback"})
	return rows
}

// Layout is selected by the feature root, independently of child rendering.
type Layout struct {
	Columns, Rows int
	Plans         bool
}
