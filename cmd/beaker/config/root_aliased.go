package config

import (
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/cmd/beaker/options"
	beakerConfig "github.com/allenai/beaker/config"
)

// NewConfigureCmd replicates the old "configure" for back-compat
func NewConfigureCmd(
	parent *kingpin.Application,
	parentOpts *options.AppOptions,
	config *beakerConfig.Config,
) {
	o := &configOptions{AppOptions: parentOpts}
	cmd := parent.Command("configure", "Manage Beaker configuration settings")
	cmd.Command("interactive", "Interactive configuration").Default().Action(
		func(c *kingpin.ParseContext) error {
			return beakerConfig.InteractiveConfiguration()
		})

	// Attach subcommands.
	newTestCmd(cmd, o, config)
}
