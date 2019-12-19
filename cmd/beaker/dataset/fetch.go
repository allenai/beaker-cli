package dataset

import (
	"context"
	"fmt"

	"github.com/beaker/fileheap/cli"
	"github.com/fatih/color"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/beaker/client/client"

	"github.com/allenai/beaker/config"
)

type fetchOptions struct {
	dataset    string
	outputPath string
}

func newFetchCmd(
	parent *kingpin.CmdClause,
	parentOpts *datasetOptions,
	config *config.Config,
) {
	o := &fetchOptions{}
	cmd := parent.Command("fetch", "Fetch an existing dataset")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := client.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Flag("output", "Target path for fetched data").Required().Short('o').StringVar(&o.outputPath)
	cmd.Arg("dataset", "Dataset name or ID").Required().StringVar(&o.dataset)
}

func (o *fetchOptions) run(beaker *client.Client) error {
	ctx := context.TODO()
	dataset, err := beaker.Dataset(ctx, o.dataset)
	if err != nil {
		return err
	}

	target := o.outputPath
	fmt.Printf("Downloading %s to %s\n", color.CyanString(dataset.ID()), color.GreenString(target+"/"))
	var tracker cli.ProgressTracker
	info, err := dataset.Storage.Info(ctx)
	if err != nil {
		return err
	}
	if info.Size != nil && info.Size.Final {
		tracker = cli.BoundedTracker(ctx, info.Size.Files, info.Size.Bytes)
	} else {
		tracker = cli.UnboundedTracker(ctx)
	}
	return cli.Download(ctx, dataset.Storage, "", o.outputPath, tracker, 32)
}
