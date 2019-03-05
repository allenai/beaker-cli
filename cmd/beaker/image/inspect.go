package image

import (
	"context"
	"encoding/json"
	"os"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/api"
	beaker "github.com/allenai/beaker/client"
	"github.com/allenai/beaker/config"
)

type inspectOptions struct {
	images []string
}

func newInspectCmd(
	parent *kingpin.CmdClause,
	parentOpts *imageOptions,
	config *config.Config,
) {
	o := &inspectOptions{}
	cmd := parent.Command("inspect", "Display detailed information about one or more images")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Arg("image", "Image name or ID").Required().StringsVar(&o.images)
}

func (o *inspectOptions) run(beaker *beaker.Client) error {
	ctx := context.TODO()

	var images []*api.Image
	for _, name := range o.images {
		image, err := beaker.Image(ctx, name)
		if err != nil {
			return err
		}

		info, err := image.Get(ctx)
		if err != nil {
			return err
		}
		images = append(images, info)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "    ")
	return encoder.Encode(images)
}
