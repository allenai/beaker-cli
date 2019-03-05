package dataset

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	beaker "github.com/allenai/beaker/client"
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
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})

	cmd.Flag("output", "Target path for fetched data").Required().Short('o').StringVar(&o.outputPath)
	cmd.Arg("dataset", "Dataset name or ID").Required().StringVar(&o.dataset)
}

func (o *fetchOptions) run(beaker *beaker.Client) error {
	ctx := context.TODO()
	dataset, err := beaker.Dataset(ctx, o.dataset)
	if err != nil {
		return err
	}

	return dataset.DownloadTo(ctx, o.outputPath)

	manifest, err := dataset.Manifest(ctx)
	if err != nil {
		return err
	}

	target := o.outputPath
	if manifest.SingleFile {
		if len(manifest.Files) == 0 {
			return errors.New("expected one file, but dataset is empty")
		}
		filename := manifest.Files[0].File

		// Mimic 'cp' rules: Copying a file to a directory places the file into the target.
		if os.IsPathSeparator(target[len(target)-1]) {
			// The target ends in an explicit path separator, so must be a directory.
			// Stat will validate that paths ending in a sepator are directories.
			if _, err := os.Stat(target); err != nil && !os.IsNotExist(err) {
				return err
			}
			target = filepath.Join(target, filename)
		} else if f, err := os.Stat(target); err == nil && f.IsDir() {
			// The target exists and is a directory.
			target = filepath.Join(target, filename)
		}

		// Check again, but error on collision. This is a no-op if target is unmodified.
		if f, err := os.Stat(target); err == nil && f.IsDir() {
			return errors.Errorf("cannot overwrite directory %s with file %s", target, filename)
		}

		if len(manifest.Files) == 0 {
			return errors.Errorf("file could not be found on remote server: %s", filename)
		}

		fmt.Printf("Downloading dataset %s to file %s ...", color.BlueString(manifest.ID), target)
		if err := dataset.FileRef(manifest.Files[0].File).DownloadTo(ctx, target); err != nil {
			fmt.Printf(" %s.\n", color.RedString("Failed"))
			return err
		}

		fmt.Println(" done.")
		return nil
	}

	fmt.Printf("Downloading dataset %s to directory %s/ ...", color.BlueString(manifest.ID), target)
	for _, file := range manifest.Files {
		fileTarget := filepath.Join(target, file.File)
		if err := dataset.FileRef(file.File).DownloadTo(ctx, fileTarget); err != nil {
			fmt.Printf(" %s.\n", color.RedString("Failed"))
			return err
		}
	}

	fmt.Println(" done.")
	return nil
}
