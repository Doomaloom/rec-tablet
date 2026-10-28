package assessment

import (
	"rectablet/next/navigation"
)

func (c *Component) PrepareEdit(id string) (navigation.Edit, bool) {
	switch c.navigation.Frame().Screen {
	case navigation.Assessment:
		return c.single.PrepareEdit(id)
	case navigation.MultiStudent:
		return c.multi.PrepareEdit(id)
	case navigation.ClassAssessment:
		return c.class.PrepareEdit(id)
	}
	return navigation.Edit{}, false
}
