package assessment

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2/test"
	"rectablet/internal/shared/ui"
	"rectablet/internal/shell"
	"rectablet/internal/workflows/swimming/curriculum"
	"rectablet/internal/workflows/swimming/navigation"
)

func TestModeAvailability(t *testing.T) {
	for _, tc := range []struct {
		name  string
		count int
		multi bool
	}{
		{"Water Smart Messages", 6, true},
		{"Rescue", 7, false},
		{"Front crawl 5 m", 4, false},
		{"Back Crawl", 3, false},
		{"Breaststroke", 4, false},
		{"Stroke evaluation", 2, false},
		{"Butterfly", 5, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			skill := curriculum.Skill{Name: tc.name, Subskills: make([]curriculum.Subskill, tc.count)}
			if got := allowsMultiMode(skill); got != tc.multi {
				t.Fatalf("Multi available = %v, want %v", got, tc.multi)
			}
			c := &Controller{}
			options := c.skillModeSwitch(skill, false)
			if options[0].Label != "Single" || !options[0].Selected || options[len(options)-1].Label != "Class" {
				t.Fatalf("wrong mode options: %+v", options)
			}
			if (len(options) == 3) != tc.multi {
				t.Fatalf("wrong option count: %d", len(options))
			}
		})
	}
}

func TestClassChecklistFitsEntireCatalog(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	app.Settings().SetTheme(ui.AppTheme{})
	for _, level := range curriculum.Bundled().Levels {
		for _, skill := range level.Skills {
			var rows []navigation.Row
			for _, subskill := range skill.Subskills {
				rows = append(rows, navigation.Row{Label: subskill.Text})
			}
			info := "Saturday morning swimming group · 9:00 AM · " + skill.Number + ". " + skill.Name
			infoFont := float32(20)
			if ui.WrappedTextHeight(info, 680, 20, true) > 2*ui.WrappedTextHeight("Ag", 680, 20, true)+0.01 {
				infoFont = 14
			}
			infoHeight := max(float32(30), ui.WrappedTextHeight(info, 680, infoFont, true)+4)
			checklistTop := ui.OuterMargin + shell.SplitHeaderHeight + ui.Gap + infoHeight + ui.Gap
			availableHeight := 480 - checklistTop - classChecklistBottom
			layout := fitClassChecklist(rows, 680, availableHeight)
			total := float32(len(rows))*layout.height + float32(max(0, len(rows)-1))*classChecklistGap
			for i, row := range rows {
				needed := ui.WrappedTextHeight(row.Label, 680-2*ui.ButtonPadding, layout.font, true)
				if needed > layout.height-4 {
					t.Fatalf("%s criterion %d clips: needs %v, has %v", skill.ID, i, needed, layout.height-4)
				}
				if layout.font < 12 {
					t.Fatalf("%s criterion %d is too small: font %v", skill.ID, i, layout.font)
				}
			}
			if total > availableHeight+0.01 {
				t.Fatalf("%s does not fit on the page: height %v", skill.ID, total)
			}
		}
	}
}

func TestClassChecklistUsesSharedFontToFitLongestLabel(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	app.Settings().SetTheme(ui.AppTheme{})
	rows := make([]navigation.Row, 7)
	for i := range rows {
		rows[i].Label = "Short criterion"
	}
	if layout := fitClassChecklist(rows, 600, 384); layout.font != 22 {
		t.Fatal("short labels should use the full font size when they all fit")
	}
	rows[3].Label = strings.Repeat("A longer criterion that needs to wrap. ", 4)
	layout := fitClassChecklist(rows, 600, 384)
	if layout.height != (384-6*classChecklistGap)/7 {
		t.Fatal("row height should depend only on the available page space and row count")
	}
	if layout.font >= 22 {
		t.Fatal("shared font did not shrink to accommodate the longest label")
	}
	for _, row := range rows {
		if ui.WrappedTextHeight(row.Label, 600-2*ui.ButtonPadding, layout.font, true) > layout.height-4 {
			t.Fatal("label does not fit using the shared font size")
		}
	}
}
