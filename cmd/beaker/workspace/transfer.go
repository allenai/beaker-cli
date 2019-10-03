package workspace

import (
	"context"
	"fmt"

	beaker "github.com/beaker/client/client"
	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type transferOptions struct {
	quiet     bool
	workspace string
	entities  []string
}

func newTransferCmd(
	parent *kingpin.CmdClause,
	parentOpts *workspaceOptions,
	config *config.Config,
) {
	o := &transferOptions{}
	cmd := parent.Command("transfer", "Transfer entities identified by IDs into a workspace")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Flag("quiet", "No console output unless an error occurs").Short('q').BoolVar(&o.quiet)
	cmd.Arg("workspace", "Destination workspace").Required().StringVar(&o.workspace)
	cmd.Arg("entities", "Entity IDs to transfer into the workspace").Required().StringsVar(&o.entities)
}

func (o *transferOptions) run(beaker *beaker.Client) error {
	if !o.quiet {
		fmt.Println(color.YellowString("Workspace commands are still under development and should be considered experimental."))
	}

	ctx := context.TODO()
	workspace, err := beaker.Workspace(ctx, o.workspace)
	if err != nil {
		return err
	}

	if err := workspace.Transfer(ctx, o.entities...); err != nil {
		return err
	}

	if !o.quiet {
		fmt.Printf("Transferred %d entities into workspace %s\n", len(o.entities), color.BlueString(workspace.ID()))
	}
	return nil
}
