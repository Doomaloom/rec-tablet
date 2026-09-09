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
	if skills != 271 || criteria != 919 {
		t.Fatalf("%d skills, %d criteria", skills, criteria)
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
		if !reflect.DeepEqual(l, expected.Levels[i]) {
			a, _ := json.MarshalIndent(l, "", "  ")
			b, _ := json.MarshalIndent(expected.Levels[i], "", "  ")
			out := t.TempDir()
			_ = os.WriteFile(filepath.Join(out, "actual.json"), a, 0600)
			_ = os.WriteFile(filepath.Join(out, "expected.json"), b, 0600)
			t.Fatalf("%s differs; actual %s", l.ID, string(a))
		}
	}
	if actual.Version != expected.Version {
		t.Fatalf("version changed: %s != %s", actual.Version, expected.Version)
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
