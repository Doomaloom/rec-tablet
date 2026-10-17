// Package emptystate provides empty content placeholders.
package emptystate

import "fyne.io/fyne/v2"

type Options struct {
	Title, Description string
	Content            fyne.CanvasObject
}

func New(o Options) fyne.CanvasObject { return build(o) }
