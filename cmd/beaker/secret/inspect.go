package secret

import (
	"context"
	"encoding/json"
	"os"

	beaker "github.com/beaker/client/client"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type inspectOptions struct {
	workspace string
	name      string
}

func newInspectCmd(
	parent *kingpin.CmdClause,
	parentOpts *secretOptions,
	config *config.Config,
) {
	o := &inspectOptions{}
	cmd := parent.Command("inspect", "Inspect secret metadata")
	cmd.Flag("workspace", "Workspace containing the secret.").Required().StringVar(&o.workspace)
	cmd.Arg("name", "The name of the secret.").Required().StringVar(&o.name)

	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})
}

func (o *inspectOptions) run(beaker *beaker.Client) error {
	ctx := context.Background()
	workspace, err := beaker.Workspace(ctx, o.workspace)
	if err != nil {
		return err
	}

	secret, err := workspace.GetSecret(ctx, o.name)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "    ")
	return encoder.Encode(secret)
}
