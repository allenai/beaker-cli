package task

import (
	"context"
	"encoding/json"
	"os"

	"github.com/allenai/beaker-api/api"
	beaker "github.com/allenai/beaker-api/client"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker-cli/config"
)

type inspectOptions struct {
	ids []string
}

func newInspectCmd(
	parent *kingpin.CmdClause,
	parentOpts *experimentOptions,
	config *config.Config,
) {
	o := &inspectOptions{}
	cmd := parent.Command("inspect", "Display detailed information about one or more tasks")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Arg("id", "Task IDs").Required().StringsVar(&o.ids)
}

func (o *inspectOptions) run(beaker *beaker.Client) error {
	ctx := context.TODO()

	var tasks []*api.Task
	for _, id := range o.ids {
		task, err := beaker.Task(ctx, id)
		if err != nil {
			return err
		}

		info, err := task.Get(ctx)
		if err != nil {
			return err
		}

		tasks = append(tasks, info)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "    ")
	return encoder.Encode(tasks)
}
