package secret

import (
	"context"
	"encoding/json"
	"os"

	beaker "github.com/beaker/client/client"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type listOptions struct {
	workspace string
}

func newListCmd(
	parent *kingpin.CmdClause,
	parentOpts *secretOptions,
	config *config.Config,
) {
	o := &listOptions{}
	cmd := parent.Command("list", "List the metadata of all secrets in a workspace")
	cmd.Flag("workspace", "Workspace to list secrets.").Required().StringVar(&o.workspace)
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})
}

func (o *listOptions) run(beaker *beaker.Client) error {
	ctx := context.Background()
	workspace, err := beaker.Workspace(ctx, o.workspace)
	if err != nil {
		return err
	}

	secrets, err := workspace.ListSecrets(ctx)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "    ")
	return encoder.Encode(secrets)
}
