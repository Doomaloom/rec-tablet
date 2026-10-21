package assessment

import (
	"rectablet/next/features/assessment/classassessment"
	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/features/assessment/multiassessment"
	"rectablet/next/features/assessment/singleassessment"
	"rectablet/next/features/assessment/skillpicker"
	"rectablet/next/features/assessment/timedheat"
	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
)

// configureLayout selects the composition and associated paging policies once.
func (c *Component) configureLayout(mode ui.DisplayMode) {
	panels := mode.IsLarge()
	c.skillPicker.Layout = skillpicker.Layout{Panels: panels}
	c.single.Layout = singleassessment.Layout{Panels: panels}
	c.class.Layout = classassessment.Layout{Panels: panels}
	c.multi.Layout = multiassessment.Layout{Panels: panels}
	c.timed.Layout = timedheat.Layout{Panels: panels}
}

func (c *Component) composeFeature(input common.Input) (navigation.State, error) {
	state, frame := input.State, input.Frame
	var err error
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
