package feedback

import (
	"fyne.io/fyne/v2/container"
	"rectablet/src/shared/components/card"
	"rectablet/src/shared/ui"
)

func renderDetail(surface *ui.Surface, row ui.Item) {

	box := card.New(container.NewWithoutLayout(), card.Panel)
	surface.AddNamed("feedback-detail-card", box, surface.BodyBounds)
	criterion := ui.NewText(row.Label, 25, ui.Navy)
	criterion.Bold = true
	criterion.MinimumFont = 18
	surface.Criterion = criterion
	surface.AddNamed("feedback-detail-criterion", criterion, func(w, h float32) (float32, float32, float32, float32) {
		x, y, width, height := surface.BodyBounds(w, h)
		height -= 2 * ui.ButtonPadding
		return x + ui.ButtonPadding, y + ui.ButtonPadding, width - 2*ui.ButtonPadding, height
	})
}
