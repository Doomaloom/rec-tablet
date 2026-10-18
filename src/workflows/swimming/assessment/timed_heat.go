package assessment

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"

	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/assessment/view/large"
	"rectablet/src/workflows/swimming/assessment/view/small"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
	"rectablet/src/workflows/swimming/navigation"
)

// TimedHeat owns its mode state and presentation. Data and actions arrive as callbacks.
type TimedHeat struct {
	selection            []string
	heat                 *timedHeatState
	lapsInput            string
	rosterPage, heatPage int
	data                 timedHeatData
	actions              timedHeatActions
	paginate             func(*navigation.State, []navigation.Row)
	small                *small.View
	large                *large.View
}

type timedHeatData struct {
	Progress func(string, string, string) (domain.Progress, error)
	Skill    func(string, string) (curriculum.Skill, error)
}

type timedHeatActions struct {
	ConfirmTimedLaps      func() error
	ConfirmTimedSelection func() error
	DiscardTimedHeat      func() error
	DiscardTimedSelection func() error
	Refresh               func() error
	RemoveLatestSuccess   func(string, string, string) error
	ResetTimedHeat        func() error
	SaveTimedHeat         func() error
	ToggleTimedTimer      func() error
}

func (c *TimedHeat) build(input modeInput) (navigation.State, error) {
	state := input.state
	f := input.frame
	roster := input.roster
	version := input.course.CurriculumVersion
	rows := []navigation.Row{}
	s := c.data
	switch f.Screen {
	case navigation.SkillSwimmers:
		state.PrimaryInfo = ""
		if f.Version != "" {
			version = f.Version
		}
		skill, e := s.Skill(version, f.Skill)
		if e != nil {
			return state, e
		}
		benchmark, _ := benchmarkFor(skill.ID)
		state.Title = "Select Swimmers"
		navigation.AppendPrimaryInfo(&state, benchmark.Description)
		for _, p := range roster {
			if !p.Active || p.Level != f.Level {
				continue
			}
			progress, e := s.Progress(p.ID, skill.ID, version)
			if e != nil {
				return state, e
			}
			selected := containsString(c.selection, p.ID)
			row := navigation.Row{ID: p.ID, Label: p.Name, Status: "unselected", ProgressSegments: min(3, max(0, progress.Successes)), Data: navigation.Parameters{Skill: skill.ID, Level: p.Level, Version: version}}
			if selected {
				row.Status = "selected"
				row.Selected = true
				row.Fill = ui.Green
			}
			if progress.Successes > 0 {
				person, skillID, curriculumVersion := p.ID, skill.ID, version
				row.OnHold = func() { _ = c.actions.RemoveLatestSuccess(person, skillID, curriculumVersion) }
				row.HoldDuration = 2 * time.Second
			}
			rows = append(rows, row)
		}
		if len(c.selection) > 0 {
			state.HomeRailAction = saveRailAction(c.actions.ConfirmTimedSelection)
			state.BackRailAction = discardRailAction(c.actions.DiscardTimedSelection)
		}
		state.Empty = "No enrolled swimmers are assigned to this skill's level."
		if state.DisplayMode.IsLarge() {
			c.configureLargeTimedSetup(&state, rows)
		}
	case navigation.TimedLaps:
		if f.Version != "" {
			version = f.Version
		}
		skill, e := s.Skill(version, f.Skill)
		if e != nil {
			return state, e
		}
		benchmark, ok := benchmarkFor(skill.ID)
		if !ok {
			return state, fmt.Errorf("skill is not a timed benchmark")
		}
		state.Title = "Laps Needed"
		navigation.AppendPrimaryInfo(&state, skill.Name+" · "+benchmark.Description)
		state.Progress = c.timedHeatInput()
		state.ProgressInBody = true
		state.HomeRailAction = saveRailAction(c.actions.ConfirmTimedLaps)
		state.BackRailAction = discardRailAction(c.actions.DiscardTimedSelection)
	case navigation.TimedHeat:
		if c.heat == nil {
			return state, fmt.Errorf("timed heat is no longer available")
		}
		state.OnLeave = c.stopTimedTicker
		benchmark, ok := benchmarkFor(c.heat.SkillID)
		if !ok {
			return state, fmt.Errorf("skill is not a timed benchmark")
		}
		state.Title = formatHeatElapsed(c.timedHeatElapsed())
		navigation.AppendPrimaryInfo(&state, benchmark.Description)
		state.HeaderTitleFontSize = 34
		startLabel := "Start"
		if c.heat.Running {
			startLabel = "Pause"
		}
		state.HeaderActions = []shell.HeaderAction{
			{Label: startLabel, Run: c.actions.ToggleTimedTimer, Fill: ui.Green, Ink: ui.White, FontSize: 18},
			{Label: "Reset", Run: c.actions.ResetTimedHeat, Fill: ui.Rail, Ink: ui.Blue, FontSize: 18},
		}
		state.HomeRailAction = saveRailAction(c.actions.SaveTimedHeat)
		state.BackRailAction = discardRailAction(c.actions.DiscardTimedHeat)
		for i := range c.heat.Participants {
			participant := &c.heat.Participants[i]
			lapSummary := fmt.Sprintf("%d/%d laps", participant.participant.LapsCompleted, participant.participant.LapsRequired)
			status := "in_progress"
			if participant.completed {
				lapSummary += " · " + formatHeatElapsed(participant.completedAt)
				status = "successful"
			}
			rows = append(rows, navigation.Row{ID: "timed-lap:" + participant.participant.PersonID, Label: participant.participant.Name + " · " + lapSummary, Status: status, ProgressValue: participant.participant.LapsCompleted, ProgressTotal: participant.participant.LapsRequired})
		}
		state.Empty = "No swimmers in this heat."
		if state.DisplayMode.IsLarge() {
			state.PageSize = max(1, len(rows))
		}

	}
	c.paginate(&state, rows)
	if state.ComposeBody == nil {
		c.compose(&state, f)
	}
	return state, nil
}

func (c *TimedHeat) timedHeatInput() string {
	if c.lapsInput == "" {
		return "Enter laps needed, then press ✓"
	}
	return "Laps: " + c.lapsInput
}

func (c *TimedHeat) timedHeatElapsed() time.Duration {
	if c.heat == nil {
		return 0
	}
	if c.heat.Running {
		return time.Since(c.heat.StartedAt)
	}
	return c.heat.Elapsed
}

func (c *TimedHeat) stopTimedTicker() {
	if c.heat == nil || c.heat.stop == nil {
		return
	}
	select {
	case <-c.heat.stop:
	default:
		close(c.heat.stop)
	}
	c.heat.stop = nil
}

func (c *TimedHeat) startTimedTicker(heat *timedHeatState) {
	if heat.stop != nil {
		return
	}
	heat.stop = make(chan struct{})
	stop := heat.stop
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				fyne.Do(func() {
					if c.heat == heat && heat.Running {
						_ = c.actions.Refresh()
					}
				})
			case <-stop:
				return
			}
		}
	}()
}

func (c *TimedHeat) compose(state *shell.State, f navigation.Frame) {
	switch f.Screen {
	case navigation.SkillSwimmers:
		state.ComposeBody = c.composeTimedSelectionBody(*state)
	case navigation.TimedLaps:
		if state.DisplayMode.IsLarge() {
			state.ComposeBody = c.large.ComposeLargeTimedLapsBody(*state)
		} else {
			state.ComposeBody = c.small.ComposeTimedLapsBody(*state)
		}
	case navigation.TimedHeat:
		if state.DisplayMode.IsLarge() {
			state.ComposeBody = c.large.ComposeLargeTimedHeatBody(*state)
		} else {
			state.ComposeBody = c.small.ComposeTimedHeatBody(*state)
		}
	}
}

func (c *TimedHeat) composeTimedSelectionBody(state shell.State) ui.BodyComposer {
	snapshot := state
	return func(surface *ui.Surface) {
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}
		rows := append([]ui.Item(nil), snapshot.Rows...)
		for i := range rows {
			if rows[i].Status == "selected" {
				rows[i].Fill = ui.Green
				rows[i].Selected = true
			}
		}
		if snapshot.DisplayMode.IsLarge() {
			c.large.RenderLargeTimedSelection(surface, rows)
		} else {
			c.small.RenderSmallTimedSelection(surface, rows)
		}
	}
}

func (c *TimedHeat) configureLargeTimedSetup(state *navigation.State, rows []navigation.Row) {
	state.Title = "Set Up Timed Heat"
	state.PageSize = max(1, len(rows))
	state.HomeRailAction = nil
	state.BackRailAction = discardRailAction(c.actions.DiscardTimedSelection)
	state.Progress = c.timedHeatInput()
	state.ProgressInBody = true
	state.ComposeBody = c.large.ComposeLargeTimedSetup(*state, rows)
}
