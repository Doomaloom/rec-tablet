package curriculum

import (
	_ "embed"
	"encoding/json"
)

// Explicit aliases bridge the website planner's grouped skills to tablet assessment components.
//
//go:embed planned_skill_ids.json
var plannedIDsJSON []byte
var plannedIDs = func() map[string][]string {
	var aliases map[string][]string
	if err := json.Unmarshal(plannedIDsJSON, &aliases); err != nil {
		panic(err)
	}
	return aliases
}()

func (l Level) PlannedSkills(id string) []Skill {
	ids, ok := plannedIDs[id]
	if !ok {
		ids = []string{id}
	}
	out := []Skill{}
	for _, id := range ids {
		for _, skill := range l.Skills {
			if skill.ID == id {
				out = append(out, skill)
				break
			}
		}
	}
	return out
}
