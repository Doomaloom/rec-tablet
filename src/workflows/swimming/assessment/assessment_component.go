package assessment

import (
	"rectablet/src/workflows/swimming/assessment/classassessment"
	"rectablet/src/workflows/swimming/assessment/internal/common"
	"rectablet/src/workflows/swimming/assessment/multiassessment"
	"rectablet/src/workflows/swimming/assessment/singleassessment"
	"rectablet/src/workflows/swimming/assessment/skillpicker"
	"rectablet/src/workflows/swimming/assessment/timedheat"
	"rectablet/src/workflows/swimming/navigation"
)

// Component coordinates persistence and navigation and has five mode components.
type Component struct {
	navigation  *navigation.Controller
	store       Repository
	skillPicker *skillpicker.Component
	single      *singleassessment.Component
	class       *classassessment.Component
	multi       *multiassessment.Component
	timed       *timedheat.Component
}

var _ navigation.Feature = (*Component)(nil)

func New(nav *navigation.Controller, store Repository) *Component {
	c := &Component{navigation: nav, store: store}
	c.bindComponents()
	return c
}

func (c *Component) Build() (navigation.State, error) {
	state, frame := c.navigation.InitialState(), c.navigation.Frame()
	course, roster, err := navigation.LessonContext(c.store, frame, &state)
	if err != nil {
		return state, err
	}
	input := common.Input{State: state, Frame: frame, Course: course, Roster: roster}
	switch frame.Screen {
	case navigation.Skills:
		return c.skillPicker.Build(input)
	case navigation.SkillSwimmers:
		if _, timed := common.BenchmarkFor(frame.Skill); timed {
			return c.timed.Build(input)
		}
		return c.skillPicker.Build(input)
	case navigation.Assessment, navigation.Attempts:
		return c.single.Build(input)
	case navigation.TimedLaps, navigation.TimedHeat:
		return c.timed.Build(input)
	case navigation.ClassAssessment:
		state, err = c.class.Build(input)
	case navigation.MultiGroupCount:
		state, err = c.multi.BuildGroupCount(input)
	case navigation.MultiSelection:
		return c.multi.BuildSelection(input)
	case navigation.MultiAssessment:
		state, err = c.multi.BuildAssessment(input)
	case navigation.MultiStudent:
		state, err = c.multi.BuildStudent(input)
	}
	if err == nil {
		c.navigation.History.Frames[len(c.navigation.History.Frames)-1].Page = state.Page
	}
	return state, err
}
