package common

import (
	"rectablet/next/navigation"
)

// Navigation supplies the operations a mode needs without embedding a controller.
type Navigation struct {
	Frame       func() navigation.Frame
	Rows        func() []navigation.Row
	Run         func(func() error) error
	Push        func(navigation.Route, func(*navigation.Frame))
	Pop         func()
	SetPage     func(int)
	UpdateFrame func(func(*navigation.Frame))
	Saved       func(string)
	Refresh     func() error
}
