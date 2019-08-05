package workspace

import (
	"context"
	"fmt"

	beaker "github.com/beaker/client/client"
	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type unArchiveOptions struct {
	workspace string
}

func newUnarchiveCmd(
	parent *kingpin.CmdClause,
	parentOpts *workspaceOptions,
	config *config.Config,
) {
	o := &unArchiveOptions{}
	cmd := parent.Command("unarchive", "Un-archive a workspace")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Arg("workspace", "Workspace ID (name not yet supported)").Required().StringVar(&o.workspace)
}

func (o *unArchiveOptions) run(beaker *beaker.Client) error {
	fmt.Println(color.RedString("Workspace commands are still under development and should be considered experimental."))

	ctx := context.TODO()

	workspace, err := beaker.Workspace(ctx, o.workspace)
	if err != nil {
		return err
	}

	err = workspace.SetArchived(ctx, false)
	if err != nil {
		return err
	}

	fmt.Printf("Workspace %s un-archived\n", color.BlueString(o.workspace))
	return nil
}
