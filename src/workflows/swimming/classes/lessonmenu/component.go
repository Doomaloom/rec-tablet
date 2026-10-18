// Package lessonmenu renders the lesson action menu.
package lessonmenu

import (
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/domain"
)

// Component has a selection callback and receives a snapshot to render.
type Component struct {
	onSelect func(string)
}

func New(onSelect func(string)) *Component {
	return &Component{onSelect: onSelect}
}

func (c *Component) Compose(state shell.State) ui.BodyComposer {
	return func(surface *ui.Surface) {
		if state.RowCount == 0 {
			ui.Empty(surface, state.Empty)
			return
		}
		if state.DisplayMode.IsLarge() {
			ui.Grid(surface, state.Rows, 3, 1, c.onSelect)
			return
		}
		ui.List(surface, state.Rows, c.onSelect)
	}
}

// Configure presents the selected lesson and its available actions.
func (c *Component) Configure(state *shell.State, course domain.Course) []shell.Row {
	if state.DisplayMode.IsLarge() {
		state.PageSize = 3
	}
	level := course.Level
	if level == "" {
		level = course.Name
	}
	state.SplitHeader = &shell.SplitHeader{Left: level, Right: course.Time}
	state.PrimaryInfo = ""
	return []shell.Row{
		{ID: "attendance", Label: "Attendance"},
		{ID: "skills", Label: "Skills"},
		{ID: "feedback_swimmers", Label: "Feedback"},
	}
}
