package image

import (
	"context"
	"encoding/json"
	"os"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/beaker/client/api"
	beaker "github.com/beaker/client/client"
	"github.com/allenai/beaker/config"
)

// InspectOptions holds the images to inspect
// TODO: make unexported once not needed by blueprint command
type InspectOptions struct {
	Images []string
}

func newInspectCmd(
	parent *kingpin.CmdClause,
	parentOpts *CmdOptions,
	config *config.Config,
) {
	o := &InspectOptions{}
	cmd := parent.Command("inspect", "Display detailed information about one or more images")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.Addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.Run(beaker)
	})

	cmd.Arg("image", "Image name or ID").Required().StringsVar(&o.Images)
}

// Run runs the inspect command on the InspectOptions images
// TODO: make unexported once not needed by blueprint command
func (o *InspectOptions) Run(beaker *beaker.Client) error {
	ctx := context.TODO()

	var images []*api.Image
	for _, name := range o.Images {
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
