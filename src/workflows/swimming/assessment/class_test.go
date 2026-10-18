package assessment

import (
	"testing"

	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/navigation"
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
			c := New(navigation.New(), nil)
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
