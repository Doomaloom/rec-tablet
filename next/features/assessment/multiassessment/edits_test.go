package multiassessment

import (
	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/navigation"
	"rectablet/src/workflows/swimming/curriculum"
	"testing"
)

func TestQuickMultiChecklistKeepsEditsLocalUntilSave(t *testing.T) {
	rows := []navigation.Row{{ID: "criterion", Label: "Swim", Status: "unmarked"}}
	c := New(Options{Navigation: common.Navigation{
		Frame: func() navigation.Frame { return navigation.Frame{Screen: navigation.MultiStudent} },
		Rows:  func() []navigation.Row { return rows },
	}})
	c.State.Session = &Session{Skill: curriculum.Skill{Subskills: []curriculum.Subskill{{ID: "criterion", Text: "Swim"}}}}
	c.State.Edit = &StudentEdit{Answers: map[string]string{}}
	edit, ok := c.PrepareEdit("criterion")
	if !ok || edit.Save != nil || c.State.Edit.Answers["criterion"] != "successful" {
		t.Fatal("quick edit must preserve local-until-save behavior")
	}
	rows[0] = edit.Row
	edit, ok = c.PrepareEdit("criterion")
	if !ok || edit.Row.Status != "unmarked" || c.State.Edit.Answers["criterion"] != "" {
		t.Fatal("repeated multi tap lost local toggle state")
	}
}
