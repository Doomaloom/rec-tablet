package attendance

import (
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/navigation"
)

func (c *Controller) composeFeatureBody(state *shell.State) {
	for i := range state.Rows {
		switch state.Rows[i].Status {
		case "present":
			state.Rows[i].Fill = ui.Green
		case "absent":
			state.Rows[i].Fill = ui.Amber
		}
	}
	if c.Frame().Screen == navigation.Attendance {
		state.ComposeBody = c.composeAttendanceRosterBody(*state)
	}
}

func (c *Controller) composeAttendanceRosterBody(state shell.State) ui.BodyComposer {
	return c.composeAttendanceListBody(state)
}

func (c *Controller) composeAttendanceListBody(state shell.State) ui.BodyComposer {
	snapshot := state
	return func(surface *ui.Surface) {
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}
		if c.DisplayMode.IsLarge() {
			c.renderLargeAttendanceRoster(surface, snapshot)
			return
		}
		c.renderSmallAttendanceRoster(surface, snapshot)
	}
}
