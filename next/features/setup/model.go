package setup

// Repository lists only the data operations this feature consumes.
type Repository interface {
	Export(string) error
}
