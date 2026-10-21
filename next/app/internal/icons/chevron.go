package icons

import (
	"fmt"

	"fyne.io/fyne/v2"
)

func Chevron(direction string, disabled bool) fyne.Resource {
	ink := "#007acc"
	if disabled {
		ink = "#669fbd"
	}
	path := "M12 24L32 44L52 24"
	switch direction {
	case "up":
		path = "M12 40L32 20L52 40"
	case "left":
		path = "M40 12L20 32L40 52"
	}
	return fyne.NewStaticResource(fmt.Sprintf("chevron-%s-%t.svg", direction, disabled), []byte(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 64 64"><path d="%s" fill="none" stroke="%s" stroke-width="7"/></svg>`, path, ink)))
}
