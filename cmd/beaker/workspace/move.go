package workspace

import (
	"context"
	"fmt"

	beaker "github.com/beaker/client/client"
	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type moveOptions struct {
	quiet     bool
	workspace string
	items     []string
}

func newMoveCmd(
	parent *kingpin.CmdClause,
	parentOpts *workspaceOptions,
	config *config.Config,
) {
	o := &moveOptions{}
	cmd := parent.Command("move", "Move items into a workspace")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Flag("quiet", "No console output unless an error occurs").Short('q').BoolVar(&o.quiet)
	cmd.Arg("workspace", "Destination workspace").Required().StringVar(&o.workspace)
	cmd.Arg("items", "IDs to transfer into the workspace").Required().StringsVar(&o.items)
}

func (o *moveOptions) run(beaker *beaker.Client) error {
	if !o.quiet {
		fmt.Println(color.YellowString("Workspace commands are still under development and should be considered experimental."))
	}

	ctx := context.TODO()
	workspace, err := beaker.Workspace(ctx, o.workspace)
	if err != nil {
		return err
	}

	if err := workspace.Transfer(ctx, o.items...); err != nil {
		return err
	}

	if !o.quiet {
		fmt.Printf("Transferred %d items into workspace %s\n", len(o.items), color.BlueString(workspace.ID()))
	}
	return nil
}
