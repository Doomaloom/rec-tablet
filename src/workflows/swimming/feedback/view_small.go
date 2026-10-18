package feedback

import (
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
)

func (c *Controller) composeFeedbackSwimmerBody(state shell.State) ui.BodyComposer {
	compose := c.composeFeedbackListBody(state)
	return func(surface *ui.Surface) { compose(ui.SelectableSurface(surface)) }
}

func (c *Controller) composeFeedbackSkillBody(state shell.State) ui.BodyComposer {
	return c.composeFeedbackListBody(state)
}

func (c *Controller) composeFeedbackDetailBody(state shell.State) ui.BodyComposer {
	snapshot := state
	return func(surface *ui.Surface) {
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}
		renderDetail(surface, snapshot.Rows[0])
	}
}
