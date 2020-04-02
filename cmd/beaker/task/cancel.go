package task

import (
	"context"
	"fmt"
	"os"

	beaker "github.com/beaker/client/client"
	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type cancelOptions struct {
	ids []string
}

func newCancelCmd(
	parent *kingpin.CmdClause,
	parentOpts *experimentOptions,
	config *config.Config,
) {
	o := &cancelOptions{}
	cmd := parent.Command("cancel", "Stop one or more running tasks")
	cmd.Action(func(c *kingpin.ParseContext) error { return o.run(parentOpts, config.UserToken) })

	cmd.Arg("id", "Task ID").Required().StringsVar(&o.ids)
}

func (o *cancelOptions) run(parentOpts *experimentOptions, userToken string) error {
	ctx := context.TODO()
	beaker, err := beaker.NewClient(parentOpts.addr, userToken)
	if err != nil {
		return err
	}

	for _, id := range o.ids {
		task := beaker.Task(id)
		if err := task.Stop(ctx); err != nil {
			// We want to cancel as many of the requested tasks as possible.
			// Therefore we print to STDERR here instead of returning.
			fmt.Fprintln(os.Stderr, color.RedString("Error:"), err)
		}

		fmt.Println(task.ID())
	}

	return nil
}
