package image

import (
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/cmd/beaker/options"
	"github.com/allenai/beaker/config"
)

type imageOptions struct {
	*options.AppOptions
	addr string
}

// NewImageCmd creates the root command for this subpackage.
func NewImageCmd(
	parent *kingpin.Application,
	parentOpts *options.AppOptions,
	config *config.Config,
) {
	o := &imageOptions{AppOptions: parentOpts}
	cmd := parent.Command("image", "Manage images")

	cmd.Flag("addr", "Address of the Beaker service.").Default(config.BeakerAddress).StringVar(&o.addr)

	// Add automatic help generation for the command group.
	var helpSubcommands []string
	cmd.Command("help", "Show help.").Hidden().Default().PreAction(func(c *kingpin.ParseContext) error {
		fullCommand := append([]string{cmd.Model().Name}, helpSubcommands...)
		parent.Usage(fullCommand)
		return nil
	}).Arg("command", "Show help on command.").StringsVar(&helpSubcommands)

	// Attach subcommands.
	newCreateCmd(cmd, o, config)
	newInspectCmd(cmd, o, config)
	newRenameCmd(cmd, o, config)
	newPullCmd(cmd, o, config)
}
