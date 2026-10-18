// Package segmentedtabs provides a wrapping group of exclusive choices.
package segmentedtabs

import (
	"fyne.io/fyne/v2/widget"
	"rectablet/src/shared/components/actionbutton"
)

type Item struct {
	Key, Label string
	Disabled   bool
}
type Shape string

const (
	Rounded Shape = "rounded"
	Pill    Shape = "pill"
)

type Options struct {
	ActiveKey string
	Shape     Shape
	Size      actionbutton.Size
	OnChanged func(string)
}
type Tabs struct {
	widget.BaseWidget
	items   []Item
	options Options
	buttons []*actionbutton.Button
}

func New(items []Item, o Options) *Tabs {
	t := &Tabs{items: append([]Item(nil), items...), options: o}
	t.ExtendBaseWidget(t)
	t.build()
	return t
}
func (t *Tabs) ActiveKey() string { return t.options.ActiveKey }

// SetActive updates a valid, enabled choice without invoking OnChanged.
func (t *Tabs) SetActive(key string) {
	for _, item := range t.items {
		if item.Key == key && !item.Disabled {
			t.options.ActiveKey = key
			t.Refresh()
			return
		}
	}
}
func (t *Tabs) choose(key string) {
	if key == t.options.ActiveKey {
		return
	}
	t.SetActive(key)
	if t.options.ActiveKey == key && t.options.OnChanged != nil {
		t.options.OnChanged(key)
	}
}
