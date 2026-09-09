package assessment

import (
	"fyne.io/fyne/v2/canvas"
	"image/color"
	"rectablet/internal/shared/ui"
)

func renderCriterion(surface *ui.Surface, row ui.Item, mark func(string)) {

	box := canvas.NewRectangle(ui.White)
	box.CornerRadius = 20
	surface.AddNamed("assessment-card", box, func(w, h float32) (float32, float32, float32, float32) { return 100, 118, w - 200, h - 200 })
	group := row.Detail
	if group == "" {
		group = "Keep working on"
		if mark != nil {
			group = "Assess this subskill"
		}
	}
	groupText := ui.NewText(group, 17, ui.Muted)
	surface.AddNamed("assessment-group", groupText, func(w, h float32) (float32, float32, float32, float32) { return 118, 132, w - 236, 38 })
	criterion := ui.NewText(row.Label, 25, ui.Navy)
	criterion.Bold = true
	criterion.MinimumFont = 18
	surface.Criterion = criterion
	surface.AddNamed("assessment-criterion", criterion, func(w, h float32) (float32, float32, float32, float32) {
		height := h - 260
		if mark != nil {
			height -= 92
		}
		return 118, 176, w - 236, height
	})
	if mark != nil {
		yes := surface.Button("successful", "Successful", func() { mark("successful") })
		yes.Fill = color.NRGBA{R: 229, G: 241, B: 238, A: 255}
		yes.Ink = ui.Green
		if row.Status == "successful" {
			yes.Detail = "Selected"
			yes.Fill = ui.Green
			yes.Ink = ui.White
			yes.Selected = true
		}
		surface.AddNamed("assessment-successful", yes, func(w, h float32) (float32, float32, float32, float32) { return 114, h - 164, (w - 238) / 2, 68 })
		no := surface.Button("not_yet", "Not yet", func() { mark("not_yet") })
		no.Fill = color.NRGBA{R: 248, G: 237, B: 223, A: 255}
		no.Ink = ui.Amber
		if row.Status == "not_yet" {
			no.Detail = "Selected"
			no.Fill = ui.Amber
			no.Ink = ui.White
			no.Selected = true
		}
		surface.AddNamed("assessment-not-yet", no, func(w, h float32) (float32, float32, float32, float32) {
			return 124 + (w-238)/2, h - 164, (w - 238) / 2, 68
		})
	}
}
