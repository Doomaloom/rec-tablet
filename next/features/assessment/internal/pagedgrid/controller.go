package pagedgrid

// Controller binds panel-local page actions to the feature's refresh operation.
type Controller struct{ Run func(func() error) error }
