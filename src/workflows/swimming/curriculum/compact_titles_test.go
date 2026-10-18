package curriculum

import (
	"encoding/json"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"rectablet/src/shared/ui"
)

func TestCompactTitlesFitBundledSkills(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	app.Settings().SetTheme(ui.AppTheme{})
	catalog := Bundled()
	for _, level := range catalog.Levels {
		for _, skill := range level.Skills {
			label := skill.Number + ". " + CompactTitle(skill)
			// A 296px header half has 12px padding on each side.
			if width := fyne.MeasureText(label, 18, fyne.TextStyle{Bold: true}).Width; width > 272 {
				t.Errorf("%s: %.0fpx: %s", skill.ID, width, label)
			}
			if strings.Contains(skill.Name, " — ") && compactTitles[skill.ID] == "" {
				t.Errorf("split component missing compact title: %s", skill.ID)
			}
		}
	}
	for id, title := range compactTitles {
		if _, err := catalog.Skill(id); err != nil {
			t.Errorf("compact title references an unknown skill: %s", id)
		}
		if strings.TrimSpace(title) != title || title == "" || strings.ContainsAny(title, "\r\n") {
			t.Errorf("invalid compact title for %s: %q", id, title)
		}
	}
}

func TestCompactTitleFallback(t *testing.T) {
	skill := Skill{ID: "custom:1", Name: "Original skill title"}
	if CompactTitle(skill) != skill.Name {
		t.Fatal("unknown skills must retain their original title")
	}
}

func TestCompactTitlesPreserveCanonicalCurriculum(t *testing.T) {
	catalog := Bundled()
	before, err := json.Marshal(catalog)
	if err != nil {
		t.Fatal(err)
	}
	for _, level := range catalog.Levels {
		for _, skill := range level.Skills {
			_ = CompactTitle(skill)
		}
	}
	after, err := json.Marshal(catalog)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) || strings.Contains(string(after), "compactTitle") {
		t.Fatal("presentation metadata must not change the serialized curriculum")
	}
	for _, tc := range []struct{ id, want string }{
		{"Splash5:8-front-crawl", "Front crawl 50m"},
		{"Splash5:8-back-crawl", "Back crawl 50m"},
		{"Splash9:9-breaststroke", "Breaststroke 100m/yd"},
	} {
		skill, err := catalog.Skill(tc.id)
		if err != nil {
			t.Fatal(err)
		}
		if CompactTitle(skill) != tc.want {
			t.Fatalf("compact title lost component identity or distance: %s", tc.id)
		}
	}
}
