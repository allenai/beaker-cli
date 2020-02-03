package group

import (
	"context"
	"fmt"

	beaker "github.com/beaker/client/client"
	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type renameOptions struct {
	quiet bool
	group string
	name  string
}

func newRenameCmd(
	parent *kingpin.CmdClause,
	parentOpts *groupOptions,
	config *config.Config,
) {
	o := &renameOptions{}
	cmd := parent.Command("rename", "Rename an group")
	cmd.Action(func(c *kingpin.ParseContext) error { return o.run(parentOpts, config.UserToken) })

	cmd.Flag("quiet", "Only display the group's unique ID").Short('q').BoolVar(&o.quiet)
	cmd.Arg("group", "Name or ID of the group to rename").Required().StringVar(&o.group)
	cmd.Arg("new-name", "Unqualified name to assign to the group").Required().StringVar(&o.name)
}

func (o *renameOptions) run(parentOpts *groupOptions, userToken string) error {
	ctx := context.TODO()
	beaker, err := beaker.NewClient(parentOpts.addr, userToken)
	if err != nil {
		return err
	}

	group, err := beaker.Group(ctx, o.group)
	if err != nil {
		return err
	}

	if err := group.SetName(ctx, o.name); err != nil {
		return err
	}

	// TODO: This info should probably be part of the client response instead of a separate get.
	info, err := group.Get(ctx)
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
