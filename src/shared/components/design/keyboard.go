package design

import "fyne.io/fyne/v2"

// Escape handlers are canvas-local and used only while a library modal is open.
// Call these functions on the Fyne UI goroutine.
var escapeHandlers = map[fyne.Canvas]func(){}

func SetEscapeHandler(c fyne.Canvas, handler func()) {
	if handler == nil {
		delete(escapeHandlers, c)
	} else {
		escapeHandlers[c] = handler
	}
}

// HandleEscape lets library controls and custom modal content dismiss the top modal.
func HandleEscape(object fyne.CanvasObject, key *fyne.KeyEvent) bool {
	if key.Name != fyne.KeyEscape {
		return false
	}
	c := fyne.CurrentApp().Driver().CanvasForObject(object)
	if handler := escapeHandlers[c]; handler != nil {
		handler()
		return true
	}
	return false
}
