package curriculum

import (
	"encoding/json"
	"os"
	"testing"
)

func TestPlannedSkillAliasesResolveWithinTheirLevel(t *testing.T) {
	catalog := Bundled()
	for source, ids := range plannedIDs {
		found := false
		for _, level := range catalog.Levels {
			skills := level.PlannedSkills(source)
			if len(skills) == len(ids) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("unresolved planned skill alias %s: %v", source, ids)
		}
	}
	level, _ := catalog.Level("Splash1")
	skills := level.PlannedSkills("Splash1:6")
	if len(skills) != 2 || skills[0].ID != "Splash1:6a" || skills[1].ID != "Splash1:6b" {
		t.Fatalf("float components: %+v", skills)
	}
	if len(level.PlannedSkills("Splash7:10")) != 0 || len(level.PlannedSkills("unknown")) != 0 {
		t.Fatal("resolved a skill outside its level")
	}
}

func TestWebsitePlannedSkillCoverage(t *testing.T) {
	data, err := os.ReadFile("../../../../../COB-Aquatics-SPA/frontend/src/features/instructor/LessonPlans/lessonSkills.json")
	if os.IsNotExist(err) {
		t.Skip("website source unavailable")
	}
	if err != nil {
		t.Fatal(err)
	}
	var levels []struct {
		ID     string `json:"id"`
		Skills []struct {
			ID string `json:"id"`
		} `json:"skills"`
	}
	if err = json.Unmarshal(data, &levels); err != nil {
		t.Fatal(err)
	}
	catalog := Bundled()
	for _, source := range levels {
		if len(source.Skills) == 0 {
			continue
		}
		level, ok := catalog.Level(source.ID)
		if !ok {
			t.Errorf("missing level %s", source.ID)
			continue
		}
		for _, skill := range source.Skills {
			if len(level.PlannedSkills(skill.ID)) == 0 {
				t.Errorf("unmapped website skill %s", skill.ID)
			}
		}
	}
}
