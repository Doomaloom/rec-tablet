package small

import (
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
)

func (c *View) ComposeAssessmentBody(state shell.State) ui.BodyComposer {
	snapshot := state
	return func(surface *ui.Surface) {
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}
		renderCriterion(surface, snapshot.Rows[0], func(status string) { _ = c.Actions.Mark(status) })
	}
}

func (c *View) ComposeSubskillChecklistBody(state shell.State) ui.BodyComposer {
	snapshot := state
	return func(surface *ui.Surface) {
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}
		rows := append([]ui.Item(nil), snapshot.Rows...)
		for i := range rows {
			if rows[i].Status == "successful" {
				rows[i].Fill = ui.Green
				rows[i].Selected = true
			}
		}
		ui.List(surface, rows, func(id string) { _ = c.Actions.Select(id) })
	}
}
