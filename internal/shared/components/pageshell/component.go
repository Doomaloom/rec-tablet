// Package pageshell provides a centered page with bounded content width.
package pageshell

import "fyne.io/fyne/v2"

type Options struct{ MaxWidth float32 }

func New(o Options, content ...fyne.CanvasObject) fyne.CanvasObject {
	if o.MaxWidth <= 0 {
		o.MaxWidth = 1152
	}
	return build(o, content)
}
