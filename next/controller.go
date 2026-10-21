// Package next assembles the independent swimming workflow and its features.
package next

import (
	"rectablet/next/contracts"
	"rectablet/next/features/assessment"
	"rectablet/next/features/attendance"
	"rectablet/next/features/classes"
	"rectablet/next/features/feedback"
	"rectablet/next/features/setup"
	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/repository"
)

// Controller assembles the workflow and exposes the actions used by the contracts.
// Navigation and every feature are named dependencies, rather than embedded APIs.
type Controller struct {
	navigation *navigation.Controller
	classes    *classes.Component
	attendance *attendance.Component
	assessment *assessment.Component
	feedback   *feedback.Component
	setup      *setup.Component
}

var _ contracts.Presenter = (*Controller)(nil)

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
