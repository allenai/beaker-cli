package cluster

import (
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/cmd/beaker/options"
	"github.com/allenai/beaker/config"
)

type clusterOptions struct {
	*options.AppOptions
	addr string
}

// NewClusterCmd creates the root command for this subpackage.
func NewClusterCmd(
	parent *kingpin.Application,
	parentOpts *options.AppOptions,
	config *config.Config,
) {
	o := &clusterOptions{AppOptions: parentOpts}
	// TODO: Remove the "under development" when it no longer applies.
	cmd := parent.Command("cluster", "Manage clusters (under development)")

	cmd.Flag("addr", "Address of the Beaker service.").Default(config.BeakerAddress).StringVar(&o.addr)

	// Add automatic help generation for the command group.
	var helpSubcommands []string
	cmd.Command("help", "Show help.").Hidden().Default().PreAction(func(c *kingpin.ParseContext) error {
		fullCommand := append([]string{cmd.Model().Name}, helpSubcommands...)
		parent.Usage(fullCommand)
		return nil
	}).Arg("command", "Show help on command.").StringsVar(&helpSubcommands)

	// Attach subcommands.
	// TODO: Define a list command.
	newCreateCmd(cmd, o, config)
	newExtendCmd(cmd, o, config) // TODO: Should this be automatic on experiment create?
	newInspectCmd(cmd, o, config)
	newTerminateCmd(cmd, o, config)
}
