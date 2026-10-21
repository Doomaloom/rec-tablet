package large

import (
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/navigation"
)

func (c *View) ComposeLargeSkillsBody(skills []navigation.Row, selected navigation.Row, swimmers []navigation.Row) ui.BodyComposer {
	return func(surface *ui.Surface) {
		left, right := Split(surface, .46)
		if len(skills) == 0 {
			ui.Empty(surface, "No skills available.")
			return
		}
		skillRows := append([]ui.Item(nil), skills...)
		for i := range skillRows {
			if skillRows[i].ID == selected.ID {
				skillRows[i].Selected = true
			}
		}
		c.largePagedGrid(ui.SelectableSurface(left), "large-skills", skillRows, 1, 6, c.Pages.Skill, func(id string) {
			if c.IsTimed(id) {
				_ = c.OpenTimedSkill(id)
				return
			}
			_ = c.SelectSkill(id)
		})
		c.largePagedGrid(right, "large-skill-swimmers", swimmers, 2, 6, c.Pages.Roster, func(id string) {
			_ = c.OpenSkillSwimmer(selected, id)
		})
	}
}

func (c *View) ComposeLargeSingleBody(rows []navigation.Row, swimmers []navigation.Row) ui.BodyComposer {
	return func(surface *ui.Surface) {
		left, right := Split(surface, .28)
		c.largePagedGrid(left, "large-single-roster", swimmers, 1, 6, c.Pages.Roster, c.SwitchPerson)
		c.RenderLargeChecklist(right, rows, func(id string) { _ = c.Actions.Select(id) })
	}
}

func (c *View) ComposeLargeSwimmerListBody(state shell.State) ui.BodyComposer {
	return func(surface *ui.Surface) {
		c.largePagedGrid(surface, "large-return-swimmers", state.Rows, 2, 6, c.Pages.Roster, func(id string) { _ = c.Actions.Select(id) })
	}
}
