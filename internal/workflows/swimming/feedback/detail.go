package feedback

import (
	"fyne.io/fyne/v2/canvas"
	"rectablet/internal/shared/ui"
)

func renderDetail(surface *ui.Surface, row ui.Item) {

	box := canvas.NewRectangle(ui.White)
	box.CornerRadius = 20
	surface.AddNamed("feedback-detail-card", box, func(w, h float32) (float32, float32, float32, float32) { return 100, 118, w - 200, h - 200 })
	group := row.Detail
	if group == "" {
		group = "Keep working on"

	}
	groupText := ui.NewText(group, 17, ui.Muted)
	surface.AddNamed("feedback-detail-group", groupText, func(w, h float32) (float32, float32, float32, float32) { return 118, 132, w - 236, 38 })
	criterion := ui.NewText(row.Label, 25, ui.Navy)
	criterion.Bold = true
	criterion.MinimumFont = 18
	surface.Criterion = criterion
	surface.AddNamed("feedback-detail-criterion", criterion, func(w, h float32) (float32, float32, float32, float32) {
		height := h - 260

		return 118, 176, w - 236, height
	})
}
