// Package shell hosts screens without knowing their workflow or storage model.
package shell

import (
	"image/color"
	"rectablet/internal/shared/ui"
)

type Row = ui.Item

// RailAction customizes one of the persistent left-rail buttons for a screen.
// Run is invoked after the button is pressed.
type RailAction struct {
	Label    string
	Run      func() error
	Fill     color.Color
	Ink      color.Color
	FontSize float32
}

// HeaderAction adds a feature-owned button beside the screen title.
type HeaderAction struct {
	Label    string
	Run      func() error
	Fill     color.Color
	Ink      color.Color
	FontSize float32
}
type State struct {
	// OnLeave releases feature resources when the route changes or the view closes.
	OnLeave                                                                               func()
	TitleDetails                                                                          bool
	Screen, Title, Subtitle, Context, Empty, Progress, Notice, HeaderAction, FooterAction string
	Error, CanBack, CanSubmit                                                             bool
	Page, PageCount, RowCount                                                             int
	Rows                                                                                  []Row
	PageSize                                                                              int
	HeaderWidth                                                                           float32
	DetailsTitle                                                                          string
	HoldTitle                                                                             func()
	HomeRailAction, BackRailAction                                                        *RailAction
	HeaderActions                                                                         []HeaderAction
	HeaderTitleFontSize                                                                   float32
	ComposeBody                                                                           ui.BodyComposer
}
type Presenter interface {
	Snapshot() State
	Subscribe(func())
	Home() error
	Back() error
	Page(int) error
	Header() error
	Submit() error
	Dismiss()
	OpenSetup() error
}

// History keeps navigation reusable; T belongs to the mounted workflow.
type History[T any] struct {
	Frames []T
	home   T
}

func NewHistory[T any](home T) *History[T] { return &History[T]{Frames: []T{home}, home: home} }
func (h *History[T]) Reset()               { h.Frames = []T{h.home} }
func (h *History[T]) Pop() {
	if len(h.Frames) > 1 {
		h.Frames = h.Frames[:len(h.Frames)-1]
	}
}
func (h *History[T]) Append(frame T) { h.Frames = append(h.Frames, frame) }
