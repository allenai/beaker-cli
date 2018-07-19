package dataset

import (
	"context"
	"fmt"

	beaker "github.com/allenai/beaker-api/client"
	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker-cli/config"
)

type renameOptions struct {
	quiet   bool
	dataset string
	name    string
}

func newRenameCmd(
	parent *kingpin.CmdClause,
	parentOpts *datasetOptions,
	config *config.Config,
) {
	o := &renameOptions{}
	cmd := parent.Command("rename", "Rename an dataset")
	cmd.Action(func(c *kingpin.ParseContext) error { return o.run(parentOpts, config.UserToken) })

	cmd.Flag("quiet", "Only display the dataset's unique ID").Short('q').BoolVar(&o.quiet)
	cmd.Arg("dataset", "Name or ID of the dataset to rename").Required().StringVar(&o.dataset)
	cmd.Arg("new-name", "Unqualified name to assign to the dataset").Required().StringVar(&o.name)
}

func (o *renameOptions) run(parentOpts *datasetOptions, userToken string) error {
	ctx := context.TODO()
	beaker, err := beaker.NewClient(parentOpts.addr, userToken)
	if err != nil {
		return err
	}

	dataset, err := beaker.Dataset(ctx, o.dataset)
	if err != nil {
		return err
	}

	if err := dataset.SetName(ctx, o.name); err != nil {
		return err
	}

	// TODO: This info should probably be part of the client response instead of a separate get.
	info, err := dataset.Get(ctx)
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
