package group

import (
	"strings"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/cmd/beaker/options"
	"github.com/allenai/beaker/config"
)

type groupOptions struct {
	*options.AppOptions
	addr string
}

// NewGroupCmd creates the root command for this subpackage.
func NewGroupCmd(
	parent *kingpin.Application,
	parentOpts *options.AppOptions,
	config *config.Config,
) {
	o := &groupOptions{AppOptions: parentOpts}
	cmd := parent.Command("group", "Manage groups")

	cmd.Flag("addr", "Address of the Beaker service.").Default(config.BeakerAddress).StringVar(&o.addr)

	// Add automatic help generation for the command group.
	var helpSubcommands []string
	cmd.Command("help", "Show help.").Hidden().Default().PreAction(func(c *kingpin.ParseContext) error {
		fullCommand := append([]string{cmd.Model().Name}, helpSubcommands...)
		parent.Usage(fullCommand)
		return nil
	}).Arg("command", "Show help on command.").StringsVar(&helpSubcommands)

	// Attach subcommands.
	newAddCmd(cmd, o, config)
	newCreateCmd(cmd, o, config)
	newDeleteCmd(cmd, o, config)
	newInspectCmd(cmd, o, config)
	newRemoveCmd(cmd, o, config)
	newRenameCmd(cmd, o, config)
}

// Trim and unique a collection of strings, typically used to pre-process IDs.
func trimAndUnique(ids []string) []string {
	if len(ids) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	var unique []string
	for _, id := range ids {
		id := strings.TrimSpace(id)
		if _, ok := seen[id]; !ok {
			seen[id] = true
			unique = append(unique, id)
		}
	}

	return unique
}
