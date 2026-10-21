package checklist

import (
	"fmt"

	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
)

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

// Every checklist criterion is visible at once. Twelve criteria use two columns
// of six, leaving at least a 48 pixel target in the large shell body.
func (c Controller) RenderGrid(surface *ui.Surface, rows []navigation.Row, selectRow func(string)) {
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

// Divide the page into equal rows and use the largest shared font that fits every label.
func fitClassChecklist(rows []navigation.Row, width, height float32) checklistLayout {
	if len(rows) == 0 {
		return checklistLayout{}
	}
	available := height - float32(len(rows)-1)*classChecklistGap
	layout := checklistLayout{height: available / float32(len(rows)), font: 22}
	for _, row := range rows {
		// Allow compact button padding plus one pixel for fractional-scale rounding.
		for layout.font > 1 && ui.WrappedTextHeight(row.Label, width-2*ui.ButtonPadding, layout.font, true)+5 > layout.height {
			layout.font--
		}
	}
	return layout
}
func (c Controller) ComposeCompact(rows []navigation.Row) ui.BodyComposer {
	return func(surface *ui.Surface) {
		var layout checklistLayout
		var lastWidth, lastHeight float32
		for i, row := range rows {
			index, row := i, row
			button := surface.Button("row:"+row.ID, row.Label, func() { _ = c.Actions.Select(row.ID) })
			button.Compact, button.MaxLabelLines = true, -1
			if index%2 == 1 {
				button.Fill = ui.Navy
			}
			if row.Status == "successful" {
				button.Fill = ui.Green
			}
			surface.AddNamed(fmt.Sprintf("class-subskill-%d", index), button, func(width, height float32) (float32, float32, float32, float32) {
				x, top, bodyWidth, bodyHeight := surface.BodyBounds(width, height)
				if width != lastWidth || height != lastHeight {
					layout = fitClassChecklist(rows, bodyWidth, bodyHeight)
					lastWidth, lastHeight = width, height
				}
				button.FontSize = layout.font
				y := top + float32(index)*(layout.height+classChecklistGap)
				return x, y, bodyWidth, layout.height
			})
		}
	}
}

type checklistLayout struct {
	height float32
	font   float32
}

const (
	classChecklistGap    float32 = 2
	classChecklistBottom float32 = ui.OuterMargin
)
