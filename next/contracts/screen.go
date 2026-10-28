// Package contracts defines the boundary between features and the persistent UI.
package contracts

import (
	"image/color"

	"fyne.io/fyne/v2"

	"rectablet/src/shared/ui"
)

// ButtonUpdate changes an existing control without rebuilding its screen.
type ButtonUpdate struct {
	ID, Text  string
	Fill, Ink color.Color
	Selected  bool
}

type Row = ui.Item

// RailAction customizes one of the persistent rail buttons for a screen.
// Run is invoked after the button is pressed.
type RailAction struct {
	ID       string
	Label    string
	Run      func() error
	Fill     color.Color
	Ink      color.Color
	FontSize float32
	Icon     fyne.Resource
}

// HeaderAction adds a feature-owned button beside the screen title.
type HeaderAction struct {
	Label    string
	Run      func() error
	Fill     color.Color
	Ink      color.Color
	FontSize float32
}

// HeaderSwitchOption is one choice in a compact header mode switch.
type HeaderSwitchOption struct {
	Label    string
	Selected bool
	Run      func() error
}

// SplitHeader replaces the normal title with two prominent, wrapping labels.
type SplitHeader struct {
	Left, Right string
}

type State struct {
	// BindTitle installs an in-place title updater; invoke updates on the UI thread.
	BindTitle   func(func(string))
	DisplayMode ui.DisplayMode
	// PrimaryInfoFontSize defaults to 20; dense screens can reserve more body space.
	PrimaryInfoFontSize float32
	// OnLeave releases feature resources when the route changes or the view closes.
	OnLeave                                                            func()
	PrimaryInfo                                                        string
	ProgressInBody                                                     bool
	Screen, Title, Empty, Progress, Notice, HeaderAction, FooterAction string
	Error, CanBack, CanSubmit                                          bool
	Page, PageCount, RowCount                                          int
	Rows                                                               []Row
	PageSize                                                           int
	AlwaysShowPagePosition                                             bool
	HeaderWidth                                                        float32
	HoldTitle                                                          func()
	HomeRailAction, BackRailAction                                     *RailAction
	// A nil rail uses its defaults. An empty, non-nil rail hides its controls.
	// Actions replace the selected rail in top-to-bottom order for this screen.
	LeftRailActions     []RailAction
	RightRailActions    []RailAction
	HeaderActions       []HeaderAction
	HeaderSwitch        []HeaderSwitchOption
	HeaderTitleFontSize float32
	// HeaderStartFraction aligns the header after a leading body column.
	HeaderStartFraction float32
	HideLeftRail        bool
	SplitHeader         *SplitHeader
	FitHeader           bool
	HideHeader          bool
	HideFooter          bool
	ComposeBody         ui.BodyComposer
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

const SplitHeaderHeight float32 = 52
