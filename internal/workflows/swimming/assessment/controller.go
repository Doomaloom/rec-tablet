package assessment

import (
	"fmt"
	"fyne.io/fyne/v2"
	"rectablet/internal/shared/ui"
	"rectablet/internal/shell"
	"rectablet/internal/workflows/swimming/curriculum"
	"rectablet/internal/workflows/swimming/domain"
	"rectablet/internal/workflows/swimming/navigation"
	"strings"
	"time"
)

type Controller struct {
	*navigation.Controller
	Store             Repository
	timedSelection    []string
	timedHeat         *timedHeatState
	timedLapsInput    string
	multiSelection    []string
	multiAttempts     []domain.MultiAttempt
	multiMode         bool
	multiSubskillPage int
	classAttempts     []domain.MultiAttempt
	classNames        map[string]string
}

func New(c *navigation.Controller, s Repository) *Controller {
	return &Controller{Controller: c, Store: s}
}

func (c *Controller) Build() (navigation.State, error) {
	f := c.Frame()
	state := c.InitialState()
	rows := []navigation.Row{}
	s := c.Store
	course, roster, e := navigation.LessonContext(s, f, c.Date, &state)
	if e != nil {
		return state, e
	}
	version := course.CurriculumVersion

	switch f.Screen {
	case navigation.Skills:
		state.Title = "Choose a Skill"
		catalog, e := s.CatalogFor(version)
		if e != nil {
			return state, e
		}
		seenLevels := map[string]bool{}
		for _, p := range roster {
			if !p.Active || seenLevels[p.Level] {
				continue
			}
			seenLevels[p.Level] = true
			level, ok := catalog.Level(p.Level)
			if !ok || len(level.Skills) == 0 {
				rows = append(rows, navigation.Row{ID: p.Level, Label: p.Level, Detail: "Missing curriculum · update swimmer level", Disabled: true})
				continue
			}
			for _, skill := range level.Skills {
				detail := level.Name
				if benchmark, ok := benchmarkFor(skill.ID); ok {
					detail += " · Timed heat · " + benchmark.Description
				}
				rows = append(rows, navigation.Row{ID: skill.ID, Label: skill.Number + ". " + curriculum.CompactTitle(skill), Detail: detail, Data: navigation.Parameters{Level: p.Level, Version: version}})
			}
		}
		state.Empty = "No enrolled swimmers with an assigned level."

	case navigation.SkillSwimmers:
		if f.Version != "" {
			version = f.Version
		}
		skill, e := s.Skill(version, f.Skill)
		if e != nil {
			return state, e
		}
		state.Title = "Choose a Swimmer"
		state.Subtitle = skill.Name
		state.Context += "\n" + skill.Name
		benchmark, timed := benchmarkFor(skill.ID)
		if timed {
			state.Title = "Select Swimmers"
			state.Context += "\n" + benchmark.Description
			for _, p := range roster {
				if !p.Active || p.Level != f.Level {
					continue
				}
				progress, e := s.Progress(p.ID, skill.ID, version)
				if e != nil {
					return state, e
				}
				selected := containsString(c.timedSelection, p.ID)
				row := navigation.Row{ID: p.ID, Label: p.Name, Status: "unselected", ProgressSegments: min(3, max(0, progress.Successes)), Data: navigation.Parameters{Skill: skill.ID, Level: p.Level, Version: version}}
				if selected {
					row.Status = "selected"
					row.Selected = true
					row.Fill = ui.Green
				}
				if progress.Successes > 0 {
					person, skillID, curriculumVersion := p.ID, skill.ID, version
					row.OnHold = func() { _ = c.RemoveLatestSuccess(person, skillID, curriculumVersion) }
					row.HoldDuration = 2 * time.Second
				}
				rows = append(rows, row)
			}
			if len(c.timedSelection) > 0 {
				state.HomeRailAction = saveRailAction(c.ConfirmTimedSelection)
				state.BackRailAction = discardRailAction(c.DiscardTimedSelection)
			}
			state.Empty = "No enrolled swimmers are assigned to this skill's level."
			break
		}
		state.HeaderSwitch = c.skillModeSwitch(skill, false)
		for _, p := range roster {
			if !p.Active || p.Level != f.Level {
				continue
			}
			progress, e := s.Progress(p.ID, skill.ID, version)
			if e != nil {
				return state, e
			}
			row := navigation.Row{
				ID:               p.ID,
				Label:            p.Name,
				ProgressSegments: min(3, max(0, progress.Successes)),
				Data:             navigation.Parameters{Skill: skill.ID, Level: p.Level, Version: version},
			}
			if progress.Successes > 0 {
				person, skillID, curriculumVersion := p.ID, skill.ID, version
				row.OnHold = func() { _ = c.RemoveLatestSuccess(person, skillID, curriculumVersion) }
				row.HoldDuration = 2 * time.Second
			}
			rows = append(rows, row)
		}
		state.Empty = "No enrolled swimmers are assigned to this skill's level."

	case navigation.MultiSelection:
		if f.Version != "" {
			version = f.Version
		}
		skill, e := s.Skill(version, f.Skill)
		if e != nil {
			return state, e
		}
		state.Title = "Choose Students"
		state.Subtitle = skill.Name
		state.Context += "\n" + skill.Name
		state.HeaderSwitch = c.skillModeSwitch(skill, true)
		for _, p := range roster {
			if !p.Active || p.Level != f.Level {
				continue
			}
			progress, e := s.Progress(p.ID, skill.ID, version)
			if e != nil {
				return state, e
			}
			selected := containsString(c.multiSelection, p.ID)
			row := navigation.Row{ID: p.ID, Label: p.Name, Detail: "", Selected: selected, ProgressSegments: min(3, max(0, progress.Successes)), Data: navigation.Parameters{Skill: skill.ID, Level: p.Level, Version: version}}
			if progress.Successes > 0 {
				person, skillID, curriculumVersion := p.ID, skill.ID, version
				row.OnHold = func() { _ = c.RemoveLatestSuccess(person, skillID, curriculumVersion) }
				row.HoldDuration = 2 * time.Second
			}
			rows = append(rows, row)
		}
		if len(c.multiSelection) > 0 {
			state.HomeRailAction = saveRailAction(c.ConfirmMultiSelection)
		}
		state.BackRailAction = discardRailAction(c.DiscardMultiSelection)
		state.Empty = "No enrolled swimmers are assigned to this skill's level."

	case navigation.ClassAssessment:
		return c.buildClassAssessmentState(course, f)

	case navigation.MultiAssessment:
		multiState, e := c.buildMultiAssessmentState(course, f)
		if e != nil {
			return state, e
		}
		state = multiState
		rows = append(rows, multiState.Rows...)

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
		state.Subtitle = skill.Name
		state.Context += "\n" + benchmark.Description
		state.Progress = c.timedHeatInput()
		state.HomeRailAction = saveRailAction(c.ConfirmTimedLaps)
		state.BackRailAction = discardRailAction(c.DiscardTimedSelection)

	case navigation.TimedHeat:
		if c.timedHeat == nil {
			return state, fmt.Errorf("timed heat is no longer available")
		}
		state.OnLeave = c.stopTimedTicker
		benchmark, ok := benchmarkFor(c.timedHeat.SkillID)
		if !ok {
			return state, fmt.Errorf("skill is not a timed benchmark")
		}
		state.Title = formatHeatElapsed(c.timedHeatElapsed())
		state.Subtitle = benchmark.Description
		state.Context += "\n" + benchmark.Description
		state.HeaderTitleFontSize = 34
		startLabel := "Start"
		if c.timedHeat.Running {
			startLabel = "Pause"
		}
		state.HeaderActions = []shell.HeaderAction{
			{Label: startLabel, Run: c.ToggleTimedTimer, Fill: ui.Green, Ink: ui.White, FontSize: 18},
			{Label: "Reset", Run: c.ResetTimedHeat, Fill: ui.Rail, Ink: ui.Blue, FontSize: 18},
		}
		state.HomeRailAction = saveRailAction(c.SaveTimedHeat)
		state.BackRailAction = discardRailAction(c.DiscardTimedHeat)
		state.Progress = benchmark.Description
		for i := range c.timedHeat.Participants {
			participant := &c.timedHeat.Participants[i]
			detail := fmt.Sprintf("%d/%d laps", participant.participant.LapsCompleted, participant.participant.LapsRequired)
			status := "in_progress"
			if participant.completed {
				detail += " · " + formatHeatElapsed(participant.completedAt)
				status = "successful"
			}
			rows = append(rows, navigation.Row{ID: "timed-lap:" + participant.participant.PersonID, Label: participant.participant.Name, Detail: detail, Status: status, ProgressValue: participant.participant.LapsCompleted, ProgressTotal: participant.participant.LapsRequired})
		}
		state.Empty = "No swimmers in this heat."

	case navigation.Attempts:
		p, e := s.GetEnrollment(f.Person)
		if e != nil {
			return state, e
		}
		if f.Version != "" {
			version = f.Version
		}
		skill, e := s.Skill(version, f.Skill)
		if e != nil {
			return state, e
		}
		state.Title = p.Name
		state.Subtitle = skill.Name
		state.Context += "\n" + state.Title + "\n" + skill.Name
		progress, e := s.Progress(f.Person, skill.ID, version)
		if e != nil {
			return state, e
		}
		state.Progress = navigation.ProgressText(progress)
		attempts, e := s.ListAttempts(f.Person, skill.ID, version)
		if e != nil {
			return state, e
		}
		for _, a := range attempts {
			label := "Not yet successful"
			if a.Success {
				label = "Successful"
			}
			if a.SubmittedAt == nil {
				label = "Unfinished attempt"
			} else if a.Editing {
				label += " · Edit saved"
			}
			created := a.CreatedAt
			clock := created
			if len(created) >= 19 {
				clock = created[11:19]
			}
			rows = append(rows, navigation.Row{ID: a.ID, Label: label, Detail: a.Date + " · " + clock + " UTC"})
		}
		state.Empty = "No attempts recorded for this skill."

	case navigation.Assessment:
		a, e := s.GetAttempt(f.Attempt)
		if e != nil {
			return state, e
		}
		p, e := s.GetEnrollment(a.EnrollmentID)
		if e != nil {
			return state, e
		}
		skill, e := s.Skill(a.CurriculumVersion, a.SkillID)
		if e != nil {
			return state, e
		}
		answers, e := s.Draft(f.Attempt)
		if e != nil {
			return state, e
		}
		lesson, e := s.GetLesson(a.LessonID)
		if e != nil {
			return state, e
		}
		state.Title = p.Name
		state.Subtitle = skill.Name
		state.Context = course.Name + " · " + lesson.Date + "\n" + state.Title + "\n" + skill.Name
		for _, subskill := range skill.Subskills {
			status := answers[subskill.ID]
			if status == "" {
				status = "unmarked"
			}
			rows = append(rows, navigation.Row{ID: subskill.ID, Label: subskill.Text, Detail: strings.Join(subskill.Group, " › "), Status: status})
		}
		if f.Review {
			state.FooterAction = "Save correction"
			state.CanSubmit = len(answers) == len(rows)
			state.PageSize = 1
			state.Progress = fmt.Sprintf("%d/%d assessed · %s", len(answers), len(rows), lesson.Date)
		} else {
			state.HomeRailAction = saveRailAction(c.SaveAttempt)
			state.BackRailAction = discardRailAction(c.DiscardAttempt)
			state.PageSize = 3
			state.Progress = fmt.Sprintf("%d/%d complete · %s", successfulAnswers(answers), len(rows), lesson.Date)
		}
	}

	c.Paginate(&state, rows)
	c.composeFeatureBody(&state)
	return state, nil
}

func successfulAnswers(answers map[string]string) int {
	count := 0
	for _, status := range answers {
		if status == "successful" {
			count++
		}
	}
	return count
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func removeString(values []string, target string) []string {
	out := values[:0]
	for _, value := range values {
		if value != target {
			out = append(out, value)
		}
	}
	return out
}

func (c *Controller) timedHeatInput() string {
	if c.timedLapsInput == "" {
		return "Enter laps needed, then press ✓"
	}
	return "Laps: " + c.timedLapsInput
}

func (c *Controller) ConfirmTimedSelection() error {
	if c.Frame().Screen != navigation.SkillSwimmers || len(c.timedSelection) == 0 {
		return nil
	}
	return c.Run(func() error {
		c.timedLapsInput = ""
		c.Push(navigation.TimedLaps, nil)
		return nil
	})
}

func (c *Controller) DiscardTimedSelection() error {
	if c.Frame().Screen != navigation.SkillSwimmers && c.Frame().Screen != navigation.TimedLaps {
		return nil
	}
	return c.Run(func() error {
		c.timedSelection = nil
		c.timedLapsInput = ""
		c.Pop()
		return nil
	})
}

func (c *Controller) ConfirmTimedLaps() error {
	if c.Frame().Screen != navigation.TimedLaps || len(c.timedSelection) == 0 {
		return nil
	}
	laps := 0
	if _, e := fmt.Sscanf(c.timedLapsInput, "%d", &laps); e != nil || laps < 1 || laps > 999 {
		return nil
	}
	return c.Run(func() error {
		f := c.Frame()
		participants, e := c.Store.NewTimedHeat(c.timedSelection, f.Lesson, f.Skill, laps)
		if e != nil {
			return nil
		}
		heat := &timedHeatState{SkillID: f.Skill, Version: f.Version, Lesson: f.Lesson, LapsRequired: laps, stop: make(chan struct{})}
		for _, participant := range participants {
			heat.Participants = append(heat.Participants, timedParticipantState{participant: timedParticipant{AttemptID: participant.AttemptID, PersonID: participant.PersonID, Name: participant.Name, LapsRequired: participant.LapsRequired}})
		}
		c.timedHeat = heat
		c.Push(navigation.TimedHeat, nil)
		return nil
	})
}

func (c *Controller) timedHeatElapsed() time.Duration {
	if c.timedHeat == nil {
		return 0
	}
	if c.timedHeat.Running {
		return time.Since(c.timedHeat.StartedAt)
	}
	return c.timedHeat.Elapsed
}

func (c *Controller) stopTimedTicker() {
	if c.timedHeat == nil || c.timedHeat.stop == nil {
		return
	}
	select {
	case <-c.timedHeat.stop:
	default:
		close(c.timedHeat.stop)
	}
	c.timedHeat.stop = nil
}

func (c *Controller) startTimedTicker(heat *timedHeatState) {
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
					if c.timedHeat == heat && heat.Running {
						_ = c.Refresh()
					}
				})
			case <-stop:
				return
			}
		}
	}()
}

func (c *Controller) ToggleTimedTimer() error {
	if c.Frame().Screen != navigation.TimedHeat || c.timedHeat == nil {
		return nil
	}
	return c.Run(func() error {
		if c.timedHeat.Running {
			c.timedHeat.Elapsed = c.timedHeatElapsed()
			c.timedHeat.Running = false
			c.stopTimedTicker()
			return nil
		}
		c.timedHeat.StartedAt = time.Now().Add(-c.timedHeat.Elapsed)
		c.timedHeat.Running = true
		c.startTimedTicker(c.timedHeat)
		return nil
	})
}

func (c *Controller) ResetTimedHeat() error {
	if c.Frame().Screen != navigation.TimedHeat || c.timedHeat == nil {
		return nil
	}
	return c.Run(func() error {
		wasRunning := c.timedHeat.Running
		if wasRunning {
			c.timedHeat.Elapsed = c.timedHeatElapsed()
		}
		c.stopTimedTicker()
		ids := []string{}
		for i := range c.timedHeat.Participants {
			participant := &c.timedHeat.Participants[i]
			ids = append(ids, participant.participant.AttemptID)
		}
		if e := c.Store.ResetTimedHeat(ids); e != nil {
			if wasRunning {
				c.timedHeat.StartedAt = time.Now().Add(-c.timedHeat.Elapsed)
				c.startTimedTicker(c.timedHeat)
			}
			return nil
		}
		c.timedHeat.Running = false
		c.timedHeat.Elapsed = 0
		for i := range c.timedHeat.Participants {
			participant := &c.timedHeat.Participants[i]
			participant.participant.LapsCompleted = 0
			participant.completed = false
			participant.completedAt = 0
		}
		return nil
	})
}

func (c *Controller) RecordTimedLap(personID string) error {
	if c.Frame().Screen != navigation.TimedHeat || c.timedHeat == nil || !c.timedHeat.Running {
		return nil
	}
	return c.Run(func() error {
		for i := range c.timedHeat.Participants {
			participant := &c.timedHeat.Participants[i]
			if participant.participant.PersonID != personID || participant.completed {
				continue
			}
			completed, e := c.Store.RecordTimedLap(participant.participant.AttemptID)
			if e != nil {
				return nil
			}
			participant.participant.LapsCompleted = completed
			if completed >= participant.participant.LapsRequired {
				participant.completed = true
				participant.completedAt = c.timedHeatElapsed()
			}
			break
		}
		return nil
	})
}

func (c *Controller) SaveTimedHeat() error {
	if c.Frame().Screen != navigation.TimedHeat || c.timedHeat == nil {
		return nil
	}
	return c.Run(func() error {
		benchmark, ok := benchmarkFor(c.timedHeat.SkillID)
		if !ok {
			return nil
		}
		wasRunning := c.timedHeat.Running
		elapsed := c.timedHeatElapsed()
		c.timedHeat.Running = false
		c.timedHeat.Elapsed = elapsed
		c.stopTimedTicker()
		results := make([]domain.TimedResult, 0, len(c.timedHeat.Participants))
		for _, participant := range c.timedHeat.Participants {
			participantElapsed := elapsed
			if participant.completed {
				participantElapsed = participant.completedAt
			}
			results = append(results, domain.TimedResult{AttemptID: participant.participant.AttemptID, LapsRequired: participant.participant.LapsRequired, LapsCompleted: participant.participant.LapsCompleted, ElapsedMS: participantElapsed.Milliseconds(), Success: participant.completed && participantElapsed <= benchmark.Limit})
		}
		if e := c.Store.SaveTimedHeat(results); e != nil {
			if wasRunning {
				c.timedHeat.StartedAt = time.Now().Add(-elapsed)
				c.timedHeat.Running = true
				c.startTimedTicker(c.timedHeat)
			}
			return nil
		}
		c.timedHeat = nil
		c.timedSelection = nil
		c.timedLapsInput = ""
		c.Pop()
		c.Pop()
		return nil
	})
}

func (c *Controller) DiscardTimedHeat() error {
	if c.Frame().Screen != navigation.TimedHeat || c.timedHeat == nil {
		return nil
	}
	return c.Run(func() error {
		ids := []string{}
		for _, participant := range c.timedHeat.Participants {
			ids = append(ids, participant.participant.AttemptID)
		}
		if e := c.Store.DiscardTimedHeat(ids); e != nil {
			return nil
		}
		c.stopTimedTicker()
		c.timedHeat = nil
		c.timedSelection = nil
		c.timedLapsInput = ""
		c.Pop()
		c.Pop()
		return nil
	})
}

func (c *Controller) Select(id string) error {
	f := c.Frame()
	var row navigation.Row
	for _, r := range c.State.Rows {
		if r.ID == id {
			row = r
			break
		}
	}
	switch f.Screen {
	case navigation.Skills:
		params := navigation.Destination(row)
		c.timedSelection = nil
		c.Push(navigation.SkillSwimmers, func(n *navigation.Frame) {
			n.Skill = id
			n.Level = params.Level
			n.Version = params.Version
		})

	case navigation.SkillSwimmers:
		if _, timed := benchmarkFor(f.Skill); timed {
			if containsString(c.timedSelection, id) {
				c.timedSelection = removeString(c.timedSelection, id)
			} else {
				c.timedSelection = append(c.timedSelection, id)
			}
			c.Saved("")
			return nil
		}
		// Skill taps are best-effort on the device. Failed reads or writes leave
		// the current list intact without surfacing an error state.
		params := navigation.Destination(row)
		version := params.Version
		if version == "" {
			version = f.Version
		}
		skillID := f.Skill
		if params.Skill != "" {
			skillID = params.Skill
		}
		progress, e := c.Store.Progress(id, skillID, version)
		if e != nil {
			return nil
		}
		if progress.Complete {
			c.Saved("")
			return nil
		}
		attempt, e := c.Store.NewAttempt(id, f.Lesson, skillID)
		if e != nil {
			return nil
		}
		c.Push(navigation.Assessment, func(n *navigation.Frame) {
			n.Attempt = attempt
			n.Person = id
			n.Skill = skillID
			n.Level = f.Level
			n.Version = version
		})

	case navigation.MultiSelection:
		if containsString(c.multiSelection, id) {
			c.multiSelection = removeString(c.multiSelection, id)
		} else {
			c.multiSelection = append(c.multiSelection, id)
		}
		c.Saved("")
		return nil

	case navigation.Attempts:
		if e := c.Store.EditAttempt(id); e != nil {
			return nil
		}
		c.Push(navigation.Assessment, func(n *navigation.Frame) { n.Attempt = id; n.Review = true })

	case navigation.Assessment:
		return c.toggleSubskill(id)

	case navigation.ClassAssessment:
		return c.toggleClassSubskill(id)

	case navigation.TimedLaps:
		if len(id) == len("timed-digit:")+1 && strings.HasPrefix(id, "timed-digit:") {
			c.timedLapsInput += strings.TrimPrefix(id, "timed-digit:")
			if len(c.timedLapsInput) > 3 {
				c.timedLapsInput = c.timedLapsInput[:3]
			}
			return nil
		}
		if id == "timed-clear" {
			c.timedLapsInput = ""
		}

	case navigation.TimedHeat:
		if strings.HasPrefix(id, "timed-lap:") {
			return c.RecordTimedLap(strings.TrimPrefix(id, "timed-lap:"))
		}
	}
	return nil
}

func (c *Controller) toggleSubskill(id string) error {
	f := c.Frame()
	status := "successful"
	for _, row := range c.State.Rows {
		if row.ID == id && row.Status == "successful" {
			status = "unmarked"
			break
		}
	}
	if e := c.Store.Mark(f.Attempt, id, status); e != nil {
		// Keep the checklist usable if a local write fails; the persisted draft
		// remains unchanged and the next refresh shows its previous state.
		return nil
	}
	c.Saved("")
	return nil
}

func (c *Controller) RemoveLatestSuccess(person, skill, version string) error {
	return c.Run(func() error {
		if e := c.Store.RemoveLatestSuccess(person, skill, version); e != nil {
			return nil
		}
		c.Saved("")
		return nil
	})
}

// SaveAttempt submits the current draft. Incomplete attempts are stored with
// success=false so they remain available for later feedback without counting
// toward skill progress.
func (c *Controller) SaveAttempt() error {
	if c.Frame().Screen != navigation.Assessment || c.Frame().Review {
		return nil
	}
	return c.Run(func() error {
		if _, e := c.Store.Submit(c.Frame().Attempt); e != nil {
			return nil
		}
		c.Pop()
		return nil
	})
}

// DiscardAttempt removes the unfinished draft and returns to the student list.
func (c *Controller) DiscardAttempt() error {
	if c.Frame().Screen != navigation.Assessment || c.Frame().Review {
		return nil
	}
	return c.Run(func() error {
		if e := c.Store.DiscardAttempt(c.Frame().Attempt); e != nil {
			return nil
		}
		c.Pop()
		return nil
	})
}

func (c *Controller) Header() error { return nil }

func (c *Controller) Mark(status string) error {
	if c.Frame().Screen != navigation.Assessment || len(c.State.Rows) == 0 {
		return nil
	}
	if e := c.Store.Mark(c.Frame().Attempt, c.State.Rows[0].ID, status); e != nil {
		return nil
	}
	c.Saved("")
	return nil
}

func (c *Controller) Submit() error {
	if c.Frame().Screen != navigation.Assessment {
		return nil
	}
	_, e := c.Store.Submit(c.Frame().Attempt)
	if e != nil {
		return nil
	}
	c.Saved("")
	c.Pop()
	return nil
}
