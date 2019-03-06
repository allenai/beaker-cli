package blueprint

import (
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/cmd/beaker/image"
	"github.com/allenai/beaker/config"
)

func newRenameCmd(
	parent *kingpin.CmdClause,
	parentOpts *image.ImageOptions,
	config *config.Config,
) {
	o := &image.RenameOptions{}
	cmd := parent.Command("rename", "Rename an blueprint")
	cmd.Action(func(c *kingpin.ParseContext) error {
		// TODO message reminding to switch to image commands
		return o.Run(parentOpts, config.UserToken)
	})

	cmd.Flag("quiet", "Only display the blueprint's unique ID").Short('q').BoolVar(&o.Quiet)
	cmd.Arg("blueprint", "Name or ID of the blueprint to rename").Required().StringVar(&o.Image)
	cmd.Arg("new-name", "Unqualified name to assign to the blueprint").Required().StringVar(&o.Name)
}
