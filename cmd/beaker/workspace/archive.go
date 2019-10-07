package workspace

import (
	"context"
	"fmt"

	beaker "github.com/beaker/client/client"
	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type archiveOptions struct {
	workspace string
}

func newArchiveCmd(
	parent *kingpin.CmdClause,
	parentOpts *workspaceOptions,
	config *config.Config,
) {
	o := &archiveOptions{}
	cmd := parent.Command("archive", "Archive a workspace, making it read-only")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Arg("workspace", "Workspace to archive").Required().StringVar(&o.workspace)
}

func (o *archiveOptions) run(beaker *beaker.Client) error {
	ctx := context.TODO()
	workspace, err := beaker.Workspace(ctx, o.workspace)
	if err != nil {
		return err
	}

	err = workspace.SetArchived(ctx, true)
	if err != nil {
		return err
	}

	fmt.Printf("Workspace %s archived\n", color.BlueString(o.workspace))
	return nil
}
