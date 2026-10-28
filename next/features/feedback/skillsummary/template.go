package skillsummary

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"rectablet/next/contracts"
	"rectablet/next/navigation"
	"rectablet/src/shared/components/card"
	"rectablet/src/shared/ui"
)

func (c *SkillSummary) Compose(state contracts.State) ui.BodyComposer {
	return func(surface *ui.Surface) {
		if state.RowCount == 0 {
			ui.Empty(surface, state.Empty)
			return
		}
		ui.Grid(surface, state.Rows, 1, max(1, state.PageSize), c.OnSelect)
	}
}
func (c *SkillSummary) ComposeLarge(data Snapshot, page ...int) ui.BodyComposer {
	return func(right *ui.Surface) {
		for i, label := range []string{"In progress", "Complete"} {
			i := i
			button := right.Button("feedback-filter:"+label, label, func() {
				if c.OnFilter != nil {
					c.OnFilter(i == 1)
				}
			})
			button.Selectable = true
			button.Selected = (i == 1) == c.Complete
			button.FontSize = 18
			right.AddNamed("feedback-filter-"+label, button, func(w, h float32) (float32, float32, float32, float32) {
				x, y, width, _ := right.BodyBounds(w, h)
				cell := (width - ui.Gap) / 2
				return x + float32(i)*(cell+ui.Gap), y, cell, 40
			})
		}

		if len(data.Entries) == 0 {
			message := "No assessed skills in progress. Unassessed skills are not included here."
			if c.Complete {
				message = "No completed skills yet."
			}
			empty := ui.NewText(message, 21, ui.Navy)
			right.AddNamed("feedback-no-skills", empty, func(w, h float32) (float32, float32, float32, float32) {
				x, y, width, height := right.BodyBounds(w, h)
				inset := float32(48)
				return x, y + inset, width, height - inset
			})
			return
		}
		pageData := data
		if len(page) > 0 {
			start := page[0] * 2
			pageData.Entries = data.Entries[start:min(start+2, len(data.Entries))]
		}
		cards := newFeedbackCards(right, c, pageData)
		right.AddNamed("feedback-skill-cards", cards, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := right.BodyBounds(w, h)
			inset := float32(48)
			return x, y + inset, width, height - inset
		})
	}
}

// Text never contributes to card size; the page owns two equal slots.
type feedbackCardLayout struct{ scale float32 }

func (l *feedbackCardLayout) MinSize([]fyne.CanvasObject) fyne.Size { return fyne.NewSize(1, 1) }
func (l *feedbackCardLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	s := max(float32(.01), l.scale)
	pad := 12 * s
	titleHeight := 56 * s
	textTop := 8*s + titleHeight + 4*s + 30*s + 8*s
	textHeight := max(1, size.Height-72*s-textTop)
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(size)
	objects[1].Move(fyne.NewPos(pad, 8*s))
	objects[1].Resize(fyne.NewSize(size.Width-2*pad, titleHeight))
	objects[2].Move(fyne.NewPos(pad, 8*s+titleHeight+4*s))
	objects[2].Resize(fyne.NewSize(size.Width-2*pad, 30*s))
	objects[3].Move(fyne.NewPos(pad, 8*s+titleHeight+4*s+30*s+8*s))
	objects[3].Resize(fyne.NewSize(size.Width-2*pad, textHeight))
	objects[4].Move(fyne.NewPos(pad, size.Height-60*s))
	buttonWidth := max(1, (size.Width-2*pad-8*s)/2)
	objects[4].Resize(fyne.NewSize(buttonWidth, 48*s))
	objects[5].Move(fyne.NewPos(pad+buttonWidth+8*s, size.Height-60*s))
	objects[5].Resize(fyne.NewSize(buttonWidth, 48*s))
}

type feedbackStackLayout struct {
	cards []*feedbackCardLayout
	scale float32
}

func (l *feedbackStackLayout) MinSize([]fyne.CanvasObject) fyne.Size { return fyne.NewSize(1, 1) }
func (l *feedbackStackLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	y := float32(0)
	for i, object := range objects {
		card := l.cards[i]
		card.scale = l.scale
		h := (size.Height - 8*l.scale) / 2
		object.Move(fyne.NewPos(0, y))
		object.Resize(fyne.NewSize(size.Width, h))
		y += h + 8*l.scale
	}
}

type feedbackCards struct {
	widget.BaseWidget
	content *fyne.Container
	stack   *feedbackStackLayout
	labels  []*ui.TextBlock
	buttons []*ui.TouchButton
}

// The host draws widgets and containers; this wrapper must remain a registered widget.
var _ fyne.Widget = (*feedbackCards)(nil)

func (s *feedbackCards) SetScale(scale float32) {
	if scale <= 0 {
		scale = 1
	}
	s.stack.scale = scale
	for _, label := range s.labels {
		label.SetScale(scale)
	}
	for _, button := range s.buttons {
		button.SetScale(scale)
	}
	s.content.Refresh()
}
func (s *feedbackCards) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(s.content)
}
func newFeedbackCards(surface *ui.Surface, c *SkillSummary, data Snapshot) *feedbackCards {
	stack := &feedbackStackLayout{scale: 1}
	result := &feedbackCards{stack: stack}
	objects := make([]fyne.CanvasObject, 0, len(data.Entries))
	for _, entry := range data.Entries {
		entry := entry
		details := c.Detail(entry)
		layout := &feedbackCardLayout{scale: 1}
		stack.cards = append(stack.cards, layout)
		bg := card.New(container.NewWithoutLayout(), card.Panel)
		title := ui.NewText(entry.Skill.Name, 23, ui.Navy)
		title.Bold = true
		title.MaxLines = 2
		progress := ui.NewText(navigation.ProgressText(entry.Progress), 18, ui.Green)
		body := ui.NewText(details, 18, ui.Navy)
		result.labels = append(result.labels, title, progress, body)
		button := surface.Button("feedback-attempts:"+entry.Skill.ID+":"+entry.Version, "Attempts", func() {
			c.OnAttempts(data.Person.ID, entry.Skill.ID, entry.Version)
		})
		button.FontSize = 18
		button.Fill = ui.Rail
		button.Ink = ui.Blue
		subskills := surface.Button("feedback-subskills:"+entry.Skill.ID+":"+entry.Version, "Subskills", func() {
			if c.OnSubskills != nil {
				c.OnSubskills(data.Person.ID, entry.Skill.ID, entry.Version)
			}
		})
		subskills.FontSize = 18
		subskills.Fill = ui.Rail
		subskills.Ink = ui.Blue
		result.buttons = append(result.buttons, button, subskills)
		objects = append(objects, container.New(layout, bg, title, progress, body, button, subskills))
	}
	result.content = container.New(stack, objects...)
	result.ExtendBaseWidget(result)
	return result
}
