package config

import (
	"fmt"
	"context"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	beaker "github.com/beaker/client/client"

	"github.com/allenai/beaker/config"
)

type selectOptions struct {
	key string
	value string
}

func newSetCmd(
	parent *kingpin.CmdClause,
	parentOpts *configOptions,
	config *config.Config,
) {
	o := &selectOptions{}
	cmd := parent.Command("set", "Set a specific config setting, identified by its YAML key")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Arg("key", "Key").Required().StringVar(&o.key)
	cmd.Arg("value", "Value").Required().StringVar(&o.value)
}

func (o *selectOptions) run(beaker *beaker.Client) error {
	ctx := context.Background()
	fmt.Println(ctx)
	return nil
}
