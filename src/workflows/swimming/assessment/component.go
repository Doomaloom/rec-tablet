package assessment

import (
	"rectablet/src/workflows/swimming/domain"
	"rectablet/src/workflows/swimming/navigation"
)

// Component coordinates persistence and navigation and has five mode components.
type Component struct {
	navigation  *navigation.Controller
	store       Repository
	skillPicker SkillPicker
	single      SingleAssessment
	class       ClassAssessment
	multi       MultiAssessment
	timed       TimedHeat
}

var _ navigation.Feature = (*Component)(nil)

func New(nav *navigation.Controller, store Repository) *Component {
	c := &Component{navigation: nav, store: store}
	c.bindComponents()
	return c
}

// modeInput is a fresh lesson snapshot, prepared by the parent for a child.
type modeInput struct {
	state  navigation.State
	frame  navigation.Frame
	course domain.Course
	roster []domain.Enrollment
}

func (c *Component) Build() (navigation.State, error) {
	state, frame := c.navigation.InitialState(), c.navigation.Frame()
	course, roster, err := navigation.LessonContext(c.store, frame, &state)
	if err != nil {
		return state, err
	}
	input := modeInput{state: state, frame: frame, course: course, roster: roster}
	switch frame.Screen {
	case navigation.Skills:
		return c.skillPicker.build(input)
	case navigation.SkillSwimmers:
		if _, timed := benchmarkFor(frame.Skill); timed {
			return c.timed.build(input)
		}
		return c.skillPicker.build(input)
	case navigation.Assessment, navigation.Attempts:
		return c.single.build(input)
	case navigation.TimedLaps, navigation.TimedHeat:
		return c.timed.build(input)
	case navigation.ClassAssessment:
		state, err = c.class.build(input)
	case navigation.MultiGroupCount:
		state, err = c.multi.buildGroupCount(input)
	case navigation.MultiSelection:
		return c.multi.buildSelection(input)
	case navigation.MultiAssessment:
		state, err = c.multi.buildAssessment(input)
	case navigation.MultiStudent:
		state, err = c.multi.buildStudent(input)
	}
	if err == nil {
		c.navigation.History.Frames[len(c.navigation.History.Frames)-1].Page = state.Page
	}
	return state, err
}
