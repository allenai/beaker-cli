package secret

import (
	"context"
	"fmt"

	beaker "github.com/beaker/client/client"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type readOptions struct {
	workspace string
	name      string
}

func newReadCmd(
	parent *kingpin.CmdClause,
	parentOpts *secretOptions,
	config *config.Config,
) {
	o := &readOptions{}
	cmd := parent.Command("read", "Read the value of a secret")
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

func (o *readOptions) run(beaker *beaker.Client) error {
	ctx := context.Background()
	workspace, err := beaker.Workspace(ctx, o.workspace)
	if err != nil {
		return err
	}

	secret, err := workspace.ReadSecret(ctx, o.name)
	if err != nil {
		return err
	}
	fmt.Printf("%s", secret)
	return nil
}
