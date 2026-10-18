// Package swimming assembles swimming features. Other workflows mount independently.
package swimming

import (
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/assessment"
	"rectablet/src/workflows/swimming/attendance"
	"rectablet/src/workflows/swimming/classes"
	"rectablet/src/workflows/swimming/feedback"
	"rectablet/src/workflows/swimming/navigation"
	"rectablet/src/workflows/swimming/repository"
	"rectablet/src/workflows/swimming/setup"
)

// Controller assembles the workflow and exposes the actions used by the shell.
// Navigation and every feature are named dependencies, rather than embedded APIs.
type Controller struct {
	navigation *navigation.Controller
	classes    *classes.Component
	attendance *attendance.Component
	assessment *assessment.Component
	feedback   *feedback.Component
	setup      *setup.Component
}

var _ shell.Presenter = (*Controller)(nil)

// ConfigureDisplay applies a device preference independently of the active dataset.
func (c *Controller) ConfigureDisplay(mode ui.DisplayMode, save func(ui.DisplayMode)) {
	c.navigation.DisplayMode = mode.Normalize()
	c.setup.SaveDisplayMode = save
	_ = c.Refresh()
}

func New(store *repository.Store) *Controller {
	nav := navigation.New()
	c := &Controller{
		navigation: nav,
		classes:    classes.New(nav, store),
		attendance: attendance.New(nav, store),
		assessment: assessment.New(nav, store),
		feedback:   feedback.New(nav, store),
		setup:      setup.New(nav, store),
	}
	register := func(feature navigation.Feature, routes ...navigation.Route) {
		for _, route := range routes {
			nav.Features[route] = feature
		}
	}
	register(c.classes, navigation.Classes, navigation.Dates, navigation.Class)
	register(c.attendance, navigation.Attendance)
	register(c.assessment,
		navigation.Skills, navigation.SkillSwimmers, navigation.MultiGroupCount,
		navigation.MultiStudent, navigation.MultiSelection, navigation.MultiAssessment,
		navigation.ClassAssessment, navigation.TimedLaps, navigation.TimedHeat,
		navigation.Attempts, navigation.Assessment,
	)
	register(c.feedback, navigation.FeedbackSwimmers, navigation.Feedback, navigation.FeedbackDetail)
	register(c.setup, navigation.Maintenance)
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
