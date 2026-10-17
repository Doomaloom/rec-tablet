// Package field groups a label, control, helper, and validation error.
package field

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type Options struct{ Label, Helper, Error string }
type Field struct {
	widget.BaseWidget
	control fyne.CanvasObject
	options Options
}

func New(control fyne.CanvasObject, o Options) *Field {
	f := &Field{control: control, options: o}
	f.ExtendBaseWidget(f)
	return f
}
func (f *Field) SetError(message string)  { f.options.Error = message; f.Refresh() }
func (f *Field) SetHelper(message string) { f.options.Helper = message; f.Refresh() }
