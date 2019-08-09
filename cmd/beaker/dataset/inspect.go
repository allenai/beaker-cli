package dataset

import (
	"context"
	"encoding/json"
	"os"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/beaker/client/api"
	beaker "github.com/beaker/client/client"

	"github.com/allenai/beaker/config"
)

type inspectOptions struct {
	datasets []string
}

func newInspectCmd(
	parent *kingpin.CmdClause,
	parentOpts *datasetOptions,
	config *config.Config,
) {
	o := &inspectOptions{}
	cmd := parent.Command("inspect", "Display detailed information about one or more datasets")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Arg("dataset", "Dataset name or ID").Required().StringsVar(&o.datasets)
}

func (o *inspectOptions) run(beaker *beaker.Client) error {
	type detail struct {
		api.Dataset
	}

	ctx := context.TODO()

	var datasets []detail
	for _, name := range o.datasets {
		dataset, err := beaker.Dataset(ctx, name)
		if err != nil {
			return err
		}

		info, err := dataset.Get(ctx)
		if err != nil {
			return err
		}

		datasets = append(datasets, detail{*info})
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "    ")
	return encoder.Encode(datasets)
}
