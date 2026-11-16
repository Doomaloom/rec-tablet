package curriculum

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func templateDirectory(t *testing.T) string {
	t.Helper()
	dir := filepath.Join("..", "..", "..", "..", "..", "COB-Aquatics-SPA", "frontend/src/shared/attendance-print/templates")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Skip("sister repository unavailable")
	} else if err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestBundledCatalog(t *testing.T) {
	c := Bundled()
	if len(c.Levels) != 23 {
		t.Fatal(len(c.Levels))
	}
	ids := map[string]bool{}
	skills, criteria, unmapped := 0, 0, 0
	for _, l := range c.Levels {
		if len(l.SourceSHA256) != 64 {
			t.Fatal(l.ID)
		}
		if l.ID == "SplashPrivate" {
			if l.Kind != "level-reference" || len(l.References) == 0 || len(l.Skills) != 0 {
				t.Fatal(l)
			}
			continue
		}
		if len(l.Skills) == 0 {
			t.Fatal(l.ID)
		}
		unmapped += len(l.UnmappedSubskills)
		for _, s := range l.Skills {
			skills++
			if ids[s.ID] || s.SourceSkillID == "" || s.SourceName == "" {
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
	if skills != 353 || criteria != 1027 || unmapped != 2 {
		t.Fatalf("%d skills, %d criteria, %d unmatched source criteria", skills, criteria, unmapped)
	}
}

func TestExtractionParity(t *testing.T) {
	actual, err := ExtractDirectory(templateDirectory(t))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actual, Bundled()) {
		t.Fatal("bundled catalog differs from front-page skills and back-page subskills; regenerate it with extract-curriculum")
	}
}

func TestFrontPageStrokeNamesAndNumbers(t *testing.T) {
	c := Bundled()
	for _, tc := range []struct{ id, number, name string }{
		{"Splash7:8-front-crawl", "8a", "Front Crawl - 50 m or yd"},
		{"Splash7:8-back-crawl", "8b", "Back Crawl - 50 m or yd"},
		{"Splash7:8-breaststroke", "8c", "Breaststroke - 50 m or yd"},
		{"Splash8:7-front-crawl", "7a", "Front Crawl - 75 m. or yd."},
		{"Splash8:7-back-crawl", "7b", "Back Crawl - 75 m. or yd."},
		{"Splash8:7-breaststroke", "7c", "Breaststroke - 75 m. or yd."},
		// Splash9 actually repeats 9 for all three front-page columns.
		{"Splash9:9-front-crawl", "9", "Front Crawl, 100 m or y."},
		{"Splash9:9-back-crawl", "9", "Back Crawl, 100 m or y."},
		{"Splash9:9-breaststroke", "9", "Breaststroke, 100 m or y."},
	} {
		skill, err := c.Skill(tc.id)
		if err != nil {
			t.Fatal(err)
		}
		if skill.Number != tc.number || skill.Name != tc.name {
			t.Errorf("%s: got %s. %s", tc.id, skill.Number, skill.Name)
		}
		for _, sub := range skill.Subskills {
			if len(sub.Group) > 0 && !matchesComponent(sub.Group[0], componentFor(skill.Name)) {
				t.Errorf("%s contains another stroke's criterion: %+v", tc.id, sub)
			}
		}
	}
}

func TestFrontPageUnitsAndSharedCriteria(t *testing.T) {
	c := Bundled()
	for _, id := range []string{"Splash2A:8-front-crawl", "Splash2A:8-back-crawl"} {
		s, err := c.Skill(id)
		if err != nil {
			t.Fatal(err)
		}
		if s.Subskills[0].Text != "Minimum distance completed (for both strokes)" {
			t.Fatal("lost shared distance criterion", id)
		}
	}
	for _, id := range []string{"LittleSplash1:5a", "LittleSplash1:5b", "Splash1:11a", "Splash1:11b", "Splash1:11c", "Splash1:11d", "Splash2B:2a", "Splash2B:2b", "Splash2B:8c", "Splash7:4a", "Splash7:4b", "Splash5:12a", "Splash5:12b"} {
		if _, err := c.Skill(id); err != nil {
			t.Fatal(err)
		}
	}
	// This is one OR skill on the front page, so it must remain combined.
	s, err := c.Skill("TeenAdult3:7")
	if err != nil || len(s.Subskills) != 8 {
		t.Fatal(s, err)
	}
}

func TestExtractionExceptions(t *testing.T) {
	c := Bundled()
	s, _ := c.Skill("Splash1:3")
	if len(s.Subskills) != 5 {
		t.Fatal(s)
	}
	for _, id := range []string{"Splash8:12a", "Splash8:12b", "Splash8:12c"} {
		if _, err := c.Skill(id); err != nil {
			t.Fatal(err)
		}
	}
	s, _ = c.Skill("Splash7:16")
	if !strings.Contains(strings.ToLower(s.Name), "30 sec") || len(s.Subskills) != 1 {
		t.Fatal(s)
	}
	for _, id := range []string{"ParentandTot1:10a", "LittleSplash1:8a", "TeenAdult1:17"} {
		s, err := c.Skill(id)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(s.Subskills[0].Text, "Participation") {
			t.Fatal("lost shared participation criterion", id)
		}
	}
}

func TestSourceWordsPreserved(t *testing.T) {
	words := regexp.MustCompile(`[\pL\pN_]+`)
	for _, level := range Bundled().Levels {
		targets := map[string]map[string]bool{}
		lines := map[string][]string{}
		add := func(source, text string) {
			if targets[source] == nil {
				targets[source] = map[string]bool{}
			}
			for _, w := range words.FindAllString(strings.ToLower(text), -1) {
				targets[source][w] = true
			}
		}
		for _, s := range level.Skills {
			lines[s.SourceSkillID] = s.SourceLines
			for _, sub := range s.Subskills {
				add(s.SourceSkillID, sub.Text+" "+strings.Join(sub.Group, " "))
			}
		}
		for _, sub := range level.UnmappedSubskills {
			source := sub.ID[:strings.LastIndex(sub.ID, ":")]
			add(source, sub.Text+" "+strings.Join(sub.Group, " "))
		}
		for source, sourceLines := range lines {
			for _, w := range words.FindAllString(strings.ToLower(strings.Join(sourceLines, " ")), -1) {
				if !targets[source][w] {
					t.Errorf("%s lost %q", source, w)
				}
			}
		}
	}
}
