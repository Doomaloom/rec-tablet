package large

import (
	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/navigation"
)

func (c *View) ComposeLargeChecklistBody(rows []navigation.Row, selectRow func(string)) ui.BodyComposer {
	return func(surface *ui.Surface) { c.RenderLargeChecklist(surface, rows, selectRow) }
}

// Every checklist criterion is visible at once. Twelve criteria use two columns
// of six, leaving at least a 48 pixel target in the large shell body.

func (c *View) RenderLargeChecklist(surface *ui.Surface, rows []navigation.Row, selectRow func(string)) {
	if len(rows) == 0 {
		ui.Empty(surface, "No criteria available.")
		return
	}
	columns := 1
	if len(rows) > 6 {
		columns = 2
	}
	visualRows := (len(rows) + columns - 1) / columns
	var lastWidth, lastHeight, sharedFont float32
	for i, row := range rows {
		i, row := i, row
		button := surface.Button("criterion:"+row.ID, row.Label, func() {
			if selectRow != nil {
				selectRow(row.ID)
			}
		})
		button.Compact, button.MaxLabelLines = true, -1
		button.Fill = ui.Rail
		button.Ink = ui.Navy
		if row.Status == "successful" {
			button.Fill, button.Ink, button.Selected = ui.Green, ui.White, true
		}
		if row.Status == "not_yet" {
			button.Fill, button.Ink = ui.Amber, ui.White
		}
		surface.AddNamed("large-criterion-"+row.ID, button, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := surface.BodyBounds(w, h)
			cellWidth := (width - float32(columns-1)*ui.Gap) / float32(columns)
			cellHeight := (height - float32(visualRows-1)*ui.Gap) / float32(visualRows)
			if width != lastWidth || height != lastHeight {
				sharedFont = fitLargeChecklistFont(rows, cellWidth, cellHeight)
				lastWidth, lastHeight = width, height
			}
			button.FontSize = sharedFont
			return x + float32(i%columns)*(cellWidth+ui.Gap), y + float32(i/columns)*(cellHeight+ui.Gap), cellWidth, cellHeight
		})
	}
}

func fitLargeChecklistFont(rows []navigation.Row, cellWidth, cellHeight float32) float32 {
	for font := float32(20); font >= 14; font-- {
		fits := true
		for _, row := range rows {
			if ui.WrappedTextHeight(row.Label, cellWidth-2*ui.ButtonPadding, font, true)+4 > cellHeight {
				fits = false
				break
			}
		}
		if fits {
			return font
		}
	}
	return 14
}
