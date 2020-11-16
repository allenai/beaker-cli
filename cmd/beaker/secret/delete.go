package secret

import (
	beaker "github.com/beaker/client/client"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type deleteOptions struct {
	workspace string
	name      string
}

func newDeleteCmd(
	parent *kingpin.CmdClause,
	parentOpts *secretOptions,
	config *config.Config,
) {
	o := &deleteOptions{}
	cmd := parent.Command("delete", "Permanently delete a secret")
	cmd.Flag("workspace", "Workspace containing the secret").Required().StringVar(&o.workspace)
	cmd.Arg("name", "The name of the secret").Required().StringVar(&o.name)

	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})
}

func (o *deleteOptions) run(beaker *beaker.Client) error {
	// TODO Delete the secret
	return nil
}
