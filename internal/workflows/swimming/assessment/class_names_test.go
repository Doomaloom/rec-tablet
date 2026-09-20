package assessment

import (
	"fmt"
	"testing"

	"rectablet/internal/workflows/swimming/domain"
)

func TestAbbreviateClassNames(t *testing.T) {
	for _, tc := range []struct {
		name        string
		names, want []string
	}{
		{"basic", []string{"Alexandria Montgomery-Sutherland", "Harrold", "  Jo   van  Smith  "}, []string{"Alexandria M.", "Harrold", "Jo S."}},
		{"second letter", []string{"Sam Taylor", "Sam Turner", "Alex Taylor"}, []string{"Sam Ta.", "Sam Tu.", "Alex T."}},
		{"deep prefix", []string{"Sam Taylor", "Sam Tayler", "Sam Turner"}, []string{"Sam Taylo.", "Sam Tayle.", "Sam Tu."}},
		{"surname prefix", []string{"Sam Lee", "Sam Leeds"}, []string{"Sam Lee", "Sam Leed."}},
		{"identical surnames", []string{"Sam John Smith", "Sam Jane Smith"}, []string{"Sam John Smith", "Sam Jane Smith"}},
		{"case insensitive", []string{"Sam Taylor", "sam turner"}, []string{"Sam Ta.", "sam tu."}},
		{"unicode", []string{"Éva Åberg", "Éva Åström", "李 雷"}, []string{"Éva Åb.", "Éva Ås.", "李 雷"}},
		{"identical names", []string{"Sam Smith", "Sam Smith"}, []string{"Sam Smith", "Sam Smith"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			students := make([]domain.MultiAttempt, len(tc.names))
			for i, name := range tc.names {
				students[i] = domain.MultiAttempt{PersonID: fmt.Sprint(i), Name: name}
			}
			labels := abbreviateClassNames(students)
			for i, want := range tc.want {
				if got := labels[students[i].PersonID]; got != want {
					t.Fatalf("%q abbreviated to %q, want %q", tc.names[i], got, want)
				}
			}
		})
	}
}
