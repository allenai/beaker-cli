package workspace

import (
	"context"
	"fmt"

	beaker "github.com/beaker/client/client"
	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type renameOptions struct {
	quiet     bool
	workspace string
	name      string
}

func newRenameCmd(
	parent *kingpin.CmdClause,
	parentOpts *workspaceOptions,
	config *config.Config,
) {
	o := &renameOptions{}
	cmd := parent.Command("rename", "Rename a workspace")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Flag("quiet", "Only display the workspace's unique ID").Short('q').BoolVar(&o.quiet)
	cmd.Arg("workspace", "Workspace to rename").Required().StringVar(&o.workspace)
	cmd.Arg("new-name", "Unqualified name to assign to the workspace").Required().StringVar(&o.name)
}

func (o *renameOptions) run(beaker *beaker.Client) error {
	ctx := context.TODO()
	workspace, err := beaker.Workspace(ctx, o.workspace)
	if err != nil {
		return err
	}

	if err := workspace.SetName(ctx, o.name); err != nil {
		return err
	}

	// TODO: This info should probably be part of the client response instead of a separate get.
	info, err := workspace.Get(ctx)
	if err != nil {
		return err
	}

	if o.quiet {
		fmt.Println(info.ID)
	} else {
		fmt.Printf("Renamed %s to %s\n", color.BlueString(o.workspace), color.BlueString(info.Name))
	}
	return nil
}
