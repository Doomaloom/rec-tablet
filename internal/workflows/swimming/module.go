// Package swimming assembles swimming features. Other workflows mount independently.
package swimming

import (
	"rectablet/internal/shared/ui"
	"rectablet/internal/workflows/swimming/assessment"
	"rectablet/internal/workflows/swimming/attendance"
	"rectablet/internal/workflows/swimming/classes"
	"rectablet/internal/workflows/swimming/feedback"
	"rectablet/internal/workflows/swimming/navigation"
	"rectablet/internal/workflows/swimming/repository"
	"rectablet/internal/workflows/swimming/setup"
)

type Controller struct {
	*navigation.Controller
	setup *setup.Controller
}

// ConfigureDisplay applies a device preference independently of the active dataset.
func (c *Controller) ConfigureDisplay(mode ui.DisplayMode, save func(ui.DisplayMode)) {
	c.DisplayMode = mode.Normalize()
	c.setup.SaveDisplayMode = save
	_ = c.Refresh()
}

func New(s *repository.Store) *Controller {
	nav := navigation.New()
	classesFeature := classes.New(nav, s)
	nav.Features[navigation.Classes] = classesFeature
	nav.Features[navigation.Dates] = classesFeature
	nav.Features[navigation.Class] = classesFeature
	attendanceFeature := attendance.New(nav, s)
	nav.Features[navigation.Attendance] = attendanceFeature
	assessmentFeature := assessment.New(nav, s)
	nav.Features[navigation.Skills] = assessmentFeature
	nav.Features[navigation.SkillSwimmers] = assessmentFeature
	nav.Features[navigation.MultiGroupCount] = assessmentFeature
	nav.Features[navigation.MultiStudent] = assessmentFeature
	nav.Features[navigation.MultiSelection] = assessmentFeature
	nav.Features[navigation.MultiAssessment] = assessmentFeature
	nav.Features[navigation.ClassAssessment] = assessmentFeature
	nav.Features[navigation.TimedLaps] = assessmentFeature
	nav.Features[navigation.TimedHeat] = assessmentFeature
	nav.Features[navigation.Attempts] = assessmentFeature
	nav.Features[navigation.Assessment] = assessmentFeature
	feedbackFeature := feedback.New(nav, s)
	nav.Features[navigation.FeedbackSwimmers] = feedbackFeature
	nav.Features[navigation.Feedback] = feedbackFeature
	nav.Features[navigation.FeedbackDetail] = feedbackFeature
	setupFeature := setup.New(nav, s)
	nav.Features[navigation.Maintenance] = setupFeature
	c := &Controller{nav, setupFeature}
	_ = c.Refresh()
	return c
}

func (c *Controller) ConfigureDeviceSessions(join func(string, func(error)), profiles []setup.ProfileOption, switchProfile func(string) error, server string, saveServer func(string) error, cancel func()) {
	c.setup.JoinShare = join
	c.setup.Profiles = profiles
	c.setup.SwitchProfile = switchProfile
	c.setup.ShareServer = server
	c.setup.SaveShareServer = saveServer
	c.setup.CancelShare = cancel
	_ = c.Refresh()
}
func (c *Controller) Import(path string) error { return c.setup.Import(path) }
func (c *Controller) Export(path string) error { return c.setup.Export(path) }
