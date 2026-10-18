// Package actionbutton provides COB-styled keyboard and touch buttons.
package actionbutton

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"rectablet/src/shared/components/design"
)

type Variant string

const (
	Primary    Variant = "primary"
	Secondary  Variant = "secondary"
	Outline    Variant = "outline"
	Danger     Variant = "danger"
	Ghost      Variant = "ghost"
	Selectable Variant = "selectable"
)

type Size string

const (
	Small  Size = "sm"
	Medium Size = "md"
	Large  Size = "lg"
)

type Options struct {
	Variant  Variant
	Size     Size
	Disabled bool
}
type Button struct {
	widget.DisableableWidget
	text                       string
	content                    fyne.CanvasObject
	variant                    Variant
	size                       Size
	selected, hovered, focused bool
	radius                     float32
	onTap                      func()
}

func New(text string, options Options, onTap func()) *Button {
	b := &Button{text: text, variant: options.Variant, size: options.Size, onTap: onTap}
	b.ExtendBaseWidget(b)
	if options.Disabled {
		b.Disable()
	}
	return b
}

// NewContent creates a leading-aligned button containing arbitrary Fyne content.
func NewContent(content fyne.CanvasObject, options Options, onTap func()) *Button {
	b := New("", options, onTap)
	b.content = content
	return b
}
func (b *Button) SetText(text string)                  { b.text = text; b.Refresh() }
func (b *Button) SetContent(content fyne.CanvasObject) { b.content = content; b.Refresh() }
func (b *Button) SetSelected(selected bool)            { b.selected = selected; b.Refresh() }
func (b *Button) Selected() bool                       { return b.selected }
func (b *Button) SetRadius(radius float32)             { b.radius = radius; b.Refresh() }
func (b *Button) Tapped(*fyne.PointEvent) {
	if !b.Disabled() && b.onTap != nil {
		b.onTap()
	}
}
func (b *Button) FocusGained()   { b.focused = true; b.Refresh() }
func (b *Button) FocusLost()     { b.focused = false; b.Refresh() }
func (b *Button) TypedRune(rune) {}
func (b *Button) TypedKey(e *fyne.KeyEvent) {
	if design.HandleEscape(b, e) {
		return
	}
	if e.Name == fyne.KeyReturn || e.Name == fyne.KeySpace {
		b.Tapped(nil)
	}
}
func (b *Button) MouseIn(*desktop.MouseEvent)    { b.hovered = true; b.Refresh() }
func (b *Button) MouseMoved(*desktop.MouseEvent) {}
func (b *Button) MouseOut()                      { b.hovered = false; b.Refresh() }
func (b *Button) AccessibilityLabel() string     { return b.text }

var _ fyne.Focusable = (*Button)(nil)
