package assessment

import (
	"rectablet/next/contracts"
	"rectablet/next/features/assessment/internal/common"
	"rectablet/src/workflows/swimming/curriculum"
)

func (c *Component) skillModeSwitch(skill curriculum.Skill, multi bool) []contracts.HeaderSwitchOption {
	options := []contracts.HeaderSwitchOption{{Label: "Single", Selected: !multi, Run: c.EnterSingleMode}}
	if common.AllowsMultiMode(skill) {
		options = append(options, contracts.HeaderSwitchOption{Label: "Multi", Selected: multi, Run: c.EnterMultiMode})
	}
	return append(options, contracts.HeaderSwitchOption{Label: "Class", Run: c.EnterClassMode})
}
