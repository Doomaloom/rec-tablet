package timedheat

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"

	"rectablet/next/contracts"
	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
)

// Component owns timed assessment state and presentation through explicit data and action callbacks.
type Component struct {
	updateTitle func(string)
	Layout      Layout
	progress    common.ProgressActions
	navigation  common.Navigation
	State       State
	storage     Storage
	paginate    func(*navigation.State, []navigation.Row)
	small       *CompactTemplate
	large       *LargeTemplate
}

// State holds mode session and paging state.
type State struct {
	Selection            []string
	Heat                 *Heat
	LapsInput            string
	RosterPage, HeatPage int
}

// Storage contains only the data operations consumed by this mode.
type Storage struct {
	DiscardTimedHeat    func([]string) error
	GetEnrollment       func(string) (domain.Enrollment, error)
	NewTimedHeat        func([]string, string, string, int) ([]domain.TimedParticipant, error)
	Progress            func(string, string, string) (domain.Progress, error)
	RecordTimedLap      func(string) (int, error)
	RemoveLatestSuccess func(string, string, string) error
	ResetTimedHeat      func([]string) error
	SaveTimedHeat       func([]domain.TimedResult) error
	Skill               func(string, string) (curriculum.Skill, error)
}

// Options provides storage and navigation operations without a parent-component dependency.
type Options struct {
	Navigation  common.Navigation
	Storage     Storage
	ViewActions common.Actions
	Paginate    func(*navigation.State, []navigation.Row)
}

func New(options Options) *Component {
	c := &Component{navigation: options.Navigation, storage: options.Storage, paginate: options.Paginate}
	c.progress = common.ProgressActions{Navigation: options.Navigation, Remove: options.Storage.RemoveLatestSuccess}
	c.small = &CompactTemplate{Actions: options.ViewActions}
	c.large = &LargeTemplate{Actions: options.ViewActions,
		Pages:          Pages{Roster: &c.State.RosterPage, Heat: &c.State.HeatPage},
		LapsInput:      func() string { return c.State.LapsInput },
		SelectionCount: func() int { return len(c.State.Selection) },
		AddDigit: func(digit string) {
			_ = options.ViewActions.Run(func() error {
				if len(c.State.LapsInput) < 3 {
					c.State.LapsInput += digit
				}
				return nil
			})
		},
		ClearLaps:        func() { _ = options.ViewActions.Run(func() error { c.State.LapsInput = ""; return nil }) },
		StartHeat:        c.StartHeat,
		SelectedSwimmers: c.SelectedSwimmers,
	}
	return c
}

func (c *Component) Build(input common.Input) (navigation.State, error) {
	state := input.State
	f := input.Frame
	roster := input.Roster
	version := input.Course.CurriculumVersion
	rows := []navigation.Row{}
	s := c.storage
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
		benchmark, _ := common.BenchmarkFor(skill.ID)
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
			selected := common.Contains(c.State.Selection, p.ID)
			row := navigation.Row{ID: p.ID, Label: p.Name, Status: "unselected", ProgressSegments: min(3, max(0, progress.Successes)), Data: navigation.Parameters{Skill: skill.ID, Level: p.Level, Version: version}}
			if selected {
				row.Status = "selected"
				row.Selected = true
				row.Fill = ui.Green
			}
			if progress.Successes > 0 {
				person, skillID, curriculumVersion := p.ID, skill.ID, version
				row.OnHold = func() { _ = c.progress.RemoveLatestSuccess(person, skillID, curriculumVersion) }
				row.HoldDuration = 2 * time.Second
			}
			rows = append(rows, row)
		}
		if len(c.State.Selection) > 0 {
			state.HomeRailAction = common.SaveRailAction(c.ConfirmTimedSelection)
			state.BackRailAction = common.DiscardRailAction(c.DiscardTimedSelection)
		}
		state.Empty = "No enrolled swimmers are assigned to this skill's level."
		if c.Layout.Panels {
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
		benchmark, ok := common.BenchmarkFor(skill.ID)
		if !ok {
			return state, fmt.Errorf("skill is not a timed benchmark")
		}
		state.Title = "Laps Needed"
		navigation.AppendPrimaryInfo(&state, skill.Name+" · "+benchmark.Description)
		state.Progress = c.Input()
		state.ProgressInBody = true
		state.HomeRailAction = common.SaveRailAction(c.ConfirmTimedLaps)
		state.BackRailAction = common.DiscardRailAction(c.DiscardTimedSelection)
	case navigation.TimedHeat:
		if c.State.Heat == nil {
			return state, fmt.Errorf("timed heat is no longer available")
		}
		state.OnLeave = c.StopTicker
		benchmark, ok := common.BenchmarkFor(c.State.Heat.SkillID)
		if !ok {
			return state, fmt.Errorf("skill is not a timed benchmark")
		}
		state.Title = formatElapsed(c.Elapsed())
		state.BindTitle = func(update func(string)) { c.updateTitle = update }
		navigation.AppendPrimaryInfo(&state, benchmark.Description)
		state.HeaderTitleFontSize = 34
		startLabel := "Start"
		if c.State.Heat.Running {
			startLabel = "Pause"
		}
		state.HeaderActions = []contracts.HeaderAction{
			{Label: startLabel, Run: c.ToggleTimedTimer, Fill: ui.Green, Ink: ui.White, FontSize: 18},
			{Label: "Reset", Run: c.ResetTimedHeat, Fill: ui.Rail, Ink: ui.Blue, FontSize: 18},
		}
		state.HomeRailAction = common.SaveRailAction(c.SaveTimedHeat)
		state.BackRailAction = common.DiscardRailAction(c.DiscardTimedHeat)
		for i := range c.State.Heat.Participants {
			participant := &c.State.Heat.Participants[i]
			lapSummary := fmt.Sprintf("%d/%d laps", participant.Participant.LapsCompleted, participant.Participant.LapsRequired)
			status := "in_progress"
			if participant.Completed {
				lapSummary += " · " + formatElapsed(participant.CompletedAt)
				status = "successful"
			}
			rows = append(rows, navigation.Row{ID: "timed-lap:" + participant.Participant.PersonID, Label: participant.Participant.Name + " · " + lapSummary, Status: status, ProgressValue: participant.Participant.LapsCompleted, ProgressTotal: participant.Participant.LapsRequired})
		}
		state.Empty = "No swimmers in this heat."
		if c.Layout.Panels {
			state.PageSize = max(1, len(rows))
		}

	}
	c.paginate(&state, rows)
	if state.ComposeBody == nil {
		c.compose(&state, f)
	}
	return state, nil
}

func (c *Component) Input() string {
	if c.State.LapsInput == "" {
		return "Enter laps needed, then press ✓"
	}
	return "Laps: " + c.State.LapsInput
}

func (c *Component) Elapsed() time.Duration {
	if c.State.Heat == nil {
		return 0
	}
	if c.State.Heat.Running {
		return time.Since(c.State.Heat.StartedAt)
	}
	return c.State.Heat.Elapsed
}

func (c *Component) StopTicker() {
	c.updateTitle = nil
	if c.State.Heat == nil || c.State.Heat.Stop == nil {
		return
	}
	select {
	case <-c.State.Heat.Stop:
	default:
		close(c.State.Heat.Stop)
	}
	c.State.Heat.Stop = nil
}

func (c *Component) StartTicker(heat *Heat) {
	if heat.Stop != nil {
		return
	}
	heat.Stop = make(chan struct{})
	stop := heat.Stop
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				fyne.Do(func() {
					c.tick(heat, stop)
				})
			case <-stop:
				return
			}
		}
	}()
}

// Layout is chosen by the assessment root.
type Layout struct{ Panels bool }

// tick updates only the clock so an in-flight press keeps its original button.
func (c *Component) tick(heat *Heat, stop chan struct{}) {
	if c.State.Heat != heat || heat.Stop != stop || !heat.Running || c.updateTitle == nil {
		return
	}
	c.updateTitle(formatElapsed(c.Elapsed()))
}
