package common

type MultiStudentSnapshot struct {
	PersonID, Name, Status string
	Passed, Total          int
}

// Actions contains only the operations that rendering controls can emit.
type Actions struct {
	Select func(string) error
	Mark   func(string) error
	Submit func() error
	Run    func(func() error) error
}
