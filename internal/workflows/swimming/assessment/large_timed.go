package assessment

import (
	"strconv"

	"rectablet/internal/shared/ui"
	"rectablet/internal/workflows/swimming/navigation"
)

func (c *Controller) configureLargeTimedSetup(state *navigation.State, rows []navigation.Row) {
	state.Title = "Set Up Timed Heat"
	state.PageSize = max(1, len(rows))
	state.HomeRailAction = nil
	state.BackRailAction = discardRailAction(c.DiscardTimedSelection)
	state.Progress = c.timedHeatInput()
	state.ProgressInBody = true
	state.ComposeBody = c.composeLargeTimedSetup(*state, rows)
}

func (c *Controller) composeLargeTimedSetup(state navigation.State, rows []navigation.Row) ui.BodyComposer {
	return func(surface *ui.Surface) {
		left, right := largeSplit(surface, .5)
		c.largePagedGrid(left, "timed-setup-swimmers", "Select swimmers", rows, 2, 6, &c.largeRosterPage, func(id string) { _ = c.Controller.Select(id) })
		c.largeLapKeypad(right)
	}
}

func (c *Controller) largeLapKeypad(surface *ui.Surface) {
	display := ui.NewText("Laps needed: "+c.timedLapsInput, 24, ui.Navy)
	display.Bold, display.Center = true, true
	surface.AddNamed("timed-laps-display", display, func(w, h float32) (float32, float32, float32, float32) {
		x, y, width, _ := surface.BodyBounds(w, h)
		return x, y, width, 48
	})
	for index, digit := range []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"} {
		index, digit := index, digit
		button := surface.Button("timed-digit:"+digit, digit, func() {
			_ = c.Run(func() error {
				if len(c.timedLapsInput) < 3 {
					c.timedLapsInput += digit
				}
				return nil
			})
		})
		button.FontSize = 28
		surface.AddNamed("timed-digit-"+digit, button, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := surface.BodyBounds(w, h)
			cellWidth := (width - 2*ui.Gap) / 3
			cellHeight := (height - 2*52 - 3*ui.Gap) / 4
			return x + float32(index%3)*(cellWidth+ui.Gap), y + 52 + float32(index/3)*(cellHeight+ui.Gap), cellWidth, cellHeight
		})
	}
	clear := surface.Button("timed-clear", "Clear", func() { _ = c.Run(func() error { c.timedLapsInput = ""; return nil }) })
	clear.Fill, clear.Ink = ui.Rail, ui.Blue
	surface.AddNamed("timed-clear", clear, func(w, h float32) (float32, float32, float32, float32) {
		x, y, width, height := surface.BodyBounds(w, h)
		cellWidth := (width - 2*ui.Gap) / 3
		cellHeight := (height - 2*52 - 3*ui.Gap) / 4
		return x + cellWidth + ui.Gap, y + 52 + 3*(cellHeight+ui.Gap), 2*cellWidth + ui.Gap, cellHeight
	})
	start := surface.Button("timed-create-heat", "Continue to heat", func() { _ = c.startLargeTimedHeat() })
	laps, _ := strconv.Atoi(c.timedLapsInput)
	start.Disabled = len(c.timedSelection) == 0 || laps < 1 || laps > 999
	start.Fill = ui.Green
	surface.AddNamed("timed-create-heat", start, func(w, h float32) (float32, float32, float32, float32) {
		x, y, width, height := surface.BodyBounds(w, h)
		return x, y + height - 48, width, 48
	})
}

func (c *Controller) startLargeTimedHeat() error {
	laps, err := strconv.Atoi(c.timedLapsInput)
	if err != nil || laps < 1 || laps > 999 || len(c.timedSelection) == 0 {
		return nil
	}
	if c.Frame().Screen != navigation.SkillSwimmers && c.Frame().Screen != navigation.TimedLaps {
		return nil
	}
	return c.Run(func() error {
		f := c.Frame()
		participants, err := c.Store.NewTimedHeat(c.timedSelection, f.Lesson, f.Skill, laps)
		if err != nil {
			return err
		}
		heat := &timedHeatState{SkillID: f.Skill, Version: f.Version, Lesson: f.Lesson, LapsRequired: laps}
		for _, person := range participants {
			heat.Participants = append(heat.Participants, timedParticipantState{participant: timedParticipant{
				AttemptID: person.AttemptID, PersonID: person.PersonID, Name: person.Name, LapsRequired: person.LapsRequired,
			}})
		}
		c.timedHeat, c.largeHeatPage = heat, 0
		c.Push(navigation.TimedHeat, nil)
		return nil
	})
}

// TimedLaps remains reachable when a draft flow was opened in compact mode.
func (c *Controller) composeLargeTimedLapsBody(state navigation.State) ui.BodyComposer {
	return func(surface *ui.Surface) {
		left, right := largeSplit(surface, .5)
		var swimmers []ui.Item
		for _, id := range c.timedSelection {
			student, err := c.Store.GetEnrollment(id)
			if err != nil {
				continue
			}
			swimmers = append(swimmers, ui.Item{ID: id, Label: student.Name, Selected: true, Fill: ui.Green})
		}
		c.largePagedGrid(left, "timed-setup-swimmers", "Selected swimmers", swimmers, 1, 6, &c.largeRosterPage, func(string) {})
		c.largeLapKeypad(right)
	}
}

func (c *Controller) composeLargeTimedHeatBody(state navigation.State) ui.BodyComposer {
	return func(surface *ui.Surface) {
		rows := append([]ui.Item(nil), state.Rows...)
		for i := range rows {
			if rows[i].Status == "successful" {
				rows[i].Fill, rows[i].Selected = ui.Green, true
			}
		}
		c.largePagedGrid(surface, "timed-heat-swimmers", "Tap a swimmer to record a lap", rows, 2, 6, &c.largeHeatPage, func(id string) { _ = c.Controller.Select(id) })
	}
}
