package lessonplan

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"rectablet/next/contracts"
	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/transfer"
)

func (c *Component) compose(state *contracts.State, selected *transfer.PlannedRow, marking navigation.Row, choices []navigation.Row, reason string) {
	snapshot := *state
	large := state.DisplayMode.IsLarge()
	state.ComposeBody = func(surface *ui.Surface) {
		if !large {
			ui.Empty(surface, "Lesson plans are available in the large layout.")
			return
		}
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}
		left := surface
		if selected != nil {
			left = ui.Panel(surface, func(w, h float32) (float32, float32, float32, float32) {
				x, y, width, height := surface.BodyBounds(w, h)
				return x, y, (width - ui.Gap) / 3, height
			})
		}
		ui.Grid(ui.SelectableSurface(left), snapshot.Rows, 1, snapshot.PageSize, func(id string) { _ = c.nav.Select(id) })
		if selected == nil {
			return
		}
		right := ui.Panel(surface, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := surface.BodyBounds(w, h)
			offset := (width-ui.Gap)/3 + ui.Gap
			return x + offset, y, width - offset, height
		})
		equipment := "Not specified"
		if len(selected.Equipment) > 0 {
			equipment = strings.Join(selected.Equipment, ", ")
		}
		activity := selected.Activity
		if strings.TrimSpace(activity) == "" {
			activity = "Not specified"
		}
		details := fmt.Sprintf("%s\n\nPlanned location\n%s\n\nEquipment required\n%s\n\nDuration\n%d minutes\n\nDrills / activities\n%s", selected.Skill, selected.Location, equipment, selected.Duration, activity)
		if reason != "" {
			details += "\n\n" + reason
		}
		choiceHeight := float32(0)
		if len(choices) > 1 {
			choiceHeight = float32(len(choices))*52 + ui.Gap
			details += "\n\nChoose the component to mark below."
		}
		scroll := newDetailScroll(details)
		right.AddNamed("lesson-plan-details", scroll, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := right.BodyBounds(w, h)
			return x, y, width, height - 64 - ui.Gap - choiceHeight
		})

		if len(choices) > 1 {
			for i, choice := range choices {
				i, choice := i, choice
				option := right.Button("lesson-plan-component:"+choice.ID, choice.Label, func() {
					_ = c.chooseComponent(choice.ID)
				})
				option.Selectable = true
				option.Selected = choice.ID == marking.ID
				option.FontSize = 18
				right.AddNamed("lesson-plan-component:"+choice.ID, option, func(w, h float32) (float32, float32, float32, float32) {
					x, y, width, height := right.BodyBounds(w, h)
					return x, y + height - 64 - ui.Gap - choiceHeight + float32(i)*52, width, 48
				})
			}
		}
		button := right.Button("lesson-plan-start-marking", "Start marking", func() { _ = c.start(marking) })
		button.Disabled = reason != ""
		right.AddNamed("lesson-plan-start-marking", button, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := right.BodyBounds(w, h)
			return x, y + height - 64, width, 64
		})
	}
}

// This scroll measures wrapped text at the actual pane width and display scale.
type detailLayout struct {
	text         string
	width, scale float32
}

func (l *detailLayout) MinSize([]fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(1, ui.WrappedTextHeight(l.text, max(1, l.width-24*l.scale), 20*l.scale, false)+24*l.scale)
}
func (l *detailLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	p := 12 * l.scale
	objects[0].Move(fyne.NewPos(p, p))
	objects[0].Resize(fyne.NewSize(max(1, size.Width-2*p), max(1, size.Height-2*p)))
}

type detailScroll struct {
	*container.Scroll
	layout *detailLayout
	label  *ui.TextBlock
}

func newDetailScroll(text string) *detailScroll {
	l := &detailLayout{text: text, width: 500, scale: 1}
	label := ui.NewText(text, 20, ui.Navy)
	return &detailScroll{Scroll: container.NewVScroll(container.New(l, label)), layout: l, label: label}
}
func (s *detailScroll) SetScale(scale float32) {
	s.layout.scale = scale
	s.label.SetScale(scale)
	s.Content.Refresh()
}
func (s *detailScroll) Resize(size fyne.Size) {
	s.layout.width = size.Width
	s.Scroll.Resize(size)
	s.Content.Refresh()
}
