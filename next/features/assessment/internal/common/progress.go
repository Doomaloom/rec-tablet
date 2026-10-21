package common

// ProgressActions coordinates the shared correction gesture used by three modes.
type ProgressActions struct {
	Navigation Navigation
	Remove     func(string, string, string) error
}

func (p *ProgressActions) RemoveLatestSuccess(person, skill, version string) error {
	return p.Navigation.Run(func() error {
		if err := p.Remove(person, skill, version); err != nil {
			return nil
		}
		p.Navigation.Saved("")
		return nil
	})
}
