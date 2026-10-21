package curriculum

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestExtractMalformedStrokeBullet(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "TeenAdult2.html")
	source := `<p class="break-before-page"></p><p><b>14. Front Crawl and Back Crawl (25-50 m each)</b><br>
 • Front Crawl<br>- Body on front<br>• Back Crawl<br>- Body on back<br>
 • Continuous, rhythmic flutter kick initiated from hips; slight knee bend; feet relaxed<br>- Relaxed breathing</p>`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	level, err := extractBack(path)
	if err != nil {
		t.Fatal(err)
	}
	subs := level.Skills[0].Subskills
	if len(subs) != 4 {
		t.Fatalf("kick requirement lost: %+v", subs)
	}
	for _, sub := range subs[1:] {
		if !reflect.DeepEqual(sub.Group, []string{"Back Crawl"}) {
			t.Errorf("back crawl criterion has wrong context: %+v", sub)
		}
	}
}

func TestExtractRepeatedWaterSmartMessage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "TeenAdult3.html")
	source := `<p class="break-before-page"></p><p><b>14. Water Smart Messages</b><br>Check the ice<br>• Snowmobilers are at risk<br>•Check the ice<br>Swim to Survive<br>• Unexpected entry</p>`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	level, err := extractBack(path)
	if err != nil {
		t.Fatal(err)
	}
	subs := level.Skills[0].Subskills
	if len(subs) != 3 || subs[1].Text != "Check the ice" || !reflect.DeepEqual(subs[1].Group, []string{"Check the ice"}) {
		t.Fatalf("message lost: %+v", subs)
	}
}

func TestExtractSeparatesAirwayProcedureSelection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Splash8.html")
	source := `<p class="break-before-page"></p><p><b>12a. Obstructed Airway</b><br>• Assessment of degree of obstruction<br>- ask: "Are you choking?"<br>Selection of appropriate procedures:</p>`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	level, err := extractBack(path)
	if err != nil {
		t.Fatal(err)
	}
	subs := level.Skills[0].Subskills
	if len(subs) != 2 || subs[0].Text != `ask: "Are you choking?"` || subs[1].Text != "Selection of appropriate procedures:" {
		t.Fatalf("instructions merged into the question: %+v", subs)
	}
}

func TestExtractDirectoryIgnoresDynamicMarkingSheet(t *testing.T) {
	dir := t.TempDir()
	for name, source := range map[string]string{
		"Splash1.html":      `<table><tr><td class="rotate-cell">1. Front entry name</td></tr></table><p class="break-before-page"></p><p><b>1. Back entry name</b><br>• Controlled entry</p>`,
		"SkillMarking.html": `<html><body>Dynamic skill headers</body></html>`,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	catalog, err := ExtractDirectory(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Levels) != 1 || catalog.Levels[0].ID != "Splash1" {
		t.Fatalf("unexpected curriculum: %+v", catalog)
	}
}

func TestFrontColumnsSelectBackCriteria(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Lesson.html")
	source := `<table><tr><td class="rotate-cell">8a. Front Crawl 50 m</td><td class="rotate-cell">8b. Back Crawl 50 m</td></tr></table>
 <p class="break-before-page"></p><p><b>8. Swim each of the following strokes</b><br>• Minimum distance completed<br>Front Crawl<br>• Face in water<br>Back Crawl<br>• Ears in water</p>`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	level, err := Extract(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(level.Skills) != 2 {
		t.Fatal(level.Skills)
	}
	for i, want := range []struct{ number, name, criterion string }{{"8a", "Front Crawl 50 m", "Face in water"}, {"8b", "Back Crawl 50 m", "Ears in water"}} {
		skill := level.Skills[i]
		if skill.Number != want.number || skill.Name != want.name || skill.SourceName != "Swim each of the following strokes" {
			t.Errorf("front identity or back provenance lost: %+v", skill)
		}
		if len(skill.Subskills) != 2 || skill.Subskills[0].Text != "Minimum distance completed" || skill.Subskills[1].Text != want.criterion {
			t.Errorf("shared/individual criteria mixed: %+v", skill.Subskills)
		}
	}
}

func TestUnmatchedBackMessageIsPreservedSeparately(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Lesson.html")
	source := `<table><tr><td class="rotate-cell">10a. Water Smart Messages - Wear a Lifejacket</td><td class="rotate-cell">10b. Water Smart Messages - Swim to Survive</td></tr></table>
 <p class="break-before-page"></p><p><b>10. Water Smart Messages</b><br>• Participation required<br>Wear a Lifejacket<br>• Fasten it<br>Swim to Survive<br>• Swim to safety<br>Within Arms' Reach<br>• Stay close</p>`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	level, err := Extract(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(level.Skills) != 2 || len(level.UnmappedSubskills) != 1 || level.UnmappedSubskills[0].Text != "Stay close" {
		t.Fatalf("unmatched source requirement hidden: %+v", level)
	}
	for _, skill := range level.Skills {
		if len(skill.Subskills) != 2 {
			t.Fatalf("unrelated requirement assigned: %+v", skill)
		}
	}
}
