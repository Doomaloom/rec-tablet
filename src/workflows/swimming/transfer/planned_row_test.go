package transfer

import (
	"encoding/json"
	"strings"
	"testing"
)

// Rows as the website's device export serialises them (triage D5).
func TestPlannedRowTitleNamesEveryRowKind(t *testing.T) {
	var plan LessonPlan
	body := `{"date":"2026-11-19","curriculumLevel":"","rows":[
		{"skill":"","activity":"What Time is it Mr. Shark?\n\nChoose one participant as the shark.","activities":[{"kind":"library","text":"What Time is it Mr. Shark?\n\nChoose one participant as the shark."}],"location":"Lane","duration":5},
		{"skill":"Front crawl 50m/yd","activity":"","location":"Lane","duration":10},
		{"skill":"","activity":"  Partner relay   with noodles \nThen swap.","activities":[{"kind":"custom","text":"  Partner relay   with noodles \nThen swap."}],"location":"Lane","duration":5},
		{"skill":"","activity":"Bobbing race","location":"Lane","duration":5},
		{"skill":"","activity":"","location":"Lane","duration":5,"equipment":["Noodles"]},
		{"skill":"","activity":"` + strings.Repeat("x", 80) + `","location":"Lane","duration":5}
	]}`
	if err := json.Unmarshal([]byte(body), &plan); err != nil {
		t.Fatal(err)
	}
	want := []string{"What Time is it Mr. Shark?", "Front crawl 50m/yd", "Partner relay with noodles", "Bobbing race", "", strings.Repeat("x", 59) + "…"}
	for i, row := range plan.Rows {
		if got := row.Title(); got != want[i] {
			t.Errorf("row %d title = %q, want %q", i, got, want[i])
		}
	}
}
