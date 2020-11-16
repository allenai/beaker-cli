package secret

import (
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/cmd/beaker/options"
	"github.com/allenai/beaker/config"
)

type secretOptions struct {
	*options.AppOptions
	addr      string
	workspace string
}

// NewSecretCmd creates the root command for this subpackage.
func NewSecretCmd(
	parent *kingpin.Application,
	parentOpts *options.AppOptions,
	config *config.Config,
) {
	o := &secretOptions{AppOptions: parentOpts}
	cmd := parent.Command("secret", "Manage secrets")

	cmd.Flag("addr", "Address of the Beaker service.").Default(config.BeakerAddress).StringVar(&o.addr)
	cmd.Flag("workspace", "Workspace containing the secret").Required().StringVar(&o.workspace)

	// Add automatic help generation for the command group.
	var helpSubcommands []string
	cmd.Command("help", "Show help.").Hidden().Default().PreAction(func(c *kingpin.ParseContext) error {
		fullCommand := append([]string{cmd.Model().Name}, helpSubcommands...)
		parent.Usage(fullCommand)
		return nil
	}).Arg("command", "Show help on command.").StringsVar(&helpSubcommands)

	newWriteCmd(cmd, o, config)
	newReadCmd(cmd, o, config)
	newInspectCmd(cmd, o, config)
	newListCmd(cmd, o, config)
	newDeleteCmd(cmd, o, config)
}
