// Package notice provides inline status messages.
package notice

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type Tone string

const (
	Info    Tone = "info"
	Success Tone = "success"
	Warning Tone = "warning"
	Danger  Tone = "danger"
)

type Notice struct {
	widget.BaseWidget
	text string
	tone Tone
	body fyne.CanvasObject
}

func New(text string, tone Tone) *Notice {
	n := &Notice{text: text, tone: tone}
	n.ExtendBaseWidget(n)
	return n
}
func (n *Notice) SetText(text string) {
	n.body = nil
	n.text = text
	n.Refresh()
}
func (n *Notice) SetTone(tone Tone) { n.tone = tone; n.Refresh() }

// NewContent renders a composed status message or inline action.
func NewContent(content fyne.CanvasObject, tone Tone) *Notice {
	n := New("", tone)
	n.body = content
	return n
}
