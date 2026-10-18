package assessment

import (
	"strings"

	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/curriculum"
)

// Dense checklists and stroke evaluations use a full-width, individual checklist.
func allowsMultiMode(skill curriculum.Skill) bool {
	name := strings.ToLower(skill.Name)
	return len(skill.Subskills) <= 6 && !strings.Contains(name, "crawl") &&
		!strings.Contains(name, "stroke") && !strings.Contains(name, "butterfly")
}

func (c *Controller) skillModeSwitch(skill curriculum.Skill, multi bool) []shell.HeaderSwitchOption {
	options := []shell.HeaderSwitchOption{{Label: "Single", Selected: !multi, Run: c.EnterSingleMode}}
	if allowsMultiMode(skill) {
		options = append(options, shell.HeaderSwitchOption{Label: "Multi", Selected: multi, Run: c.EnterMultiMode})
	}
	return append(options, shell.HeaderSwitchOption{Label: "Class", Run: c.EnterClassMode})
}
