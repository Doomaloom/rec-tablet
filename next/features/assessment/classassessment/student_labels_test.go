package classassessment

import (
	"reflect"
	"testing"

	"rectablet/src/workflows/swimming/domain"
)

func TestStudentLabels(t *testing.T) {
	for _, tc := range []struct {
		name        string
		names, want []string
	}{
		{"initials", []string{"Alex Smith", "Blair Smith", "Alex Jones"}, []string{"Alex S", "Blair S", "Alex J"}},
		{"matching initials", []string{"Alex Smith", "Alex Smyth", "Alex Stone", "Blair Smith", "Alex Jones"}, []string{"Alex Smi", "Alex Smy", "Alex St", "Blair S", "Alex J"}},
		{"case insensitive", []string{"Alex Smith", "alex SMYTH"}, []string{"Alex Smi", "alex SMY"}},
		{"unicode", []string{"Zoë Émile", "Zoë Étienne"}, []string{"Zoë Ém", "Zoë Ét"}},
		{"prefix and duplicate", []string{"Alex Li", "Alex Lin", "Sam Smith", "Sam Smith"}, []string{"Alex Li", "Alex Lin", "Sam Smith", "Sam Smith"}},
		{"single and compound names", []string{" Alex ", "Blair van Dyke", ""}, []string{"Alex", "Blair v", ""}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var students []domain.MultiAttempt
			for _, name := range tc.names {
				students = append(students, domain.MultiAttempt{Name: name})
			}
			if got := studentLabels(students); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("labels = %q, want %q", got, tc.want)
			}
		})
	}
}
