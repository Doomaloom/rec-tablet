// Package modalshell provides a dismissible, canvas-local modal overlay.
package modalshell

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"rectablet/src/shared/components/actionbutton"
	"rectablet/src/shared/components/design"
)

type Options struct {
	Title, Description string
	Notice             fyne.CanvasObject
	// Footer stays visible while long modal content scrolls.
	Footer   fyne.CanvasObject
	MaxWidth float32
	OnClose  func()
}
type Modal struct {
	widget.BaseWidget
	window        fyne.Window
	content       fyne.CanvasObject
	options       Options
	panel         fyne.CanvasObject
	body          fyne.CanvasObject
	close         *actionbutton.Button
	shown         bool
	previousFocus fyne.Focusable
	panelPos      fyne.Position
	panelSize     fyne.Size
}

func New(window fyne.Window, content fyne.CanvasObject, o Options) *Modal {
	m := &Modal{window: window, content: content, options: o}
	if m.options.MaxWidth <= 0 {
		m.options.MaxWidth = 896
	}
	m.ExtendBaseWidget(m)
	m.build()
	return m
}

var modalStacks = map[fyne.Canvas][]*Modal{}

var previousKeys = map[fyne.Canvas]func(*fyne.KeyEvent){}

func (m *Modal) Show() {
	if m.shown {
		return
	}
	c := m.window.Canvas()
	m.previousFocus = c.Focused()
	m.shown = true
	m.BaseWidget.Show()
	if len(modalStacks[c]) == 0 {
		previousKeys[c] = c.OnTypedKey()
		c.SetOnTypedKey(func(key *fyne.KeyEvent) {
			if key.Name == fyne.KeyEscape {
				stack := modalStacks[c]
				if len(stack) > 0 {
					stack[len(stack)-1].Hide()
				}
			} else if old := previousKeys[c]; old != nil {
				old(key)
			}
		})
	}
	modalStacks[c] = append(modalStacks[c], m)
	design.SetEscapeHandler(c, func() {
		stack := modalStacks[c]
		if len(stack) > 0 {
			stack[len(stack)-1].Hide()
		}
	})
	c.Overlays().Add(m)
	c.Focus(m.close)
}
func (m *Modal) Hide() {
	if !m.shown {
		return
	}
	c := m.window.Canvas()
	stack := modalStacks[c]
	// Fyne removes overlays above a removed overlay; close child modals first.
	for i, item := range stack {
		if item == m {
			for len(modalStacks[c]) > i+1 {
				children := modalStacks[c]
				children[len(children)-1].Hide()
			}
			break
		}
	}
	stack = modalStacks[c]
	top := len(stack) > 0 && stack[len(stack)-1] == m
	for i, item := range stack {
		if item == m {
			stack = append(stack[:i], stack[i+1:]...)
			break
		}
	}
	if len(stack) == 0 {
		delete(modalStacks, c)
		design.SetEscapeHandler(c, nil)
		c.SetOnTypedKey(previousKeys[c])
		delete(previousKeys, c)
	} else {
		modalStacks[c] = stack
	}
	c.Overlays().Remove(m)
	m.shown = false
	m.BaseWidget.Hide()
	if top {
		if len(stack) > 0 {
			c.Focus(stack[len(stack)-1].close)
		} else if m.previousFocus != nil {
			c.Focus(m.previousFocus)
		} else {
			c.Unfocus()
		}
	}
	if m.options.OnClose != nil {
		m.options.OnClose()
	}
}
func (m *Modal) Visible() bool { return m.shown }
func (m *Modal) Tapped(e *fyne.PointEvent) {
	p := e.Position
	if p.X < m.panelPos.X || p.Y < m.panelPos.Y || p.X > m.panelPos.X+m.panelSize.Width || p.Y > m.panelPos.Y+m.panelSize.Height {
		m.Hide()
	}
}
