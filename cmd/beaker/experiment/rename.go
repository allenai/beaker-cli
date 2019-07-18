package experiment

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	beaker "github.com/beaker/client/client"

	"github.com/allenai/beaker/config"
)

type renameOptions struct {
	quiet      bool
	experiment string
	name       string
}

func newRenameCmd(
	parent *kingpin.CmdClause,
	parentOpts *experimentOptions,
	config *config.Config,
) {
	o := &renameOptions{}
	cmd := parent.Command("rename", "Rename an experiment")
	cmd.Action(func(c *kingpin.ParseContext) error { return o.run(parentOpts, config.UserToken) })

	cmd.Flag("quiet", "Only display the experiment's unique ID").Short('q').BoolVar(&o.quiet)
	cmd.Arg("experiment", "Name or ID of the experiment to rename").Required().StringVar(&o.experiment)
	cmd.Arg("new-name", "Unqualified name to assign to the experiment").Required().StringVar(&o.name)
}

func (o *renameOptions) run(parentOpts *experimentOptions, userToken string) error {
	ctx := context.TODO()
	beaker, err := beaker.NewClient(parentOpts.addr, userToken)
	if err != nil {
		return err
	}

	experiment, err := beaker.Experiment(ctx, o.experiment)
	if err != nil {
		return err
	}

	if err := experiment.SetName(ctx, o.name); err != nil {
		return err
	}

	// TODO: This info should probably be part of the client response instead of a separate get.
	exp, err := experiment.Get(ctx)
	if err != nil {
		return err
	}

	if o.quiet {
		fmt.Println(exp.ID)
	} else {
		fmt.Printf("Renamed %s to %s\n", color.BlueString(exp.ID), exp.DisplayID())
	}
	return nil
}
