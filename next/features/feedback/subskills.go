package feedback

import (
	"fmt"
	"fyne.io/fyne/v2/container"
	"rectablet/src/shared/components/card"

	"rectablet/next/contracts"
	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/domain"
)

func configureSubskills(state *contracts.State, person domain.Enrollment, entries []domain.Feedback, skill, version string) []contracts.Row {
	state.Title = "Subskills · " + person.Name
	state.Empty = "No subskills missed across saved attempts."
	state.PageSize = 6
	state.AlwaysShowPagePosition = true
	rows := []contracts.Row{}
	for _, entry := range entries {
		if entry.Skill.ID != skill || entry.Version != version {
			continue
		}
		state.PrimaryInfo = entry.Skill.Name
		for _, miss := range entry.MissedSubskills {
			rows = append(rows, contracts.Row{
				ID: miss.Subskill.ID, Label: miss.Subskill.Text,
				Status: fmt.Sprintf("Missed %d of %d saved attempts", miss.Count, entry.Assessed),
			})
		}
	}
	return rows
}

// Six full-width cards keep the criterion beside its miss count.
func composeSubskills(state contracts.State) ui.BodyComposer {
	return func(surface *ui.Surface) {
		if len(state.Rows) == 0 {
			ui.Empty(surface, state.Empty)
			return
		}
		for i, row := range state.Rows {
			i, row := i, row
			bounds := func(w, h float32) (float32, float32, float32, float32) {
				x, y, width, height := surface.BodyBounds(w, h)
				cellHeight := (height - 5*ui.Gap) / 6
				return x, y + float32(i)*(cellHeight+ui.Gap), width, cellHeight
			}
			panel := card.New(container.NewWithoutLayout(), card.Panel)
			surface.AddNamed("feedback-subskill-card-"+row.ID, panel, bounds)
			criterion := ui.NewText(row.Label, 22, ui.Navy)
			criterion.Bold = true
			criterion.MinimumFont = 18
			surface.AddNamed("feedback-subskill-text-"+row.ID, criterion, func(w, h float32) (float32, float32, float32, float32) {
				x, y, width, height := bounds(w, h)
				return x + ui.ButtonPadding, y + ui.ButtonPadding, width*.72 - 2*ui.ButtonPadding, height - 2*ui.ButtonPadding
			})
			count := ui.NewText(row.Status, 18, ui.Blue)
			count.Bold = true
			surface.AddNamed("feedback-subskill-misses-"+row.ID, count, func(w, h float32) (float32, float32, float32, float32) {
				x, y, width, height := bounds(w, h)
				return x + width*.72, y + ui.ButtonPadding, width*.28 - ui.ButtonPadding, height - 2*ui.ButtonPadding
			})
		}
	}
}
