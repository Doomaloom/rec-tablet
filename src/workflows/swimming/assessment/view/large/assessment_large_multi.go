package large

import (
	"fmt"

	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/assessment/view/shared"
	"rectablet/src/workflows/swimming/navigation"
)

func (c *View) ComposeLargeMultiSelectionBody(state navigation.State, group, count int) ui.BodyComposer {
	return func(surface *ui.Surface) {
		left, right := Split(surface, .45)
		groupsPanel := ui.Panel(left, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := left.BodyBounds(w, h)
			return x, y, width, (height - ui.Gap) * .47
		})
		membersPanel := ui.Panel(left, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := left.BodyBounds(w, h)
			upper := (height - ui.Gap) * .47
			return x, y + upper + ui.Gap, width, height - upper - ui.Gap
		})
		groups, members := c.MultiRows(group, count)
		c.largePagedGrid(groupsPanel, "multi-groups", "Groups", groups, 1, 2, c.Pages.Group, c.SelectGroup)
		c.largePagedGrid(membersPanel, "multi-group-members", "Group members", members, 1, 3, c.Pages.Member, func(id string) { _ = c.Actions.Select(id) })
		c.largePagedGrid(right, "multi-assign-roster", "Assign swimmers", state.Rows, 2, 5, c.Pages.MultiRoster, func(id string) { _ = c.Actions.Select(id) })
	}
}

func (c *View) ComposeLargeMultiAssessmentBody(students []shared.MultiStudentSnapshot) ui.BodyComposer {
	return func(surface *ui.Surface) {
		pageCount := max(1, (len(students)+5)/6)
		*c.Pages.Table = max(0, min(*c.Pages.Table, pageCount-1))
		if pageCount > 1 {
			for direction, caption := range map[int]string{-1: "‹ Previous swimmers", 1: "Next swimmers ›"} {
				direction, caption := direction, caption
				button := surface.Button(fmt.Sprintf("multi-table-page-%d", direction), caption, func() {
					_ = c.Actions.Run(func() error { *c.Pages.Table = max(0, min(*c.Pages.Table+direction, pageCount-1)); return nil })
				})
				button.Fill, button.Ink = ui.Rail, ui.Blue
				surface.AddNamed(fmt.Sprintf("multi-table-navigation-%d", direction), button, func(w, h float32) (float32, float32, float32, float32) {
					x, y, width, _ := surface.BodyBounds(w, h)
					if direction < 0 {
						return x, y, (width - ui.Gap) / 2, 44
					}
					return x + (width+ui.Gap)/2, y, (width - ui.Gap) / 2, 44
				})
			}
		}
		table := ui.Panel(surface, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := surface.BodyBounds(w, h)
			if pageCount > 1 {
				return x, y + 48, width, height - 48
			}
			return x, y, width, height
		})
		start := *c.Pages.Table * 6
		c.Shared.ComposeMultiAssessmentBody(students[start:min(start+6, len(students))])(table)
	}
}
