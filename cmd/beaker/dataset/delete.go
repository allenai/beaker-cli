package dataset

import (
	"context"

	beaker "github.com/beaker/client/client"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

type deleteOptions struct {
	dataset string
}

func newDeleteCmd(
	parent *kingpin.CmdClause,
	parentOpts *datasetOptions,
	config *config.Config,
) {
	o := &deleteOptions{}
	cmd := parent.Command("delete", "Permanently delete a dataset")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Arg("dataset", "Dataset name or ID").Required().StringVar(&o.dataset)
}

func (o *deleteOptions) run(beaker *beaker.Client) error {
	ctx := context.TODO()

	dataset, err := beaker.Dataset(ctx, o.dataset)
	if err != nil {
		return err
	}

	return dataset.Delete(ctx)
}
