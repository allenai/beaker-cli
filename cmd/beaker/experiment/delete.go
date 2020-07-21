package experiment

import (
	"context"
	"fmt"

	beaker "github.com/beaker/client/client"
	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type deleteOptions struct {
	experiment string
}

func newDeleteCmd(
	parent *kingpin.CmdClause,
	parentOpts *experimentOptions,
	config *config.Config,
) {
	o := &deleteOptions{}
	cmd := parent.Command("delete", "Permanently delete a experiment")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Arg("experiment", "Experiment name or ID").Required().StringVar(&o.experiment)
}

func (o *deleteOptions) run(beaker *beaker.Client) error {
	ctx := context.TODO()

	experiment, err := beaker.Experiment(ctx, o.experiment)
	if err != nil {
		return err
	}

	fmt.Printf("Deleted %s\n", color.BlueString(o.experiment))
	return experiment.Delete(ctx)
}
