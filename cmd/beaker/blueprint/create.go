package blueprint

import (
	"context"
	"os"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	beaker "github.com/beaker/client/client"
	"github.com/allenai/beaker/cmd/beaker/image"
	"github.com/allenai/beaker/config"
)

func newCreateCmd(
	parent *kingpin.CmdClause,
	parentOpts *image.CmdOptions,
	config *config.Config,
) {
	opts := &image.CreateOptions{}
	imageRef := new(string)

	cmd := parent.Command("create", "Create a new blueprint")
	cmd.Flag("desc", "Assign a description to the blueprint").StringVar(&opts.Description)
	cmd.Flag("name", "Assign a name to the blueprint").Short('n').StringVar(&opts.Name)
	cmd.Flag("quiet", "Only display created blueprint's ID").Short('q').BoolVar(&opts.Quiet)
	cmd.Arg("image", "Docker image ID").Required().StringVar(imageRef)

	cmd.Action(func(c *kingpin.ParseContext) error {
		printDeprecationWarning()
		beaker, err := beaker.NewClient(parentOpts.Addr, config.UserToken)
		if err != nil {
			return err
		}
		_, err = image.Create(context.TODO(), os.Stdout, beaker, *imageRef, opts)
		return err
	})
}
