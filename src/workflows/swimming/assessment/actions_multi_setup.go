package assessment

import (
	"fmt"

	"rectablet/src/workflows/swimming/navigation"
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

func (c *Component) adjustMultiGroupCount(direction int) error {
	if c.navigation.Frame().Screen != navigation.MultiGroupCount || c.multi.session == nil || len(c.multi.session.Roster) == 0 {
		return nil
	}
	return c.navigation.Run(func() error {
		c.multi.session.ProposedGroupCount = max(1, min(c.multi.session.ProposedGroupCount+direction, len(c.multi.session.Roster)))
		return nil
	})
}

func (c *Component) chooseMultiGroupCount() error {
	if c.navigation.Frame().Screen != navigation.MultiGroupCount || c.multi.session == nil {
		return nil
	}
	count := c.multi.session.ProposedGroupCount
	sizes := balancedMultiSizes(len(c.multi.session.Roster), count)
	if len(sizes) == 0 {
		return nil
	}
	if count != c.multi.session.GroupCount {
		c.multi.session.Groups = make([]multiGroup, count)
		for i, size := range sizes {
			c.multi.session.Groups[i] = multiGroup{Number: i + 1, Capacity: size}
		}
		c.multi.session.GroupCount = count
	}
	c.multi.session.AssignmentGroup = 0
	c.navigation.Push(navigation.MultiSelection, nil)
	return nil
}

func (c *Component) changeMultiCount() error {
	if c.navigation.Frame().Screen != navigation.MultiSelection {
		return nil
	}
	return c.navigation.Run(func() error {
		c.multi.session.ProposedGroupCount = c.multi.session.GroupCount
		c.navigation.Pop()
		return nil
	})
}

func (c *Component) pageMultiGroup(direction int) error {
	if c.navigation.Frame().Screen != navigation.MultiSelection || c.multi.session == nil {
		return nil
	}
	return c.navigation.Run(func() error {
		c.multi.session.AssignmentGroup = max(0, min(c.multi.session.AssignmentGroup+direction, len(c.multi.session.Groups)-1))
		c.navigation.History.Frames[len(c.navigation.History.Frames)-1].Page = 0
		return nil
	})
}
