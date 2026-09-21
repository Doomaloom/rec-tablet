package feedback

import (
	"fyne.io/fyne/v2/canvas"
	"rectablet/internal/shared/ui"
)

func renderDetail(surface *ui.Surface, row ui.Item) {

	box := canvas.NewRectangle(ui.White)
	box.CornerRadius = 0
	surface.AddNamed("feedback-detail-card", box, surface.BodyBounds)
	groupHeight := float32(0)
	if row.Detail != "" {
		groupHeight = 28
		groupText := ui.NewText(row.Detail, 17, ui.Muted)
		surface.AddNamed("feedback-detail-group", groupText, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, _ := surface.BodyBounds(w, h)
			return x + ui.ButtonPadding, y + ui.ButtonPadding, width - 2*ui.ButtonPadding, groupHeight
		})
	}
	criterion := ui.NewText(row.Label, 25, ui.Navy)
	criterion.Bold = true
	criterion.MinimumFont = 18
	surface.Criterion = criterion
	surface.AddNamed("feedback-detail-criterion", criterion, func(w, h float32) (float32, float32, float32, float32) {
		x, y, width, height := surface.BodyBounds(w, h)
		height -= 2*ui.ButtonPadding + groupHeight
		return x + ui.ButtonPadding, y + ui.ButtonPadding + groupHeight, width - 2*ui.ButtonPadding, height
	})
}
