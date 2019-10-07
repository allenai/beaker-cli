package workspace

import (
	"context"
	"fmt"

	"github.com/beaker/client/api"
	beaker "github.com/beaker/client/client"
	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type createOptions struct {
	description string
	name        string
	quiet       bool
	org         string
}

func newCreateCmd(
	parent *kingpin.CmdClause,
	parentOpts *workspaceOptions,
	config *config.Config,
) {
	o := &createOptions{}
	cmd := parent.Command("create", "Create a new workspace")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		if o.org == "" {
			o.org = config.DefaultOrg
		}
		return o.run(beaker)
	})

	cmd.Flag("desc", "Assign a description to the workspace").StringVar(&o.description)
	cmd.Flag("quiet", "Only display created workspace's ID").Short('q').BoolVar(&o.quiet)
	cmd.Flag("org", "Org that will own the created workspace").Short('o').StringVar(&o.org)
	cmd.Arg("name", "The name of the workspace").Required().StringVar(&o.name)
}

func (o *createOptions) run(beaker *beaker.Client) error {
	ctx := context.TODO()

	spec := api.WorkspaceSpec{
		Name:         o.name,
		Description:  o.description,
		Organization: o.org,
	}

	workspace, err := beaker.CreateWorkspace(ctx, spec)
	if err != nil {
		return err
	}

	if o.quiet {
		fmt.Println(workspace.ID())
	} else {
		fmt.Printf("Workspace %s created (ID %s)\n", color.BlueString(spec.Name), color.BlueString(workspace.ID()))
	}
	return nil
}
