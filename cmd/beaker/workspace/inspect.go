package workspace

import (
	"context"
	"encoding/json"
	"os"

	"github.com/beaker/client/api"
	beaker "github.com/beaker/client/client"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type inspectOptions struct {
	workspaces []string
}

func newInspectCmd(
	parent *kingpin.CmdClause,
	parentOpts *workspaceOptions,
	config *config.Config,
) {
	o := &inspectOptions{}
	cmd := parent.Command("inspect", "Display detailed information about one or more workspaces")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Arg("workspace", "Workspace(s) to inspect").Required().StringsVar(&o.workspaces)
}

func (o *inspectOptions) run(beaker *beaker.Client) error {
	ctx := context.TODO()

	var workspaces []*api.Workspace
	for _, id := range o.workspaces {
		workspace, err := beaker.Workspace(ctx, id)
		if err != nil {
			return err
		}

		info, err := workspace.Get(ctx)
		if err != nil {
			return err
		}
		workspaces = append(workspaces, info)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "    ")
	return encoder.Encode(workspaces)
}
