package secret

import (
	beaker "github.com/beaker/client/client"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type listOptions struct{}

func newListCmd(
	parent *kingpin.CmdClause,
	parentOpts *secretOptions,
	config *config.Config,
) {
	o := &listOptions{}
	cmd := parent.Command("list", "List the metadata of all secrets in a workspace")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})
}

func (o *listOptions) run(beaker *beaker.Client) error {
	// TODO List secret metadata
	return nil
}
