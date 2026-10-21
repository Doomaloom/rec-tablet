package classassessment

import (
	"strings"
	"unicode"

	"rectablet/src/workflows/swimming/domain"
)

// studentLabels uses the shortest surname prefix that distinguishes matching given names.
// Use the whole session roster so labels stay stable when students are saved or paged.
func studentLabels(students []domain.MultiAttempt) []string {
	type name struct {
		given   string
		surname []rune
	}
	names := make([]name, len(students))
	for i, student := range students {
		parts := strings.Fields(student.Name)
		if len(parts) > 0 {
			names[i].given = parts[0]
			names[i].surname = []rune(strings.Join(parts[1:], " "))
		}
	}
	labels := make([]string, len(names))
	for i, student := range names {
		labels[i] = student.given
		if len(student.surname) == 0 {
			continue
		}
		length := 1
		for j, other := range names {
			if i == j || !strings.EqualFold(student.given, other.given) {
				continue
			}
			common := 0
			for common < len(student.surname) && common < len(other.surname) && unicode.ToLower(student.surname[common]) == unicode.ToLower(other.surname[common]) {
				common++
			}
			length = max(length, min(len(student.surname), common+1))
		}
		labels[i] += " " + string(student.surname[:length])
	}
	return labels
}
