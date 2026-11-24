package curriculum

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestBundledCatalogIsReportCardCatalog(t *testing.T) {
	c := Bundled()
	if !strings.HasPrefix(c.Version, "reportcards-") || c.Version == BundledPrevious().Version {
		t.Fatal(c.Version)
	}
	if len(c.Levels) != 23 {
		t.Fatal(len(c.Levels))
	}
	ids := map[string]bool{}
	skills, subskills := 0, 0
	for _, l := range c.Levels {
		if l.ID == "SplashPrivate" {
			if l.Kind != "level-reference" || len(l.Skills) != 0 {
				t.Fatal(l)
			}
			continue
		}
		if len(l.Skills) == 0 {
			t.Fatal(l.ID)
		}
		for _, s := range l.Skills {
			skills++
			if ids[s.ID] || len(s.Subskills) == 0 || strings.TrimSpace(s.Name) == "" {
				t.Fatalf("invalid skill %s", s.ID)
			}
			ids[s.ID] = true
			for _, sub := range s.Subskills {
				subskills++
				if ids[sub.ID] || !strings.HasPrefix(sub.ID, s.ID+":") || strings.TrimSpace(sub.Text) == "" {
					t.Fatalf("invalid subskill %s of %s", sub.ID, s.ID)
				}
				ids[sub.ID] = true
			}
		}
	}
	if skills != 363 || subskills != 963 {
		t.Fatalf("%d skills, %d subskills", skills, subskills)
	}
	for _, id := range []string{"Splash7:10", "Splash8:9", "Splash9:11"} {
		if _, err := c.Skill(id); err != nil {
			t.Errorf("timed benchmark skill lost: %s", id)
		}
		if !IsTimedSkill(id) {
			t.Errorf("%s is not treated as timed", id)
		}
	}
}

// Levels with a report card take their hash from the transcription file the
// catalog was generated from; the two uncarded levels are copied unchanged.
func TestCatalogLevelsMatchReportCardFiles(t *testing.T) {
	for _, l := range Bundled().Levels {
		if !strings.HasPrefix(l.Source, "reportcards/") {
			if l.ID != "SplashFitness" && l.ID != "SplashPrivate" {
				t.Errorf("%s has no report card source", l.ID)
			}
			continue
		}
		data, err := os.ReadFile(l.Source)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != l.SourceSHA256 {
			t.Errorf("%s: catalog was not generated from the current %s", l.ID, l.Source)
		}
		var card struct {
			Level  string `json:"level"`
			Skills []struct {
				ID        string `json:"id"`
				Subskills []struct {
					ID string `json:"id"`
				} `json:"subskills"`
			} `json:"skills"`
		}
		if err := json.Unmarshal(data, &card); err != nil {
			t.Fatal(err)
		}
		if card.Level != l.ID || len(card.Skills) != len(l.Skills) {
			t.Fatalf("%s: card and catalog differ", l.ID)
		}
		for i, s := range card.Skills {
			if l.Skills[i].ID != s.ID || len(l.Skills[i].Subskills) != len(s.Subskills) {
				t.Errorf("%s: skill %s differs from its card", l.ID, s.ID)
			}
		}
	}
}

// Every old subskill is either carried into a new subskill or listed as
// dropped by its level's report card transcription.
func TestEveryOldSubskillIsMappedOrDropped(t *testing.T) {
	m, previous := BundledMigration(), BundledPrevious()
	if m.From != previous.Version || m.To != Bundled().Version {
		t.Fatalf("migration %s -> %s does not match bundled catalogs", m.From, m.To)
	}
	carried := map[string]bool{}
	for _, origin := range m.Subskills {
		for _, id := range origin.From {
			carried[id] = true
		}
	}
	dropped := map[string]bool{}
	for _, l := range Bundled().Levels {
		if !strings.HasPrefix(l.Source, "reportcards/") {
			continue
		}
		data, err := os.ReadFile(l.Source)
		if err != nil {
			t.Fatal(err)
		}
		var card struct {
			Dropped []struct {
				OldID string `json:"oldId"`
			} `json:"dropped"`
		}
		if err := json.Unmarshal(data, &card); err != nil {
			t.Fatal(err)
		}
		for _, d := range card.Dropped {
			dropped[d.OldID] = true
		}
	}
	for _, l := range previous.Levels {
		for _, s := range l.Skills {
			if len(m.NewSkills(s.ID)) == 0 {
				t.Errorf("old skill %s feeds no new skill", s.ID)
			}
			for _, sub := range s.Subskills {
				if !carried[sub.ID] && !dropped[sub.ID] {
					t.Errorf("old subskill %s is neither mapped nor dropped", sub.ID)
				}
				if carried[sub.ID] && dropped[sub.ID] {
					t.Errorf("old subskill %s is both mapped and dropped", sub.ID)
				}
			}
		}
	}
	for id := range dropped {
		if carried[id] {
			t.Errorf("dropped %s is also mapped", id)
		}
	}
}

func TestMigrationReferencesExistingIds(t *testing.T) {
	m, c, previous := BundledMigration(), Bundled(), BundledPrevious()
	oldSkills, oldSubs := map[string]bool{}, map[string]bool{}
	for _, l := range previous.Levels {
		for _, s := range l.Skills {
			oldSkills[s.ID] = true
			for _, x := range s.Subskills {
				oldSubs[x.ID] = true
			}
		}
	}
	newSubs := map[string]bool{}
	for _, l := range c.Levels {
		for _, s := range l.Skills {
			if _, ok := m.Skills[s.ID]; !ok {
				t.Errorf("new skill %s missing from migration", s.ID)
			}
			for _, o := range m.Skills[s.ID] {
				if !oldSkills[o] {
					t.Errorf("%s maps from unknown old skill %s", s.ID, o)
				}
			}
			for _, x := range s.Subskills {
				newSubs[x.ID] = true
				origin, ok := m.Subskills[x.ID]
				if !ok {
					t.Errorf("new subskill %s missing from migration", x.ID)
				}
				for _, o := range origin.From {
					if !oldSubs[o] {
						t.Errorf("%s maps from unknown old subskill %s", x.ID, o)
					}
				}
			}
		}
	}
	for id := range m.Subskills {
		if !newSubs[id] {
			t.Errorf("migration lists unknown new subskill %s", id)
		}
	}
}

func TestPlannedSkillsResolveSplitSkillsThroughMigration(t *testing.T) {
	c := Bundled()
	for level, split := range map[string]map[string][]string{
		"LittleSplash2": {"LittleSplash2:6": {"LittleSplash2:6-front-to-back", "LittleSplash2:6-back-to-front"}},
		"LittleSplash3": {"LittleSplash3:8": {"LittleSplash3:8-front-to-back", "LittleSplash3:8-back-to-front"}},
		"Splash1":       {"Splash1:7": {"Splash1:7-front-to-back", "Splash1:7-back-to-front"}},
		"TeenAdult1": {
			"TeenAdult1:11": {"TeenAdult1:11-front-to-back", "TeenAdult1:11-back-to-front"},
			"TeenAdult1:17": {"TeenAdult1:17-swim-with-a-buddy", "TeenAdult1:17-wear-a-lifejacket", "TeenAdult1:17-check-the-ice", "TeenAdult1:17-swim-to-survive", "TeenAdult1:17-within-arms-reach"},
		},
	} {
		l, _ := c.Level(level)
		for old, want := range split {
			got := l.PlannedSkills(old)
			if len(got) != len(want) {
				t.Fatalf("%s resolved to %d skills", old, len(got))
			}
			have := map[string]bool{}
			for _, s := range got {
				have[s.ID] = true
			}
			for _, id := range want {
				if !have[id] {
					t.Errorf("%s does not map to %s", old, id)
				}
			}
		}
	}
	l, _ := c.Level("Splash1")
	if got := l.PlannedSkills("Splash1:2"); len(got) != 1 || got[0].ID != "Splash1:2" {
		t.Fatalf("unsplit skill: %+v", got)
	}
}

func TestCompactTitlesCoverSplitSkills(t *testing.T) {
	for id := range compactTitles {
		_, err := Bundled().Skill(id)
		if _, oldErr := BundledPrevious().Skill(id); err != nil && oldErr != nil {
			t.Errorf("stale compact title %s", id)
		}
	}
	for _, id := range []string{"Splash1:7-front-to-back", "Splash1:7-back-to-front", "TeenAdult1:17-check-the-ice"} {
		s, _ := Bundled().Skill(id)
		if CompactTitle(s) == s.Name {
			t.Errorf("%s has no compact title", id)
		}
	}
}
