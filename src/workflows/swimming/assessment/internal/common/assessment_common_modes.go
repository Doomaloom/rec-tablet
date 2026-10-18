package common

import (
	"strings"

	"rectablet/src/workflows/swimming/curriculum"
)

// Dense checklists and stroke evaluations use a full-width, individual checklist.
func AllowsMultiMode(skill curriculum.Skill) bool {
	name := strings.ToLower(skill.Name)
	return len(skill.Subskills) <= 6 && !strings.Contains(name, "crawl") &&
		!strings.Contains(name, "stroke") && !strings.Contains(name, "butterfly")
}
