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

// PlannedSkills maps a website lesson-plan skill id to tablet skills. The
// website keeps the previous skill list, so an id absent from the explicit
// aliases resolves through the migration: an old skill that was split maps to
// every new skill it became. Unknown ids resolve to themselves.
func (l Level) PlannedSkills(id string) []Skill {
	ids, explicit := plannedIDs[id]
	if !explicit {
		ids = []string{id}
	}
	byID := map[string]Skill{}
	for _, skill := range l.Skills {
		byID[skill.ID] = skill
	}
	out := []Skill{}
	if explicit {
		for _, id := range ids {
			if skill, ok := byID[id]; ok {
				out = append(out, skill)
			}
		}
		return out
	}
	migrated := map[string]bool{}
	for _, nid := range BundledMigration().NewSkills(id) {
		migrated[nid] = true
	}
	for _, skill := range l.Skills {
		if migrated[skill.ID] {
			out = append(out, skill)
		}
	}
	if len(out) == 0 {
		if skill, ok := byID[id]; ok {
			out = append(out, skill)
		}
	}
	return out
}
