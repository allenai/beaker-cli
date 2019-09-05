package config

import (
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	config "github.com/allenai/beaker/config"
)

func newInteractiveCmd(
	parent *kingpin.CmdClause,
	parentOpts *configOptions,
	cfg *config.Config,
) {
	cmd := parent.Command("interactive", "Test the configuration")
	cmd.Action(func(c *kingpin.ParseContext) error {
		return config.InteractiveConfiguration()
	})
}
