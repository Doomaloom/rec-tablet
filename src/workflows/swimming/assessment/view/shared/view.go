package shared

import (
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/navigation"
)

func (c *View) ComposeSkillListBody(state shell.State) ui.BodyComposer {
	compose := c.ComposeAssessmentListBody(state)
	return func(surface *ui.Surface) { compose(ui.SelectableSurface(surface)) }
}

func (c *View) ComposeAttemptListBody(state shell.State) ui.BodyComposer {
	return c.ComposeAssessmentListBody(state)
}

func (c *View) ComposeAssessmentListBody(state shell.State) ui.BodyComposer {
	snapshot := state
	return func(surface *ui.Surface) {
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}
		ui.List(surface, snapshot.Rows, func(id string) { _ = c.Controller.Select(id) })
	}
}

// View renders common assessment content using navigation and feature callbacks.
type View struct {
	*navigation.Controller
	AdjustGroups func(int) error
	PageGroup    func(int) error
	PassStudent  func(string) error
	EditStudent  func(string) error
}

type MultiStudentSnapshot struct {
	PersonID, Name, Status string
	Passed, Total          int
}
