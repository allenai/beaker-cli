package image

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	beaker "github.com/allenai/beaker/client"
	"github.com/allenai/beaker/config"
)

// RenameOptions defines settings for image rename command
// TODO: make RenameOptions and fields unexported once not needed by blueprint command
type RenameOptions struct {
	Quiet bool
	Image string
	Name  string
}

func newRenameCmd(
	parent *kingpin.CmdClause,
	parentOpts *CmdOptions,
	config *config.Config,
) {
	o := &RenameOptions{}
	cmd := parent.Command("rename", "Rename an image")
	cmd.Action(func(c *kingpin.ParseContext) error { return o.Run(parentOpts, config.UserToken) })

	cmd.Flag("quiet", "Only display the image's unique ID").Short('q').BoolVar(&o.Quiet)
	cmd.Arg("image", "Name or ID of the image to rename").Required().StringVar(&o.Image)
	cmd.Arg("new-name", "Unqualified name to assign to the image").Required().StringVar(&o.Name)
}

// Run executes beaker image rename command
// TODO: make Run unexported once not needed by blueprint command
func (o *RenameOptions) Run(parentOpts *CmdOptions, userToken string) error {
	ctx := context.TODO()
	beaker, err := beaker.NewClient(parentOpts.Addr, userToken)
	if err != nil {
		return err
	}
	image, err := beaker.Image(ctx, o.Image)
	if err != nil {
		return err
	}

	if err := image.SetName(ctx, o.Name); err != nil {
		return err
	}

	// TODO: This info should probably be part of the client response instead of a separate get.
	info, err := image.Get(ctx)
	if err != nil {
		return err
	}

	if o.Quiet {
		fmt.Println(info.ID)
	} else {
		fmt.Printf("Renamed %s to %s\n", color.BlueString(info.ID), info.DisplayID())
	}
	return nil
}
