package assessment

import (
	"rectablet/next/features/assessment/classassessment"
	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/features/assessment/multiassessment"
	"rectablet/next/features/assessment/singleassessment"
	"rectablet/next/features/assessment/skillpicker"
	"rectablet/next/features/assessment/timedheat"
	"rectablet/next/navigation"
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
	c.configureLayout(state.DisplayMode)
	course, roster, err := navigation.LessonContext(c.store, frame, &state)
	if err != nil {
		return state, err
	}
	input := common.Input{State: state, Frame: frame, Course: course, Roster: roster}
	return c.composeFeature(input)
}
