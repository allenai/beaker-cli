package workspace

import (
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/cmd/beaker/options"
	"github.com/allenai/beaker/config"
)

type workspaceOptions struct {
	*options.AppOptions
	addr string
}

// NewWorkspaceCmd creates the root command for this subpackage.
func NewWorkspaceCmd(
	parent *kingpin.Application,
	parentOpts *options.AppOptions,
	config *config.Config,
) {
	o := &workspaceOptions{AppOptions: parentOpts}
	// TODO: Remove the "under development" when it no longer applies.
	cmd := parent.Command("workspace", "Manage workspaces (under development)")

	cmd.Flag("addr", "Address of the Beaker service.").Default(config.BeakerAddress).StringVar(&o.addr)

	// Add automatic help generation for the command group.
	var helpSubcommands []string
	cmd.Command("help", "Show help.").Hidden().Default().PreAction(func(c *kingpin.ParseContext) error {
		fullCommand := append([]string{cmd.Model().Name}, helpSubcommands...)
		parent.Usage(fullCommand)
		return nil
	}).Arg("command", "Show help on command.").StringsVar(&helpSubcommands)

	// Attach subcommands.
	newArchiveCmd(cmd, o, config)
	newCreateCmd(cmd, o, config)
	newInspectCmd(cmd, o, config)
	newMoveCmd(cmd, o, config)
	newRenameCmd(cmd, o, config)
	newUnarchiveCmd(cmd, o, config)
}
