package curriculum

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestBundledCatalog(t *testing.T) {
	c := Bundled()
	if len(c.Levels) != 23 {
		t.Fatal(len(c.Levels))
	}
	ids := map[string]bool{}
	skills, criteria := 0, 0
	for _, l := range c.Levels {
		if len(l.SourceSHA256) != 64 {
			t.Fatal(l.ID)
		}
		if l.ID == "SplashPrivate" {
			if l.Kind != "level-reference" || len(l.References) == 0 {
				t.Fatal(l)
			}
			continue
		}
		if len(l.Skills) == 0 {
			t.Fatal(l.ID)
		}
		for _, s := range l.Skills {
			skills++
			if ids[s.ID] {
				t.Fatal(s.ID)
			}
			ids[s.ID] = true
			if len(s.Subskills) == 0 {
				t.Fatal(s.ID)
			}
			for _, sub := range s.Subskills {
				criteria++
				if ids[sub.ID] || strings.TrimSpace(sub.Text) == "" {
					t.Fatal(sub)
				}
				ids[sub.ID] = true
			}
		}
	}
	if skills != 290 || criteria != 937 {
		t.Fatalf("%d skills, %d criteria", skills, criteria)
	}
}

var splitSkillBases = map[string][]string{
	"Splash1:10":    {"front-crawl-wearing-pfd-5-m", "back-crawl-wearing-pfd-5-m"},
	"Splash2A:8":    {"front-crawl", "back-crawl"},
	"Splash2B:8":    {"front-crawl", "back-crawl"},
	"Splash3:10":    {"front-crawl", "back-crawl"},
	"Splash4:7":     {"front-crawl", "back-crawl"},
	"Splash5:8":     {"front-crawl", "back-crawl"},
	"Splash6:8":     {"front-crawl", "back-crawl"},
	"Splash7:8":     {"front-crawl", "back-crawl", "breaststroke"},
	"Splash8:7":     {"front-crawl", "back-crawl", "breaststroke"},
	"Splash9:9":     {"front-crawl", "back-crawl", "breaststroke"},
	"TeenAdult1:12": {"front-glide", "back-glide", "side-glide"},
	"TeenAdult1:15": {"front-crawl", "back-crawl"},
	"TeenAdult2:14": {"front-crawl", "back-crawl"},
	"TeenAdult3:7":  {"eggbeater-kick", "scissor-kick"},
	"TeenAdult3:9":  {"front-crawl", "back-crawl"},
}

func splitOverrideSkillID(id string) bool {
	for base := range splitSkillBases {
		if id == base {
			return true
		}
		prefix := base + "-"
		if !strings.HasPrefix(id, prefix) {
			continue
		}
		// The source extractor already has a numeric duplicate-skill suffix
		// such as Splash2B:8-2; that remains part of the source catalog.
		suffix := strings.TrimPrefix(id, prefix)
		if suffix != "2" && suffix != "3" {
			return true
		}
	}
	return false
}

func TestSplitMultiSkills(t *testing.T) {
	catalog := Bundled()
	byID := map[string]Skill{}
	for _, level := range catalog.Levels {
		for _, skill := range level.Skills {
			byID[skill.ID] = skill
		}
	}
	for base, suffixes := range splitSkillBases {
		if _, ok := byID[base]; ok {
			t.Fatalf("combined skill %s was not replaced", base)
		}
		baseNumber := strings.TrimPrefix(base, strings.Split(base, ":")[0]+":")
		for _, suffix := range suffixes {
			id := base + "-" + suffix
			skill, ok := byID[id]
			if !ok {
				t.Fatalf("split skill %s is missing", id)
			}
			if skill.Number != baseNumber {
				t.Fatalf("split skill %s has number %q, want %q", id, skill.Number, baseNumber)
			}
			componentWords := strings.Fields(strings.ReplaceAll(suffix, "-", " "))
			if len(componentWords) > 2 {
				componentWords = componentWords[:2]
			}
			componentName := strings.Join(componentWords, " ")
			if !strings.Contains(strings.ToLower(skill.Name), componentName) {
				t.Fatalf("split skill %s does not identify its component: %q", id, skill.Name)
			}
		}
	}
	for _, id := range []string{"Splash2A:8-front-crawl", "Splash2A:8-back-crawl"} {
		skill := byID[id]
		found := false
		for _, subskill := range skill.Subskills {
			if subskill.Text == "Minimum distance completed (for both strokes)" {
				found = true
			}
		}
		if !found {
			t.Fatalf("shared criterion missing from %s", id)
		}
	}
	for _, id := range []string{"Splash7:13", "Splash9:10"} {
		if _, ok := byID[id]; !ok {
			t.Fatalf("single-topic skill %s was unexpectedly split", id)
		}
	}
}

func TestExtractionParity(t *testing.T) {
	dir := filepath.Join("..", "..", "..", "..", "..", "COB-Aquatics-SPA", "frontend/src/features/attendance-print/templates")
	if _, e := os.Stat(dir); e != nil {
		t.Skip("sister repository unavailable")
	}
	actual, e := ExtractDirectory(dir)
	if e != nil {
		t.Fatal(e)
	}
	expected := Bundled()
	for i, l := range actual.Levels {
		bundled := expected.Levels[i]
		if len(l.Skills) == 0 {
			if !reflect.DeepEqual(l, bundled) {
				t.Fatalf("%s differs from extracted source", l.ID)
			}
			continue
		}
		actualSkills := map[string]Skill{}
		bundledSkills := map[string]Skill{}
		for _, skill := range l.Skills {
			if !splitOverrideSkillID(skill.ID) {
				actualSkills[skill.ID] = skill
			}
		}
		for _, skill := range bundled.Skills {
			if !splitOverrideSkillID(skill.ID) {
				bundledSkills[skill.ID] = skill
			}
		}
		if !reflect.DeepEqual(actualSkills, bundledSkills) {
			a, _ := json.MarshalIndent(actualSkills, "", "  ")
			t.Fatalf("%s differs outside the intentional split override: %s", l.ID, string(a))
		}
	}
}
func TestExtractionExceptions(t *testing.T) {
	c := Bundled()
	s, _ := c.Skill("Splash1:3")
	if len(s.Subskills) != 5 {
		t.Fatal(s)
	}
	if _, e := c.Skill("Splash2B:8-2"); e != nil {
		t.Fatal(e)
	}
	for _, id := range []string{"Splash8:12a", "Splash8:12b", "Splash8:12c"} {
		if _, e := c.Skill(id); e != nil {
			t.Fatal(e)
		}
	}
	s, _ = c.Skill("Splash7:16")
	if !strings.Contains(s.Name, "30 sec.") || len(s.Subskills) != 1 {
		t.Fatal(s)
	}
	for _, id := range []string{"ParentandTot1:10", "LittleSplash1:8", "TeenAdult1:17"} {
		s, _ = c.Skill(id)
		found := false
		for _, sub := range s.Subskills {
			if strings.Contains(sub.Text, "Participation") {
				found = true
			}
		}
		if !found {
			t.Fatal(id)
		}
	}
}
func TestSourceWordsPreserved(t *testing.T) {
	words := regexp.MustCompile(`[\pL\pN_]+`)
	for _, l := range Bundled().Levels {
		for _, s := range l.Skills {
			if splitOverrideSkillID(s.ID) {
				continue
			}
			target := map[string]bool{}
			for _, c := range s.Subskills {
				for _, w := range words.FindAllString(strings.ToLower(c.Text+" "+strings.Join(c.Group, " ")), -1) {
					target[w] = true
				}
			}
			for _, w := range words.FindAllString(strings.ToLower(strings.Join(s.SourceLines, " ")), -1) {
				if !target[w] {
					t.Errorf("%s lost %q", s.ID, w)
				}
			}
		}
	}
}
