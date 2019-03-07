package blueprint

import (
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	beaker "github.com/allenai/beaker/client"
	"github.com/allenai/beaker/cmd/beaker/image"
	"github.com/allenai/beaker/config"
)

func newInspectCmd(
	parent *kingpin.CmdClause,
	parentOpts *image.CmdOptions,
	config *config.Config,
) {
	o := &image.InspectOptions{}
	cmd := parent.Command("inspect", "Display detailed information about one or more blueprints")
	cmd.Action(func(c *kingpin.ParseContext) error {
		printDeprecationWarning()
		beaker, err := beaker.NewClient(parentOpts.Addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.Run(beaker)
	})

	cmd.Arg("blueprint", "Blueprint name or ID").Required().StringsVar(&o.Images)
}
