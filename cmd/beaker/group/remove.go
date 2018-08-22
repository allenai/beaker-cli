package group

import (
	"context"
	"fmt"

	beaker "github.com/allenai/beaker-api/client"
	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type removeOptions struct {
	quiet         bool
	group         string
	experimentIDs []string
}

func newRemoveCmd(
	parent *kingpin.CmdClause,
	parentOpts *groupOptions,
	config *config.Config,
) {
	o := &removeOptions{}
	cmd := parent.Command("remove", "Remove experiments from an existing group")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Flag("quiet", "Only display the group's unique ID").Short('q').BoolVar(&o.quiet)
	cmd.Arg("group", "Group name or ID").Required().StringVar(&o.group)
	cmd.Arg("experiment", "ID of experiment to remove from the group").Required().StringsVar(&o.experimentIDs)
}

func (o *removeOptions) run(beaker *beaker.Client) error {
	ctx := context.TODO()
	group, err := beaker.Group(ctx, o.group)
	if err != nil {
		return err
	}

	ids := trimAndUnique(o.experimentIDs)
	if err := group.RemoveExperiments(ctx, ids); err != nil {
		return err
	}

	if o.quiet {
		fmt.Println(group.ID())
	} else {
		fmt.Printf("Removed experiments from %s: %s\n", color.BlueString(group.ID()), ids)
	}
	return nil
}
