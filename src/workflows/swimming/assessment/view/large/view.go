package large

import (
	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/assessment/view/shared"
	"rectablet/src/workflows/swimming/navigation"
)

// Pages points to controller-owned positions that persist across renders.
type Pages struct{ Skill, Roster, Class, Group, Member, MultiRoster, Table, Heat *int }

// View renders large assessment layouts; feature data and actions arrive as callbacks.
type View struct {
	*navigation.Controller
	Pages             Pages
	Shared            *shared.View
	IsTimed           func(string) bool
	OpenTimedSkill    func(string) error
	OpenSkillSwimmer  func(navigation.Row, string) error
	SelectSkill       func(string) error
	SwitchPerson      func(string)
	ClassRoster       func() []ui.Item
	SelectClassPerson func(string)
	MultiRows         func(int, int) ([]ui.Item, []ui.Item)
	SelectGroup       func(string)
	LapsInput         func() string
	SelectionCount    func() int
	AddDigit          func(string)
	ClearLaps         func()
	StartHeat         func() error
	SelectedSwimmers  func() []ui.Item
}
