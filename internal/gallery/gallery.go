// Package gallery contains interactive examples with local demonstration state.
package gallery

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"rectablet/internal/shared/components/field"
	"rectablet/internal/shared/components/modalshell"
	"rectablet/internal/shared/components/notice"
	"rectablet/internal/shared/components/selectablecard"
	"rectablet/internal/shared/components/textinput"
	"strings"
)

type gallery struct {
	window           fyne.Window
	status           *notice.Notice
	name             *textinput.Input
	nameField        *field.Field
	choiceA, choiceB *selectablecard.Card
	modal            *modalshell.Modal
	tabText          *widget.Label
}

func New(window fyne.Window) fyne.CanvasObject { g := &gallery{window: window}; return g.build() }
func (g *gallery) report(text string)          { g.status.SetTone(notice.Info); g.status.SetText(text) }
func (g *gallery) save() {
	if strings.TrimSpace(g.name.Text) == "" {
		g.nameField.SetError("Enter a first name to continue.")
		g.status.SetTone(notice.Warning)
		g.status.SetText("Review the highlighted field.")
		return
	}
	g.nameField.SetError("")
	g.status.SetTone(notice.Success)
	g.status.SetText("Saved locally for this demo: " + g.name.Text)
}
func (g *gallery) choose(a bool) {
	g.choiceA.SetSelected(a)
	g.choiceB.SetSelected(!a)
	if a {
		g.report("Selected Morning lessons")
	} else {
		g.report("Selected Afternoon lessons")
	}
}
func (g *gallery) openModal() {
	if g.modal == nil {
		g.modal = modalshell.New(g.window, modalContent(), modalshell.Options{Title: "Edit lesson details", Description: "A reusable modal with an editable field and scrollable content.", Notice: notice.New("Changes here are demonstration data.", notice.Info), OnClose: func() { g.report("Modal closed") }})
	}
	g.modal.Show()
}
func modalContent() fyne.CanvasObject {
	input := textinput.New(textinput.Options{Value: "Morning lessons", Placeholder: "Lesson name"})
	return field.New(input.Object(), field.Options{Label: "Lesson name", Helper: "Close with ×, Escape, or the backdrop."})
}
