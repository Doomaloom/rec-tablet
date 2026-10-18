package assessment

import (
	"fmt"

	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/assessment/view/large"
	"rectablet/src/workflows/swimming/assessment/view/shared"
	"rectablet/src/workflows/swimming/assessment/view/small"
)

func (c *Component) bindComponents() {
	actions := shared.Actions{Select: c.navigation.Select, Mark: c.navigation.Mark, Submit: c.navigation.Submit, Run: c.navigation.Run}
	c.skillPicker.paginate = c.navigation.Paginate
	if c.store != nil {
		c.skillPicker.data = skillPickerData{
			CatalogFor: c.store.CatalogFor,
			Progress:   c.store.Progress,
			Skill:      c.store.Skill,
		}
	}
	c.skillPicker.actions = skillPickerActions{
		RemoveLatestSuccess:   c.RemoveLatestSuccess,
		openLargeSelectedMode: c.openLargeSelectedMode,
		skillModeSwitch:       c.skillModeSwitch,
	}
	c.skillPicker.shared = &shared.View{Actions: actions}
	c.skillPicker.large = &large.View{Actions: actions, Shared: c.skillPicker.shared,
		Pages:            large.Pages{Skill: &c.skillPicker.skillPage, Roster: &c.skillPicker.rosterPage},
		IsTimed:          func(id string) bool { _, ok := benchmarkFor(id); return ok },
		OpenTimedSkill:   c.openLargeTimedSkill,
		OpenSkillSwimmer: c.openLargeSkillSwimmer,
		SelectSkill: func(id string) error {
			return c.navigation.Run(func() error { c.skillPicker.selectedID = id; c.skillPicker.rosterPage = 0; return nil })
		},
	}
	c.single.paginate = c.navigation.Paginate
	if c.store != nil {
		c.single.data = singleAssessmentData{
			Draft:         c.store.Draft,
			GetAttempt:    c.store.GetAttempt,
			GetEnrollment: c.store.GetEnrollment,
			ListAttempts:  c.store.ListAttempts,
			Progress:      c.store.Progress,
			Skill:         c.store.Skill,
		}
	}
	c.single.actions = singleAssessmentActions{
		DiscardAttempt:                c.DiscardAttempt,
		RemoveLatestSuccess:           c.RemoveLatestSuccess,
		SaveAttempt:                   c.SaveAttempt,
		toggleLargeCorrectionSubskill: c.toggleLargeCorrectionSubskill,
	}
	c.single.small = &small.View{Actions: actions}
	c.single.shared = &shared.View{Actions: actions}
	c.single.large = &large.View{Actions: actions, Shared: c.single.shared,
		Pages: large.Pages{Roster: &c.single.rosterPage}, SwitchPerson: c.switchLargeAssessmentPerson,
	}
	if c.store != nil {
		c.class.data = classAssessmentData{
			Draft: c.store.Draft,
			Skill: c.store.Skill,
		}
	}
	c.class.actions = classAssessmentActions{
		classRailActions: c.classRailActions,
	}
	c.class.small = &small.View{Actions: actions}
	c.class.large = &large.View{Actions: actions,
		Pages: large.Pages{Class: &c.class.rosterPage}, ClassRoster: func() []ui.Item { return c.class.rosterRows(c.navigation.Frame().Page) }, SelectClassPerson: c.selectLargeClassPerson,
	}
	c.multi.paginate = c.navigation.Paginate
	if c.store != nil {
		c.multi.data = multiAssessmentData{
			Draft: c.store.Draft,
		}
	}
	c.multi.actions = multiAssessmentActions{
		DiscardMultiStudent: c.DiscardMultiStudent,
		EnterSingleMode:     c.EnterSingleMode,
		SaveMultiStudent:    c.SaveMultiStudent,
		changeMultiCount:    c.changeMultiCount,
		multiRailActions:    c.multiRailActions,
		skillModeSwitch:     c.skillModeSwitch,
	}
	c.multi.small = &small.View{Actions: actions}
	c.multi.shared = &shared.View{Actions: actions, AdjustGroups: c.adjustMultiGroupCount, PageGroup: c.pageMultiGroup, PassStudent: c.PassMultiStudent, EditStudent: c.EditMultiStudent}
	c.multi.large = &large.View{Actions: actions, Shared: c.multi.shared,
		Pages: large.Pages{Group: &c.multi.groupPage, Member: &c.multi.memberPage, MultiRoster: &c.multi.rosterPage, Table: &c.multi.tablePage}, MultiRows: c.multi.largeMultiRows, SelectGroup: c.selectLargeMultiGroup,
	}
	c.timed.paginate = c.navigation.Paginate
	if c.store != nil {
		c.timed.data = timedHeatData{
			Progress: c.store.Progress,
			Skill:    c.store.Skill,
		}
	}
	c.timed.actions = timedHeatActions{
		ConfirmTimedLaps:      c.ConfirmTimedLaps,
		ConfirmTimedSelection: c.ConfirmTimedSelection,
		DiscardTimedHeat:      c.DiscardTimedHeat,
		DiscardTimedSelection: c.DiscardTimedSelection,
		Refresh:               c.navigation.Refresh,
		RemoveLatestSuccess:   c.RemoveLatestSuccess,
		ResetTimedHeat:        c.ResetTimedHeat,
		SaveTimedHeat:         c.SaveTimedHeat,
		ToggleTimedTimer:      c.ToggleTimedTimer,
	}
	c.timed.small = &small.View{Actions: actions}
	c.timed.large = &large.View{Actions: actions,
		Pages:          large.Pages{Roster: &c.timed.rosterPage, Heat: &c.timed.heatPage},
		LapsInput:      func() string { return c.timed.lapsInput },
		SelectionCount: func() int { return len(c.timed.selection) },
		AddDigit: func(digit string) {
			_ = c.navigation.Run(func() error {
				if len(c.timed.lapsInput) < 3 {
					c.timed.lapsInput += digit
				}
				return nil
			})
		},
		ClearLaps:        func() { _ = c.navigation.Run(func() error { c.timed.lapsInput = ""; return nil }) },
		StartHeat:        c.startLargeTimedHeat,
		SelectedSwimmers: c.largeSelectedSwimmers,
	}
}

func (c *Component) selectLargeClassPerson(id string) {
	_ = c.navigation.Run(func() error {
		for i, student := range c.class.attempts {
			if student.PersonID == id {
				c.navigation.History.Frames[len(c.navigation.History.Frames)-1].Page = i
				break
			}
		}
		return nil
	})
}

func (c *Component) selectLargeMultiGroup(id string) {
	var index int
	if _, err := fmt.Sscan(id, &index); err != nil {
		return
	}
	_ = c.navigation.Run(func() error { c.multi.session.AssignmentGroup = index; c.multi.memberPage = 0; return nil })
}

func (c *Component) largeSelectedSwimmers() []ui.Item {
	var swimmers []ui.Item
	for _, id := range c.timed.selection {
		student, err := c.store.GetEnrollment(id)
		if err != nil {
			continue
		}
		swimmers = append(swimmers, ui.Item{ID: id, Label: student.Name, Selected: true, Fill: ui.Green})
	}
	return swimmers
}
