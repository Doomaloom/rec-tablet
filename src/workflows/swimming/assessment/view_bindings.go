package assessment

import (
	"fmt"

	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/assessment/view/large"
	"rectablet/src/workflows/swimming/assessment/view/shared"
	"rectablet/src/workflows/swimming/assessment/view/small"
)

func (c *Controller) smallView() *small.View {
	return &small.View{Controller: c.Controller}
}

func (c *Controller) sharedView() *shared.View {
	return &shared.View{
		Controller:   c.Controller,
		AdjustGroups: c.adjustMultiGroupCount,
		PageGroup:    c.pageMultiGroup,
		PassStudent:  c.PassMultiStudent,
		EditStudent:  c.EditMultiStudent,
	}
}

func (c *Controller) largeView() *large.View {
	return &large.View{
		Controller: c.Controller,
		Shared:     c.sharedView(),
		Pages: large.Pages{
			Skill: &c.largeSkillPage, Roster: &c.largeRosterPage,
			Class: &c.largeClassPage, Group: &c.largeGroupPage,
			Member: &c.largeMemberPage, MultiRoster: &c.largeMultiRosterPage,
			Table: &c.largeTablePage, Heat: &c.largeHeatPage,
		},
		IsTimed:          func(id string) bool { _, timed := benchmarkFor(id); return timed },
		OpenTimedSkill:   c.openLargeTimedSkill,
		OpenSkillSwimmer: c.openLargeSkillSwimmer,
		SelectSkill: func(id string) error {
			return c.Run(func() error { c.largeSkillID = id; c.largeRosterPage = 0; return nil })
		},
		SwitchPerson:      c.switchLargeAssessmentPerson,
		ClassRoster:       c.largeClassRows,
		SelectClassPerson: c.selectLargeClassPerson,
		MultiRows:         c.largeMultiRows,
		SelectGroup:       c.selectLargeMultiGroup,
		LapsInput:         func() string { return c.timedLapsInput },
		SelectionCount:    func() int { return len(c.timedSelection) },
		AddDigit: func(digit string) {
			_ = c.Run(func() error {
				if len(c.timedLapsInput) < 3 {
					c.timedLapsInput += digit
				}
				return nil
			})
		},
		ClearLaps:        func() { _ = c.Run(func() error { c.timedLapsInput = ""; return nil }) },
		StartHeat:        c.startLargeTimedHeat,
		SelectedSwimmers: c.largeSelectedSwimmers,
	}
}

func (c *Controller) largeClassRows() []ui.Item {
	var rows []ui.Item
	current := c.classAttempts[c.Frame().Page].PersonID
	for _, student := range c.classRoster {
		status := c.classStatus[student.PersonID]
		if status == "" {
			status = "Pending"
		}
		row := ui.Item{ID: student.PersonID, Label: student.Name + " · " + status, Selected: student.PersonID == current, Disabled: status != "Pending"}
		if row.Selected {
			row.Fill = ui.Green
		}
		rows = append(rows, row)
	}
	return rows
}

func (c *Controller) selectLargeClassPerson(id string) {
	_ = c.Run(func() error {
		for i, student := range c.classAttempts {
			if student.PersonID == id {
				c.Frames[len(c.Frames)-1].Page = i
				break
			}
		}
		return nil
	})
}

func (c *Controller) largeMultiRows(group, count int) ([]ui.Item, []ui.Item) {
	var groups, members []ui.Item
	for i := 0; i < count; i++ {
		g := c.multi.Groups[i]
		row := ui.Item{ID: fmt.Sprint(i), Label: fmt.Sprintf("Group %d · %d/%d", g.Number, len(g.Members), g.Capacity), Selected: i == group}
		if row.Selected {
			row.Fill = ui.Green
		}
		groups = append(groups, row)
	}
	for _, student := range c.multi.Roster {
		if containsString(c.multi.Groups[group].Members, student.ID) {
			members = append(members, ui.Item{ID: student.ID, Label: student.Name, Selected: true, Fill: ui.Green})
		}
	}
	return groups, members
}

func (c *Controller) selectLargeMultiGroup(id string) {
	var index int
	if _, err := fmt.Sscan(id, &index); err != nil {
		return
	}
	_ = c.Run(func() error { c.multi.AssignmentGroup = index; c.largeMemberPage = 0; return nil })
}

func (c *Controller) largeSelectedSwimmers() []ui.Item {
	var swimmers []ui.Item
	for _, id := range c.timedSelection {
		student, err := c.Store.GetEnrollment(id)
		if err != nil {
			continue
		}
		swimmers = append(swimmers, ui.Item{ID: id, Label: student.Name, Selected: true, Fill: ui.Green})
	}
	return swimmers
}
