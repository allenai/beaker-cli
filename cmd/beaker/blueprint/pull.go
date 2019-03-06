package blueprint

import (
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	beaker "github.com/allenai/beaker/client"
	"github.com/allenai/beaker/cmd/beaker/image"
	"github.com/allenai/beaker/config"
)

func newPullCmd(
	parent *kingpin.CmdClause,
	parentOpts *image.ImageOptions,
	config *config.Config,
) {
	o := &image.PullOptions{}
	cmd := parent.Command("pull", "Pull the blueprint's Docker image")
	cmd.Flag("quiet", "Only display the pulled image's tag").Short('q').BoolVar(&o.Quiet)
	cmd.Arg("blueprint", "Blueprint name or ID").Required().StringVar(&o.Image)
	cmd.Arg("tag", "Name and optional tag in the 'name:tag' format").StringVar(&o.Tag)

	cmd.Action(func(c *kingpin.ParseContext) error {
		PrintBeakerDeprecationWarning()
		beaker, err := beaker.NewClient(parentOpts.Addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.Run(beaker)
	})
}
