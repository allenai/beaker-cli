package image

import (
	"context"
	"fmt"

	beaker "github.com/beaker/client/client"
	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type deleteOptions struct {
	image string
}

func newDeleteCmd(
	parent *kingpin.CmdClause,
	parentOpts *CmdOptions,
	config *config.Config,
) {
	o := &deleteOptions{}
	cmd := parent.Command("delete", "Permanently delete a image")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.Addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Arg("image", "Image name or ID").Required().StringVar(&o.image)
}

func (o *deleteOptions) run(beaker *beaker.Client) error {
	ctx := context.TODO()

	image, err := beaker.Image(ctx, o.image)
	if err != nil {
		return err
	}

	fmt.Printf("Deleted %s\n", color.BlueString(o.image))
	return image.Delete(ctx)
}
