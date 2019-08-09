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
	manifest bool
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

	cmd.Flag("manifest", "Include file manifest in output").BoolVar(&o.manifest)
	cmd.Arg("dataset", "Dataset name or ID").Required().StringsVar(&o.datasets)
}

func (o *inspectOptions) run(beaker *beaker.Client) error {
	type detail struct {
		api.Dataset
		Manifest *api.DatasetManifest `json:"manifest,omitempty"`
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

		var manifest *api.DatasetManifest
		if o.manifest {
			manifest, err = dataset.Manifest(ctx)
			if err != nil {
				return err
			}
		}

		datasets = append(datasets, detail{*info, manifest})
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "    ")
	return encoder.Encode(datasets)
}
