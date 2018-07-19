package main

import (
	"fmt"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type versionOptions struct{}

// These variables are set externally by the linker.
var (
	version = "dev"
	commit  = "unknown"
)

// NewVersionCmd creates a new Kingpin 'version' subcommand.
func NewVersionCmd(parent *kingpin.Application) {
	o := &versionOptions{}
	parent.Command("version", "Print the Beaker version").Action(o.run)
}

func (o *versionOptions) run(_ *kingpin.ParseContext) error {
	fmt.Println(makeVersion())
	return nil
}

func makeVersion() string {
	return fmt.Sprintf("Beaker %s ('%s')", version, commit)
}
