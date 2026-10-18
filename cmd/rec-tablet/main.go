package main

import (
	"errors"
	"flag"
	"fmt"
	application "rectablet/src/app"

	"os"
	"path/filepath"
	"rectablet/src/platform/files"

	"rectablet/src/workflows/swimming/curriculum"

	"rectablet/src/workflows/swimming/transfer"
)

func main() {
	if err := run(os.Args[1:]); err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Fprintln(os.Stderr, "Rec Tablet:", err)
		os.Exit(1)
	}
}
func run(args []string) error {
	if len(args) > 0 {
		switch args[0] {
		case "adapt-rosters":
			return adapt(args[1:])
		case "extract-curriculum":
			return extract(args[1:])
		}
	}
	home, e := os.UserHomeDir()
	if e != nil {
		return e
	}
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = filepath.Join(home, ".local/share")
	}
	flags := flag.NewFlagSet("rec-tablet", flag.ContinueOnError)
	database := flags.String("database", filepath.Join(dataHome, "rec-tablet/lessons.sqlite3"), "SQLite database path")
	shareServerDefault := os.Getenv("REC_TABLET_SHARE_SERVER")
	if shareServerDefault == "" {
		shareServerDefault = transfer.DefaultShareServer
	}
	shareServer := flags.String("share-server", shareServerDefault, "temporary data share server URL")
	fullscreen := flags.Bool("fullscreen", false, "start fullscreen")
	smoke := flags.Bool("smoke-test", false, "render briefly and exit")
	screenshot := flags.String("screenshot", "", "PNG output during --smoke-test")
	if e = flags.Parse(args); e != nil {
		return e
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	if *screenshot != "" && !*smoke {
		return fmt.Errorf("--screenshot requires --smoke-test")
	}
	return application.Run(application.Options{Database: *database, ShareServer: *shareServer, Fullscreen: *fullscreen, SmokeTest: *smoke, Screenshot: *screenshot})
}

func adapt(args []string) error {
	f := flag.NewFlagSet("adapt-rosters", flag.ContinueOnError)
	dataset := f.String("dataset-id", "", "stable dataset identifier")
	instructor := f.String("instructor", "", "instructor name")
	if e := f.Parse(args); e != nil {
		return e
	}
	if f.NArg() != 2 || *dataset == "" || *instructor == "" {
		return fmt.Errorf("usage: rec-tablet adapt-rosters --dataset-id ID --instructor NAME INPUT.json OUTPUT.json")
	}
	data, e := os.ReadFile(f.Arg(0))
	if e != nil {
		return e
	}
	p, e := transfer.Adapt(data, *dataset, *instructor)
	if e != nil {
		return e
	}
	return files.WriteJSON(f.Arg(1), p)
}
func extract(args []string) error {
	f := flag.NewFlagSet("extract-curriculum", flag.ContinueOnError)
	output := f.String("output", "curriculum.json", "output catalog file")
	if e := f.Parse(args); e != nil {
		return e
	}
	if f.NArg() != 1 {
		return fmt.Errorf("usage: rec-tablet extract-curriculum --output FILE TEMPLATE_DIRECTORY")
	}
	c, e := curriculum.ExtractDirectory(f.Arg(0))
	if e != nil {
		return e
	}
	if e = files.WriteJSON(*output, c); e != nil {
		return e
	}
	fmt.Printf("Extracted %d templates · %s\n", len(c.Levels), c.Version)
	return nil
}
