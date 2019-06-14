package dataset

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/beaker/fileheap/cli"
	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/client"
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
	if dataset.Storage != nil {
		err := cli.Download(ctx, dataset.Storage, "", o.outputPath, cli.UnboundedTracker(ctx), 32)
		if err != nil {
			return err
		}
	} else {
		// Create target directory explicitly for empty datasets.
		if err := os.MkdirAll(target, 0755); err != nil {
			fmt.Printf(" %s.\n", color.RedString("Failed"))
			return err
		}

		files, err := dataset.Files(ctx, "")
		if err != nil {
			fmt.Printf(" %s.\n", color.RedString("Failed"))
			return err
		}

		for {
			file, info, err := files.Next()
			if err == client.ErrDone {
				break
			}
			if err != nil {
				fmt.Printf(" %s.\n", color.RedString("Failed"))
				return err
			}
			if err := file.DownloadTo(ctx, filepath.Join(target, info.Path)); err != nil {
				fmt.Printf(" %s.\n", color.RedString("Failed"))
				return err
			}
		}
		fmt.Println(" done.")
	}

	return nil
}
