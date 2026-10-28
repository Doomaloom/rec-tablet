package feedback

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"rectablet/src/shared/components/card"
	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/navigation"
)

func (c *SkillSummary) ComposeLarge(data summaryData, page int) ui.BodyComposer {
	return func(right *ui.Surface) {
		heading := data.person.Name
		if !data.person.Active {
			heading += " · Former enrollment"
		}
		title := ui.NewText(heading, 26, ui.Navy)
		title.Bold = true
		selectedTitleHeight := func(w, h float32) float32 {
			_, _, width, _ := right.BodyBounds(w, h)
			return max(38, ui.WrappedTextHeight(heading, width, 26, true)+4)
		}
		right.AddNamed("feedback-selected-swimmer", title, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, _ := right.BodyBounds(w, h)
			return x, y, width, selectedTitleHeight(w, h)
		})
		summary := ui.NewText(fmt.Sprintf("%d assessed skills in progress", len(data.entries)), 18, ui.Muted)
		right.AddNamed("feedback-selected-summary", summary, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, _ := right.BodyBounds(w, h)
			return x, y + selectedTitleHeight(w, h), width, 28
		})
		if len(data.entries) == 0 {
			empty := ui.NewText("No assessed skills remain incomplete. Unassessed skills are not included here.", 21, ui.Navy)
			right.AddNamed("feedback-no-skills", empty, func(w, h float32) (float32, float32, float32, float32) {
				x, y, width, height := right.BodyBounds(w, h)
				inset := selectedTitleHeight(w, h) + 40
				return x, y + inset, width, height - inset
			})
			return
		}
		pageData := data
		pageData.entries = data.entries[page : page+1]
		scroll := newFeedbackScroll(right, c, pageData)
		right.AddNamed("feedback-skill-scroll", scroll, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := right.BodyBounds(w, h)
			inset := selectedTitleHeight(w, h) + 32
			return x, y + inset, width, height - inset
		})
	}
}

type feedbackCardLayout struct {
	title        string
	details      string
	width, scale float32
}

func (l *feedbackCardLayout) titleHeight(width float32) float32 {
	s := max(float32(.01), l.scale)
	return max(32*s, ui.WrappedTextHeight(l.title, max(1, width-24*s), 23*s, true)+4*s)
}

func (l *feedbackCardLayout) cardHeight() float32 {
	s := max(float32(.01), l.scale)
	width := max(float32(100), l.width-28*s)
	textHeight := ui.WrappedTextHeight(l.details, width, 18*s, false) + 6*s
	return 8*s + l.titleHeight(l.width) + 4*s + 30*s + 8*s + textHeight + 12*s + 48*s + 12*s
}

func (l *feedbackCardLayout) MinSize([]fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(1, l.cardHeight())
}
func (l *feedbackCardLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	s := max(float32(.01), l.scale)
	pad := 12 * s
	titleHeight := l.titleHeight(size.Width)
	textHeight := ui.WrappedTextHeight(l.details, max(100, size.Width-2*pad), 18*s, false) + 6*s
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(size)
	objects[1].Move(fyne.NewPos(pad, 8*s))
	objects[1].Resize(fyne.NewSize(size.Width-2*pad, titleHeight))
	objects[2].Move(fyne.NewPos(pad, 8*s+titleHeight+4*s))
	objects[2].Resize(fyne.NewSize(size.Width-2*pad, 30*s))
	objects[3].Move(fyne.NewPos(pad, 8*s+titleHeight+4*s+30*s+8*s))
	objects[3].Resize(fyne.NewSize(size.Width-2*pad, textHeight))
	objects[4].Move(fyne.NewPos(pad, size.Height-60*s))
	objects[4].Resize(fyne.NewSize(max(1, size.Width-2*pad), 48*s))
}

type feedbackStackLayout struct {
	cards        []*feedbackCardLayout
	width, scale float32
}

func (l *feedbackStackLayout) MinSize([]fyne.CanvasObject) fyne.Size {
	height := float32(0)
	for _, card := range l.cards {
		card.width, card.scale = l.width, l.scale
		height += card.cardHeight() + 8*l.scale
	}
	return fyne.NewSize(1, max(1, height))
}
func (l *feedbackStackLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	y := float32(0)
	for i, object := range objects {
		card := l.cards[i]
		card.width, card.scale = size.Width, l.scale
		h := card.cardHeight()
		object.Move(fyne.NewPos(0, y))
		object.Resize(fyne.NewSize(size.Width, h))
		y += h + 8*l.scale
	}
}

type feedbackScroll struct {
	*container.Scroll
	stack   *feedbackStackLayout
	labels  []*ui.TextBlock
	buttons []*ui.TouchButton
}

func (s *feedbackScroll) SetScale(scale float32) {
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
	s.Content.Refresh()
}

func (s *feedbackScroll) Resize(size fyne.Size) {
	s.stack.width = size.Width
	s.Scroll.Resize(size)
	s.Content.Refresh()
}

func newFeedbackScroll(surface *ui.Surface, c *SkillSummary, data summaryData) *feedbackScroll {
	stack := &feedbackStackLayout{scale: 1, width: 500}
	result := &feedbackScroll{stack: stack}
	objects := make([]fyne.CanvasObject, 0, len(data.entries))
	for _, entry := range data.entries {
		entry := entry
		details := c.detail.Text(entry)
		layout := &feedbackCardLayout{title: entry.Skill.Name, details: details, width: 500, scale: 1}
		stack.cards = append(stack.cards, layout)
		bg := card.New(container.NewWithoutLayout(), card.Panel)
		title := ui.NewText(entry.Skill.Name, 23, ui.Navy)
		title.Bold = true
		progress := ui.NewText(navigation.ProgressText(entry.Progress), 18, ui.Green)
		body := ui.NewText(details, 18, ui.Navy)
		result.labels = append(result.labels, title, progress, body)
		button := surface.Button("feedback-attempts:"+entry.Skill.ID+":"+entry.Version, "Attempts", func() {
			c.onAttempts(data.person.ID, entry.Skill.ID, entry.Version)
		})
		button.FontSize = 18
		button.Fill = ui.Rail
		button.Ink = ui.Blue
		result.buttons = append(result.buttons, button)
		objects = append(objects, container.New(layout, bg, title, progress, body, button))
	}
	result.Scroll = container.NewVScroll(container.New(stack, objects...))
	return result
}
