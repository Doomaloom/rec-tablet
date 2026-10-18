package assessment

import (
	"rectablet/src/workflows/swimming/assessment/classassessment"
	"rectablet/src/workflows/swimming/assessment/internal/common"
	"rectablet/src/workflows/swimming/assessment/multiassessment"
	"rectablet/src/workflows/swimming/assessment/singleassessment"
	"rectablet/src/workflows/swimming/assessment/skillpicker"
	"rectablet/src/workflows/swimming/assessment/timedheat"
	"rectablet/src/workflows/swimming/assessment/view/shared"
	"rectablet/src/workflows/swimming/navigation"
)

func (c *Component) bindComponents() {
	nav := common.Navigation{
		Frame: c.navigation.Frame,
		Rows:  func() []navigation.Row { return c.navigation.State.Rows },
		Run:   c.navigation.Run, Push: c.navigation.Push, Pop: c.navigation.Pop,
		SetPage: func(page int) { c.navigation.History.Frames[len(c.navigation.History.Frames)-1].Page = page },
		UpdateFrame: func(change func(*navigation.Frame)) {
			change(&c.navigation.History.Frames[len(c.navigation.History.Frames)-1])
		},
		Saved: c.navigation.Saved, Refresh: c.navigation.Refresh,
	}
	viewActions := shared.Actions{Select: c.navigation.Select, Mark: c.navigation.Mark, Submit: c.navigation.Submit, Run: c.navigation.Run}
	var skillPickerStorage skillpicker.Storage
	if c.store != nil {
		skillPickerStorage = skillpicker.Storage{
			CatalogFor:          c.store.CatalogFor,
			Progress:            c.store.Progress,
			RemoveLatestSuccess: c.store.RemoveLatestSuccess,
			Skill:               c.store.Skill,
		}
	}
	c.skillPicker = skillpicker.New(skillpicker.Options{Storage: skillPickerStorage, Navigation: nav, ViewActions: viewActions, Paginate: c.navigation.Paginate,
		Actions: skillpicker.Actions{OpenSelectedMode: c.openLargeSelectedMode, ModeSwitch: c.skillModeSwitch, OpenTimedSkill: c.openLargeTimedSkill, OpenSkillSwimmer: c.openLargeSkillSwimmer},
	})
	var singleStorage singleassessment.Storage
	if c.store != nil {
		singleStorage = singleassessment.Storage{
			DiscardAttempt:      c.store.DiscardAttempt,
			Draft:               c.store.Draft,
			EditAttempt:         c.store.EditAttempt,
			GetAttempt:          c.store.GetAttempt,
			GetEnrollment:       c.store.GetEnrollment,
			ListAttempts:        c.store.ListAttempts,
			Mark:                c.store.Mark,
			NewAttempt:          c.store.NewAttempt,
			Progress:            c.store.Progress,
			RemoveLatestSuccess: c.store.RemoveLatestSuccess,
			Skill:               c.store.Skill,
			Submit:              c.store.Submit,
		}
	}
	c.single = singleassessment.New(singleassessment.Options{Storage: singleStorage, Navigation: nav, ViewActions: viewActions, Paginate: c.navigation.Paginate})
	var classStorage classassessment.Storage
	if c.store != nil {
		classStorage = classassessment.Storage{
			DiscardAttempts:   c.store.DiscardAttempts,
			Draft:             c.store.Draft,
			ListRoster:        c.store.ListRoster,
			Mark:              c.store.Mark,
			OpenMultiAttempts: c.store.OpenMultiAttempts,
			Skill:             c.store.Skill,
			SubmitAttempts:    c.store.SubmitAttempts,
		}
	}
	c.class = classassessment.New(classassessment.Options{Storage: classStorage, Navigation: nav, ViewActions: viewActions, Paginate: c.navigation.Paginate,
		Actions: classassessment.Actions{BeforeEnter: func() { c.multi.LeaveSetup(); c.multi.State.Session, c.multi.State.Edit = nil, nil }},
	})
	var multiStorage multiassessment.Storage
	if c.store != nil {
		multiStorage = multiassessment.Storage{
			Attendance:           c.store.Attendance,
			DiscardMultiAttempts: c.store.DiscardMultiAttempts,
			Draft:                c.store.Draft,
			GetCourse:            c.store.GetCourse,
			ListRoster:           c.store.ListRoster,
			OpenMultiAttempts:    c.store.OpenMultiAttempts,
			ReplaceDraftAnswers:  c.store.ReplaceDraftAnswers,
			Skill:                c.store.Skill,
			SubmitAttempts:       c.store.SubmitAttempts,
		}
	}
	c.multi = multiassessment.New(multiassessment.Options{Storage: multiStorage, Navigation: nav, ViewActions: viewActions, Paginate: c.navigation.Paginate,
		Actions: multiassessment.Actions{EnterSingleMode: c.EnterSingleMode, ModeSwitch: c.skillModeSwitch},
	})
	var timedStorage timedheat.Storage
	if c.store != nil {
		timedStorage = timedheat.Storage{
			DiscardTimedHeat:    c.store.DiscardTimedHeat,
			GetEnrollment:       c.store.GetEnrollment,
			NewTimedHeat:        c.store.NewTimedHeat,
			Progress:            c.store.Progress,
			RecordTimedLap:      c.store.RecordTimedLap,
			RemoveLatestSuccess: c.store.RemoveLatestSuccess,
			ResetTimedHeat:      c.store.ResetTimedHeat,
			SaveTimedHeat:       c.store.SaveTimedHeat,
			Skill:               c.store.Skill,
		}
	}
	c.timed = timedheat.New(timedheat.Options{Storage: timedStorage, Navigation: nav, ViewActions: viewActions, Paginate: c.navigation.Paginate})
}
