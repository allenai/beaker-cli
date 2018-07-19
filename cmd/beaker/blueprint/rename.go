package blueprint

import (
	"context"
	"fmt"

	beaker "github.com/allenai/beaker-api/client"
	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker-cli/config"
)

type renameOptions struct {
	quiet     bool
	blueprint string
	name      string
}

func newRenameCmd(
	parent *kingpin.CmdClause,
	parentOpts *blueprintOptions,
	config *config.Config,
) {
	o := &renameOptions{}
	cmd := parent.Command("rename", "Rename an blueprint")
	cmd.Action(func(c *kingpin.ParseContext) error { return o.run(parentOpts, config.UserToken) })

	cmd.Flag("quiet", "Only display the blueprint's unique ID").Short('q').BoolVar(&o.quiet)
	cmd.Arg("blueprint", "Name or ID of the blueprint to rename").Required().StringVar(&o.blueprint)
	cmd.Arg("new-name", "Unqualified name to assign to the blueprint").Required().StringVar(&o.name)
}

func (o *renameOptions) run(parentOpts *blueprintOptions, userToken string) error {
	ctx := context.TODO()
	beaker, err := beaker.NewClient(parentOpts.addr, userToken)
	if err != nil {
		return err
	}
	blueprint, err := beaker.Blueprint(ctx, o.blueprint)
	if err != nil {
		return err
	}

	if err := blueprint.SetName(ctx, o.name); err != nil {
		return err
	}

	// TODO: This info should probably be part of the client response instead of a separate get.
	info, err := blueprint.Get(ctx)
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
