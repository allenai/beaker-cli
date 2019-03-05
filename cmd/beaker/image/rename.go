package image

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	beaker "github.com/allenai/beaker/client"
	"github.com/allenai/beaker/config"
)

type renameOptions struct {
	quiet bool
	image string
	name  string
}

func newRenameCmd(
	parent *kingpin.CmdClause,
	parentOpts *imageOptions,
	config *config.Config,
) {
	o := &renameOptions{}
	cmd := parent.Command("rename", "Rename an image")
	cmd.Action(func(c *kingpin.ParseContext) error { return o.run(parentOpts, config.UserToken) })

	cmd.Flag("quiet", "Only display the image's unique ID").Short('q').BoolVar(&o.quiet)
	cmd.Arg("image", "Name or ID of the image to rename").Required().StringVar(&o.image)
	cmd.Arg("new-name", "Unqualified name to assign to the image").Required().StringVar(&o.name)
}

func (o *renameOptions) run(parentOpts *imageOptions, userToken string) error {
	ctx := context.TODO()
	beaker, err := beaker.NewClient(parentOpts.addr, userToken)
	if err != nil {
		return err
	}
	image, err := beaker.Image(ctx, o.image)
	if err != nil {
		return err
	}

	if err := image.SetName(ctx, o.name); err != nil {
		return err
	}

	// TODO: This info should probably be part of the client response instead of a separate get.
	info, err := image.Get(ctx)
	if err != nil {
		return err
	}

	if o.quiet {
		fmt.Println(info.ID)
	} else {
		fmt.Printf("Renamed %s to %s\n", color.BlueString(info.ID), info.DisplayID())
	}
	return nil
}
