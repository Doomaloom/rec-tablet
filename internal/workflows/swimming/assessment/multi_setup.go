package assessment

import (
	"fmt"

	"rectablet/internal/shared/ui"
	"rectablet/internal/shell"
	"rectablet/internal/workflows/swimming/navigation"

	"fyne.io/fyne/v2/theme"
)

func balancedMultiSizes(students, groups int) []int {
	if students < 1 || groups < 1 || groups > students {
		return nil
	}
	sizes := make([]int, groups)
	for i := range sizes {
		sizes[i] = students / groups
		if i < students%groups {
			sizes[i]++
		}
	}
	return sizes
}

func multiSplitLabel(students, groups int) string {
	sizes := balancedMultiSizes(students, groups)
	if len(sizes) == 0 {
		return ""
	}
	groupLabel := func(n, size int) string {
		noun := "groups"
		if n == 1 {
			noun = "group"
		}
		return fmt.Sprintf("%d %s of %d", n, noun, size)
	}
	larger := students % groups
	if larger == 0 {
		return groupLabel(groups, sizes[0])
	}
	return groupLabel(larger, sizes[0]) + ", " + groupLabel(groups-larger, sizes[len(sizes)-1])
}

func (c *Controller) buildMultiGroupCountState() (navigation.State, error) {
	state := c.InitialState()
	if c.multi == nil {
		return state, fmt.Errorf("group setup is no longer available")
	}
	state.Title = "Number of Groups"
	state.PrimaryInfo = fmt.Sprintf("%s · %s · %d students present", c.multi.CourseInfo, c.multi.Skill.Name, len(c.multi.Roster))
	state.HeaderSwitch = c.skillModeSwitch(c.multi.Skill, true)
	state.Empty = "No students marked present. Take attendance first."
	state.BackRailAction = &shell.RailAction{Label: "Back", Run: c.EnterSingleMode, Icon: theme.NavigateBackIcon(), Fill: ui.Rail, Ink: ui.Blue}
	state.HideFooter = true
	state.CanSubmit = len(c.multi.Roster) > 0
	state.Progress = multiSplitLabel(len(c.multi.Roster), c.multi.ProposedGroupCount)
	state.ProgressInBody = true
	state.ComposeBody = c.composeMultiGroupCountBody(state, c.multi.ProposedGroupCount, len(c.multi.Roster))
	return state, nil
}

func (c *Controller) adjustMultiGroupCount(direction int) error {
	if c.Frame().Screen != navigation.MultiGroupCount || c.multi == nil || len(c.multi.Roster) == 0 {
		return nil
	}
	return c.Run(func() error {
		c.multi.ProposedGroupCount = max(1, min(c.multi.ProposedGroupCount+direction, len(c.multi.Roster)))
		return nil
	})
}

func (c *Controller) chooseMultiGroupCount() error {
	if c.Frame().Screen != navigation.MultiGroupCount || c.multi == nil {
		return nil
	}
	count := c.multi.ProposedGroupCount
	sizes := balancedMultiSizes(len(c.multi.Roster), count)
	if len(sizes) == 0 {
		return nil
	}
	if count != c.multi.GroupCount {
		c.multi.Groups = make([]multiGroup, count)
		for i, size := range sizes {
			c.multi.Groups[i] = multiGroup{Number: i + 1, Capacity: size}
		}
		c.multi.GroupCount = count
	}
	c.multi.AssignmentGroup = 0
	c.Push(navigation.MultiSelection, nil)
	return nil
}

func (c *Controller) changeMultiCount() error {
	if c.Frame().Screen != navigation.MultiSelection {
		return nil
	}
	return c.Run(func() error { c.multi.ProposedGroupCount = c.multi.GroupCount; c.Pop(); return nil })
}

func (c *Controller) pageMultiGroup(direction int) error {
	if c.Frame().Screen != navigation.MultiSelection || c.multi == nil {
		return nil
	}
	return c.Run(func() error {
		c.multi.AssignmentGroup = max(0, min(c.multi.AssignmentGroup+direction, len(c.multi.Groups)-1))
		c.Frames[len(c.Frames)-1].Page = 0
		return nil
	})
}

func (c *Controller) multiReady() bool {
	if c.multi == nil || len(c.multi.Groups) == 0 {
		return false
	}
	eligible := make(map[string]bool, len(c.multi.Roster))
	for _, student := range c.multi.Roster {
		eligible[student.ID] = true
	}
	seen := make(map[string]bool, len(eligible))
	for _, group := range c.multi.Groups {
		if len(group.Members) != group.Capacity {
			return false
		}
		for _, person := range group.Members {
			if !eligible[person] || seen[person] {
				return false
			}
			seen[person] = true
		}
	}
	return len(seen) == len(eligible)
}

func (c *Controller) toggleMultiMember(person string) error {
	if c.multi == nil || len(c.multi.Groups) == 0 {
		return nil
	}
	eligible := false
	for _, student := range c.multi.Roster {
		if student.ID == person {
			eligible = true
			break
		}
	}
	if !eligible {
		return nil
	}
	group := &c.multi.Groups[c.multi.AssignmentGroup]
	if containsString(group.Members, person) {
		group.Members = removeString(group.Members, person)
		return nil
	}
	if len(group.Members) >= group.Capacity {
		return nil
	}
	for _, other := range c.multi.Groups {
		if containsString(other.Members, person) {
			return nil
		}
	}
	group.Members = append(group.Members, person)
	return nil
}

func (c *Controller) buildMultiSelectionState() (navigation.State, error) {
	state := c.InitialState()
	if c.multi == nil || len(c.multi.Groups) == 0 {
		return state, fmt.Errorf("choose the number of groups first")
	}
	group := c.multi.Groups[c.multi.AssignmentGroup]
	state.Title = fmt.Sprintf("Assign Group %d of %d", group.Number, c.multi.GroupCount)
	selectionInfo := fmt.Sprintf("%d of %d selected · %s", len(group.Members), group.Capacity, c.multi.Skill.Name)
	state.PrimaryInfo = c.multi.CourseInfo + " · " + selectionInfo
	state.HeaderActions = []shell.HeaderAction{{Label: "Groups", Run: c.changeMultiCount}}
	state.HomeRailAction = discardRailAction(c.EnterSingleMode)
	state.HomeRailAction.Label = "Cancel setup"
	state.BackRailAction = &shell.RailAction{Label: "Group count", Run: c.changeMultiCount, Icon: theme.NavigateBackIcon(), Fill: ui.Rail, Ink: ui.Blue}
	state.FooterAction, state.CanSubmit = "Start marking", c.multiReady()
	assigned := 0
	for _, g := range c.multi.Groups {
		assigned += len(g.Members)
	}
	state.Progress = fmt.Sprintf("%d/%d assigned", assigned, len(c.multi.Roster))
	var rows []navigation.Row
	for _, student := range c.multi.Roster {
		row := navigation.Row{ID: student.ID, Label: student.Name + " · Unassigned"}
		for i, other := range c.multi.Groups {
			if containsString(other.Members, student.ID) {
				row.Label = fmt.Sprintf("%s · Group %d", student.Name, other.Number)
				row.Selected = i == c.multi.AssignmentGroup
				row.Disabled = !row.Selected
				if row.Selected {
					row.Fill = ui.Green
				}
			}
		}
		if !row.Selected && len(group.Members) == group.Capacity {
			row.Disabled = true
		}
		rows = append(rows, row)
	}
	c.Paginate(&state, rows)
	state.ComposeBody = c.composeMultiSelectionBody(state, c.multi.AssignmentGroup, len(c.multi.Groups))
	return state, nil
}
