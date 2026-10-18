package small

import (
	"fmt"
	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/navigation"
)

type checklistLayout struct {
	height float32
	font   float32
}

const (
	classChecklistGap    float32 = 2
	classChecklistBottom float32 = ui.OuterMargin
)

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

func (c *View) ComposeClassChecklist(rows []navigation.Row) ui.BodyComposer {
	return func(surface *ui.Surface) {
		var layout checklistLayout
		var lastWidth, lastHeight float32
		for i, row := range rows {
			index, row := i, row
			button := surface.Button("row:"+row.ID, row.Label, func() { _ = c.Controller.Select(row.ID) })
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
