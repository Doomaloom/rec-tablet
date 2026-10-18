package setup

type ProfileOption struct{ Key, Label string }

// Repository lists only the data operations this feature consumes.
type Repository interface {
	Meta(string) (string, error)
	DatabasePath() string
	ImportFile(string) (int, error)
	Dates() ([]string, error)
	Export(string) error
}
