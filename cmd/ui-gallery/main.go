// ui-gallery is a standalone UI playground; it never opens lesson storage.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"rectablet/src/gallery"
	"rectablet/src/shared/components/design"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	smoke := flag.Bool("smoke-test", false, "render briefly and exit")
	width := flag.Float64("width", 1000, "initial window width")
	height := flag.Float64("height", 800, "initial window height")
	flag.Parse()
	if flag.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", flag.Args())
	}
	if *width < 320 || *height < 320 {
		return fmt.Errorf("window dimensions must be at least 320")
	}
	a := app.NewWithID("org.rectablet.ui-gallery")
	a.Settings().SetTheme(design.Theme{})
	w := a.NewWindow("COB Aquatics · UI Gallery")
	w.SetContent(gallery.New(w))
	w.Resize(fyne.NewSize(float32(*width), float32(*height)))
	if *smoke {
		go func() { time.Sleep(800 * time.Millisecond); fyne.Do(w.Close) }()
	}
	w.ShowAndRun()
	return nil
}
