package config

import (
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/cmd/beaker/options"
	"github.com/allenai/beaker/config"
)

type configOptions struct {
	*options.AppOptions
	addr string
}

// NewConfigCmd creates the root command for this subpackage.
func NewConfigCmd(
	parent *kingpin.Application,
	parentOpts *options.AppOptions,
	config *config.Config,
) {
	o := &configOptions{AppOptions: parentOpts}
	cmd := parent.Command("config", "Manage Beaker configuration settings")

	cmd.Flag("addr", "Address of the Beaker service.").Default(config.BeakerAddress).StringVar(&o.addr)

	// Add automatic help generation for the command group.      
	var helpSubcommands []string
	cmd.Command("help", "Show help.").Hidden().Default().PreAction(func(c *kingpin.ParseContext) error {
		fullCommand := append([]string{cmd.Model().Name}, helpSubcommands...)
		parent.Usage(fullCommand)
		return nil
	}).Arg("command", "Show help on command.").StringsVar(&helpSubcommands)

	// Attach subcommands.
	newInteractiveCmd(cmd, o, config)
	newSetCmd(cmd, o, config)
	newTestCmd(cmd, o, config)
}
