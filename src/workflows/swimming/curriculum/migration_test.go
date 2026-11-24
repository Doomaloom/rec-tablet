package curriculum

import "testing"

func testMigration(t *testing.T) (Migration, Catalog) {
	t.Helper()
	return BundledMigration(), Bundled()
}

func results(pairs ...string) map[string]string {
	out := map[string]string{}
	for i := 0; i+1 < len(pairs); i += 2 {
		out[pairs[i]] = pairs[i+1]
	}
	return out
}

func TestTranslateSplitSkillFeedsBothNewSkills(t *testing.T) {
	m, c := testMigration(t)
	old := OldAttempt{ID: "a1", SkillID: "Splash1:7", Success: true, Results: results("Splash1:7:1", "successful", "Splash1:7:2", "successful", "Splash1:7:3", "successful")}
	got := m.Translate(c, old)
	if len(got) != 2 {
		t.Fatalf("%+v", got)
	}
	ids := map[string]bool{}
	for _, tr := range got {
		ids[tr.SkillID] = true
		if !tr.Success || tr.SourceID != "a1" || len(tr.Results) != 3 {
			t.Errorf("%+v", tr)
		}
	}
	if !ids["Splash1:7-front-to-back"] || !ids["Splash1:7-back-to-front"] {
		t.Fatal(ids)
	}
	// A miss on one old criterion carries into both halves because each half
	// inherits all of the old criteria.
	old.Success = false
	old.Results["Splash1:7:2"] = "not_yet"
	for _, tr := range m.Translate(c, old) {
		if tr.Success || tr.Results[tr.SkillID+":2"] != "not_yet" || tr.Results[tr.SkillID+":1"] != "successful" {
			t.Errorf("%+v", tr)
		}
	}
}

func TestTranslateFromSkillSuccessBox(t *testing.T) {
	m, c := testMigration(t)
	skill, err := c.Skill("LittleSplash3:8-front-to-back")
	if err != nil {
		t.Fatal(err)
	}
	var boxID string
	for _, sub := range skill.Subskills {
		if o := m.Subskills[sub.ID]; o.FromSkillSuccess {
			boxID = sub.ID
		}
	}
	if boxID == "" {
		t.Fatal("expected a fromSkillSuccess box")
	}
	all := OldAttempt{ID: "a", SkillID: "LittleSplash3:8", Success: true, Results: map[string]string{}}
	for _, sub := range skill.Subskills {
		for _, id := range m.Subskills[sub.ID].From {
			all.Results[id] = "successful"
		}
	}
	tr, ok := m.TranslateTo(skill, all)
	if !ok || !tr.Success || tr.Results[boxID] != "successful" {
		t.Fatalf("%+v", tr)
	}
	all.Success = false
	tr, _ = m.TranslateTo(skill, all)
	if tr.Success || tr.Results[boxID] != "not_yet" {
		t.Fatalf("box without old origin must follow the old attempt's success: %+v", tr)
	}
}

func TestTranslateMultipleOldIdsAndMissingResults(t *testing.T) {
	m := Migration{Skills: map[string][]string{"N:1": {"O:1"}}, Subskills: map[string]SubskillOrigin{
		"N:1:1": {From: []string{"O:1:1", "O:1:2"}},
		"N:1:2": {From: []string{"O:1:3"}},
		"N:1:3": {},
	}}
	skill := Skill{ID: "N:1", Subskills: []Subskill{{ID: "N:1:1"}, {ID: "N:1:2"}, {ID: "N:1:3"}}}
	tr, ok := m.TranslateTo(skill, OldAttempt{SkillID: "O:1", Results: results("O:1:1", "successful", "O:1:2", "not_yet", "O:1:3", "successful")})
	if !ok || tr.Results["N:1:1"] != "not_yet" || tr.Results["N:1:2"] != "successful" {
		t.Fatalf("%+v", tr)
	}
	// Only one of the merged criteria present and successful still counts.
	tr, _ = m.TranslateTo(skill, OldAttempt{SkillID: "O:1", Results: results("O:1:2", "successful")})
	if tr.Results["N:1:1"] != "successful" || tr.Results["N:1:2"] != "not_yet" {
		t.Fatalf("%+v", tr)
	}
	// A subskill with neither origin nor fromSkillSuccess can never be earned.
	if tr.Results["N:1:3"] != "not_yet" || tr.Success {
		t.Fatalf("%+v", tr)
	}
	if _, ok := m.TranslateTo(skill, OldAttempt{SkillID: "other"}); ok {
		t.Fatal("unrelated skill translated")
	}
}

func TestTranslateTimedSkillNeedsOldSuccess(t *testing.T) {
	m, c := testMigration(t)
	skill, err := c.Skill("Splash7:10")
	if err != nil {
		t.Fatal(err)
	}
	old := OldAttempt{ID: "t", SkillID: "Splash7:10", Results: map[string]string{}}
	for _, sub := range skill.Subskills {
		for _, id := range m.Subskills[sub.ID].From {
			old.Results[id] = "successful"
		}
	}
	tr, ok := m.TranslateTo(skill, old)
	if !ok || tr.Success {
		t.Fatalf("slow timed attempt must stay unsuccessful: %+v", tr)
	}
	old.Success = true
	if tr, _ = m.TranslateTo(skill, old); !tr.Success {
		t.Fatalf("%+v", tr)
	}
}

func TestEveryOldSkillTranslates(t *testing.T) {
	m, c := testMigration(t)
	for _, l := range BundledPrevious().Levels {
		for _, s := range l.Skills {
			a := OldAttempt{ID: "x", SkillID: s.ID, Success: true, Results: map[string]string{}}
			for _, sub := range s.Subskills {
				a.Results[sub.ID] = "successful"
			}
			got := m.Translate(c, a)
			if len(got) == 0 {
				t.Errorf("%s translates to nothing", s.ID)
			}
			for _, tr := range got {
				if !tr.Success {
					t.Errorf("fully successful old %s is not successful as %s: %+v", s.ID, tr.SkillID, tr.Results)
				}
			}
		}
	}
}

func TestNewSkillsInverse(t *testing.T) {
	m := BundledMigration()
	if got := m.NewSkills("TeenAdult1:17"); len(got) != 5 {
		t.Fatal(got)
	}
	if got := m.NewSkills("Splash1:2"); len(got) != 1 || got[0] != "Splash1:2" {
		t.Fatal(got)
	}
	if len(m.NewSkills("nope")) != 0 {
		t.Fatal("unknown skill")
	}
}
