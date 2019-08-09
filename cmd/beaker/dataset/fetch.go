package dataset

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/beaker/fileheap/cli"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

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
	if dataset.IsFile() {
		files, err := dataset.Files(ctx, "")
		if err != nil {
			return err
		}
		file, info, err := files.Next()
		if err != nil {
			return err
		}

		// Mimic 'cp' rules: Copying a file to a directory places the file into the target.
		if os.IsPathSeparator(target[len(target)-1]) {
			// The target ends in an explicit path separator, so must be a directory.
			// Stat will validate that paths ending in a sepator are directories.
			if _, err := os.Stat(target); err != nil && !os.IsNotExist(err) {
				return err
			}
			target = filepath.Join(target, info.Path)
		} else if f, err := os.Stat(target); err == nil && f.IsDir() {
			// The target exists and is a directory.
			target = filepath.Join(target, info.Path)
		}

		// Check again, but error on collision. This is a no-op if target is unmodified.
		if f, err := os.Stat(target); err == nil && f.IsDir() {
			return errors.Errorf("cannot overwrite directory %s with file %s", target, info.Path)
		}

		fmt.Printf("Downloading dataset %s to file %s ...", color.CyanString(dataset.ID()), color.GreenString(target))
		if err := file.DownloadTo(ctx, target); err != nil {
			fmt.Printf(" %s.\n", color.RedString("Failed"))
			return err
		}

		fmt.Println(" done.")
		return nil
	}

	fmt.Printf("Downloading %s to %s\n", color.CyanString(dataset.ID()), color.GreenString(target+"/"))
	return cli.Download(ctx, dataset.Storage, "", o.outputPath, cli.UnboundedTracker(ctx), 32)
}
