package group

import (
	"context"
	"fmt"

	beaker "github.com/beaker/client/client"
	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type deleteOptions struct {
	quiet bool
	group string
}

func newDeleteCmd(
	parent *kingpin.CmdClause,
	parentOpts *groupOptions,
	config *config.Config,
) {
	o := &deleteOptions{}
	cmd := parent.Command("delete", "Delete an experiment group")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Flag("quiet", "Only display the group's unique ID").Short('q').BoolVar(&o.quiet)
	cmd.Arg("group", "Group name or ID").Required().StringVar(&o.group)
}

func (o *deleteOptions) run(beaker *beaker.Client) error {
	ctx := context.TODO()
	group, err := beaker.Group(ctx, o.group)
	if err != nil {
		return err
	}

	if err := group.Delete(ctx); err != nil {
		return err
	}

	if o.quiet {
		fmt.Println(group.ID())
	} else {
		fmt.Println("Deleted group " + color.BlueString(group.ID()))
	}
	return nil
}
