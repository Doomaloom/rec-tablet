package assessment

import (
	"rectablet/src/workflows/swimming/navigation"
)

func (c *Component) toggleSubskill(id string) error {
	f := c.navigation.Frame()
	status := "successful"
	for _, row := range c.navigation.State.Rows {
		if row.ID == id && row.Status == "successful" {
			status = "unmarked"
			break
		}
	}
	if e := c.store.Mark(f.Attempt, id, status); e != nil {
		// Keep the checklist usable if a local write fails; the persisted draft
		// remains unchanged and the next refresh shows its previous state.
		return nil
	}
	c.navigation.Saved("")
	return nil
}

// SaveAttempt stores an incomplete attempt without counting it toward progress.
func (c *Component) SaveAttempt() error {
	if c.navigation.Frame().Screen != navigation.Assessment || c.navigation.Frame().Review {
		return nil
	}
	return c.navigation.Run(func() error {
		if _, e := c.store.Submit(c.navigation.Frame().Attempt); e != nil {
			return nil
		}
		c.navigation.Pop()
		return nil
	})
}

// DiscardAttempt removes the unfinished draft and returns to the student list.
func (c *Component) DiscardAttempt() error {
	if c.navigation.Frame().Screen != navigation.Assessment || c.navigation.Frame().Review {
		return nil
	}
	return c.navigation.Run(func() error {
		if e := c.store.DiscardAttempt(c.navigation.Frame().Attempt); e != nil {
			return nil
		}
		c.navigation.Pop()
		return nil
	})
}
