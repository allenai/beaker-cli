package experiment

import (
	"context"
	"fmt"
	"os"

	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	beaker "github.com/allenai/beaker/client"
	"github.com/allenai/beaker/config"
)

type stopOptions struct {
	experiments []string
}

func newStopCmd(
	parent *kingpin.CmdClause,
	parentOpts *experimentOptions,
	config *config.Config,
) {
	o := &stopOptions{}
	cmd := parent.Command("stop", "Stop one or more running experiments")
	cmd.Action(func(c *kingpin.ParseContext) error { return o.run(parentOpts, config.UserToken) })

	cmd.Arg("experiment", "Experiment name or ID").Required().StringsVar(&o.experiments)
}

func (o *stopOptions) run(parentOpts *experimentOptions, userToken string) error {
	ctx := context.TODO()
	beaker, err := beaker.NewClient(parentOpts.addr, userToken)
	if err != nil {
		return err
	}

	for _, name := range o.experiments {
		experiment, err := beaker.Experiment(ctx, name)
		if err != nil {
			return err
		}

		if err := experiment.Stop(ctx); err != nil {
			// We want to stop as many of the requested experiments as possible.
			// Therefore we print to STDERR here instead of returning.
			fmt.Fprintln(os.Stderr, color.RedString("Error:"), err)
		}

		fmt.Println(experiment.ID())
	}

	return nil
}
