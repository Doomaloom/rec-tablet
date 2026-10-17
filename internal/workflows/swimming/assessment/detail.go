package assessment

import (
	"fyne.io/fyne/v2/container"
	"image/color"
	"rectablet/internal/shared/components/card"
	"rectablet/internal/shared/ui"
)

func renderCriterion(surface *ui.Surface, row ui.Item, mark func(string)) {

	box := card.New(container.NewWithoutLayout(), card.Panel)
	surface.AddNamed("assessment-card", box, surface.BodyBounds)
	criterion := ui.NewText(row.Label, 25, ui.Navy)
	criterion.Bold = true
	criterion.MinimumFont = 18
	surface.Criterion = criterion
	surface.AddNamed("assessment-criterion", criterion, func(w, h float32) (float32, float32, float32, float32) {
		x, y, width, height := surface.BodyBounds(w, h)
		height -= 2 * ui.ButtonPadding
		if mark != nil {
			height -= ui.HeaderHeight + ui.ButtonPadding
		}
		return x + ui.ButtonPadding, y + ui.ButtonPadding, width - 2*ui.ButtonPadding, height
	})
	if mark != nil {
		yes := surface.Button("successful", "Successful", func() { mark("successful") })
		yes.Fill = color.NRGBA{R: 229, G: 241, B: 238, A: 255}
		yes.Ink = ui.Green
		if row.Status == "successful" {
			yes.Fill = ui.Green
			yes.Ink = ui.White
			yes.Selected = true
		}
		surface.AddNamed("assessment-successful", yes, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := surface.BodyBounds(w, h)
			return x + ui.ButtonPadding, y + height - ui.ButtonPadding - ui.HeaderHeight, (width - 3*ui.ButtonPadding) / 2, ui.HeaderHeight
		})
		no := surface.Button("not_yet", "Not yet", func() { mark("not_yet") })
		no.Fill = color.NRGBA{R: 248, G: 237, B: 223, A: 255}
		no.Ink = ui.Amber
		if row.Status == "not_yet" {
			no.Fill = ui.Amber
			no.Ink = ui.White
			no.Selected = true
		}
		surface.AddNamed("assessment-not-yet", no, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := surface.BodyBounds(w, h)
			buttonWidth := (width - 3*ui.ButtonPadding) / 2
			return x + 2*ui.ButtonPadding + buttonWidth, y + height - ui.ButtonPadding - ui.HeaderHeight, buttonWidth, ui.HeaderHeight
		})
	}
}
