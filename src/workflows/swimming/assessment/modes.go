package assessment

import (
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/assessment/internal/common"
	"rectablet/src/workflows/swimming/curriculum"
)

func (c *Component) skillModeSwitch(skill curriculum.Skill, multi bool) []shell.HeaderSwitchOption {
	options := []shell.HeaderSwitchOption{{Label: "Single", Selected: !multi, Run: c.EnterSingleMode}}
	if common.AllowsMultiMode(skill) {
		options = append(options, shell.HeaderSwitchOption{Label: "Multi", Selected: multi, Run: c.EnterMultiMode})
	}
	return append(options, shell.HeaderSwitchOption{Label: "Class", Run: c.EnterClassMode})
}
