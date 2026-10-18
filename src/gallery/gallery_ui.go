package gallery

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"rectablet/src/shared/components/actionbutton"
	"rectablet/src/shared/components/card"
	"rectablet/src/shared/components/design"
	"rectablet/src/shared/components/emptystate"
	"rectablet/src/shared/components/field"
	"rectablet/src/shared/components/notice"
	"rectablet/src/shared/components/pageshell"
	"rectablet/src/shared/components/segmentedtabs"
	selectcontrol "rectablet/src/shared/components/select"
	"rectablet/src/shared/components/selectablecard"
	"rectablet/src/shared/components/textarea"
	"rectablet/src/shared/components/textinput"
)

func section(title, description string, objects ...fyne.CanvasObject) fyne.CanvasObject {
	items := []fyne.CanvasObject{design.Label(title, 22, design.Secondary, true), design.Label(description, 14, design.Text, false)}
	items = append(items, objects...)
	return card.New(design.Column(16, items...), card.Panel)
}
func flow(objects ...fyne.CanvasObject) fyne.CanvasObject {
	return container.New(layout.NewGridWrapLayout(fyne.NewSize(200, 56)), objects...)
}
func (g *gallery) build() fyne.CanvasObject {
	g.status = notice.New("Try a control. All changes stay in this gallery.", notice.Info)
	header := design.Column(4, design.Label("COB AQUATICS / REC TABLET", 12, design.Header, true), design.Label("A familiar foundation.", 32, design.Secondary, true), design.Label("Rounded surfaces, clear blue actions, and Rubik typography — translated into native Fyne components.", 16, design.Text, false))
	palette := []fyne.CanvasObject{}
	for _, token := range []struct {
		name, hex string
		ink       color.Color
	}{{"Primary", "#007acc", design.Primary}, {"Secondary", "#005f99", design.Secondary}, {"Background", "#e6f7ff", design.Background}, {"Header", "#c45300", design.Header}, {"Hover", "#66d9ef", design.Hover}, {"Danger", "#d64545", design.Danger}} {
		swatch := canvas.NewRectangle(token.ink)
		swatch.CornerRadius = 12
		swatch.SetMinSize(fyne.NewSize(100, 32))
		palette = append(palette, design.Column(0, swatch, design.Label(token.name+" "+token.hex, 12, design.Secondary, true)))
	}
	buttons := []fyne.CanvasObject{}
	for _, v := range []actionbutton.Variant{actionbutton.Primary, actionbutton.Secondary, actionbutton.Outline, actionbutton.Danger, actionbutton.Ghost} {
		v := v
		buttons = append(buttons, actionbutton.New(string(v), actionbutton.Options{Variant: v}, func() { g.report("Activated " + string(v)) }))
	}
	buttons = append(buttons, actionbutton.New("Disabled", actionbutton.Options{Disabled: true}, nil))
	sizes := []fyne.CanvasObject{}
	for _, s := range []actionbutton.Size{actionbutton.Small, actionbutton.Medium, actionbutton.Large} {
		s := s
		sizes = append(sizes, actionbutton.New("Size "+string(s), actionbutton.Options{Variant: actionbutton.Primary, Size: s}, func() { g.report("Size " + string(s)) }))
	}
	cards := []fyne.CanvasObject{}
	for _, v := range []card.Variant{card.Panel, card.Subtle, card.Interactive, card.Plain} {
		cards = append(cards, card.New(design.Label(string(v)+" surface", 14, design.Secondary, true), v))
	}
	g.choiceA = selectablecard.New(widget.NewLabel("Morning lessons"), selectablecard.Options{Selected: true}, func() { g.choose(true) })
	g.choiceB = selectablecard.New(widget.NewLabel("Afternoon lessons"), selectablecard.Options{}, func() { g.choose(false) })
	disabledCard := selectablecard.New(widget.NewLabel("Unavailable"), selectablecard.Options{Disabled: true}, nil)
	g.tabText = widget.NewLabel("Showing attendance")
	tabs := segmentedtabs.New([]segmentedtabs.Item{{Key: "attendance", Label: "Attendance"}, {Key: "skills", Label: "Skills"}, {Key: "feedback", Label: "Feedback"}, {Key: "locked", Label: "Unavailable", Disabled: true}}, segmentedtabs.Options{ActiveKey: "attendance", OnChanged: func(key string) { g.tabText.SetText("Showing " + key) }})
	pills := segmentedtabs.New([]segmentedtabs.Item{{Key: "one", Label: "Single"}, {Key: "multi", Label: "Multi"}, {Key: "class", Label: "Class"}}, segmentedtabs.Options{ActiveKey: "one", Shape: segmentedtabs.Pill, Size: actionbutton.Small, OnChanged: func(key string) { g.report("Evaluation mode: " + key) }})
	g.name = textinput.New(textinput.Options{Placeholder: "First name"})
	g.nameField = field.New(g.name.Object(), field.Options{Label: "First name", Helper: "Required for the demo save action."})
	notes := textarea.New(textarea.Options{Placeholder: "Write a lesson note…"})
	level := selectcontrol.New(selectcontrol.Options{Options: []string{"Swimmer 1", "Swimmer 2", "Swimmer 3"}, Placeholder: "Choose a level", OnChanged: func(value string) { g.report("Level: " + value) }})
	disabledInput := textinput.New(textinput.Options{Value: "Read only while disabled", Disabled: true})
	save := actionbutton.New("Save example", actionbutton.Options{Variant: actionbutton.Primary}, g.save)
	form := design.Column(12, g.nameField, field.New(level.Object(), field.Options{Label: "Level"}), field.New(notes.Object(), field.Options{Label: "Notes", Helper: "Native multiline editing, wrapping, and selection."}), field.New(disabledInput.Object(), field.Options{Label: "Disabled input"}), save)
	notices := []fyne.CanvasObject{}
	for _, tone := range []notice.Tone{notice.Info, notice.Success, notice.Warning, notice.Danger} {
		notices = append(notices, notice.New(fmt.Sprintf("%s — a concise message with a clear next step.", tone), tone))
	}
	page := pageshell.New(pageshell.Options{MaxWidth: 1000}, header, g.status,
		section("Design tokens", "The shared web palette and embedded Rubik regular / semibold fonts.", container.New(layout.NewGridWrapLayout(fyne.NewSize(200, 88)), palette...)),
		section("Action buttons", "Five variants, three sizes, keyboard focus, and disabled feedback.", flow(buttons...), flow(sizes...)),
		section("Cards", "Panel, subtle, interactive hover, and plain surfaces.", design.Column(12, cards...)),
		section("Selectable cards", "Selection is explicit; callers own whether choices are exclusive.", design.Column(8, g.choiceA, g.choiceB, disabledCard)),
		section("Segmented tabs", "Rounded and pill shapes, with local selection state.", tabs, g.tabText, pills),
		section("Fields and inputs", "Try saving with an empty name, then enter a name and save again.", form),
		section("Notices", "Info, success, warning, and danger follow the web app’s tones.", design.Column(8, notices...)),
		section("Empty state", "A quiet placeholder with an optional title and body.", emptystate.New(emptystate.Options{Title: "No lessons yet", Description: "Choose a session to see its lessons here."})),
		section("Modal shell", "A separate overlay with backdrop dismissal and keyboard support.", actionbutton.New("Open modal", actionbutton.Options{Variant: actionbutton.Secondary}, g.openModal)),
		design.Label("PageShell centers this gallery and limits its content width. Resize the window to explore the layout.", 12, design.Secondary, false))
	return design.Scope(container.NewVScroll(page))
}
