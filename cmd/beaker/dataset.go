package main

import (
	"fmt"
	"os"

	"github.com/beaker/client/api"
	"github.com/beaker/fileheap/cli"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newDatasetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dataset",
		Short: "Manage datasets",
	}
	cmd.AddCommand(newDatasetCreateCommand())
	return cmd
}

func newDatasetCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <source>",
		Short: "Create a new dataset",
		Args:  cobra.ExactArgs(1),
	}

	var description string
	var name string
	var workspace string

	cmd.Flags().StringVar(&description, "desc", "", "Assign a description to the dataset")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Assign a name to the dataset")
	cmd.Flags().StringVarP(&workspace, "workspace", "w", "", "Workspace where the dataset will be placed")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		source := args[0]

		info, err := os.Stat(source)
		if err != nil {
			return err
		}
		if info.Mode()&(os.ModeSymlink|os.ModeNamedPipe|os.ModeSocket|os.ModeDevice) != 0 {
			return errors.Errorf("%s is a %s", source, modeToString(info.Mode()))
		}

		spec := api.DatasetSpec{
			Description: description,
			Workspace:   workspace,
			FileHeap:    true,
		}

		dataset, err := beaker.CreateDataset(ctx, spec, name)
		if err != nil {
			return err
		}

		if !quiet {
			if name == "" {
				fmt.Printf("Uploading %s to %s\n", color.GreenString(source), color.CyanString(dataset.ID()))
			} else {
				fmt.Printf("Uploading %s to %s (%s)\n", color.GreenString(source), color.CyanString(name), dataset.ID())
			}
		}

		if info.IsDir() {
			var tracker cli.ProgressTracker = cli.NoTracker
			if !quiet {
				files, bytes, err := cli.UploadStats(source)
				if err != nil {
					return err
				}
				tracker = cli.BoundedTracker(ctx, files, bytes)
			}
			if err := cli.Upload(ctx, source, dataset.Storage, "", tracker, 32); err != nil {
				return err
			}
		} else {
			file, err := os.Open(source)
			if err != nil {
				return errors.WithStack(err)
			}
			defer func() { _ = file.Close() }()

			if err := dataset.Storage.WriteFile(ctx, info.Name(), file, info.Size()); err != nil {
				return err
			}
		}

		if err := dataset.Commit(ctx); err != nil {
			return errors.WithMessage(err, "failed to commit dataset")
		}

		if quiet {
			fmt.Println(dataset.ID())
		} else if !info.IsDir() {
			fmt.Println("Done.")
		}
		return nil
	}

	return cmd
}

func modeToString(mode os.FileMode) string {
	switch {
	case mode&os.ModeDir != 0:
		return "directory"
	case mode&os.ModeSymlink != 0:
		return "symbolic link"
	case mode&os.ModeNamedPipe != 0:
		return "named pipe"
	case mode&os.ModeSocket != 0:
		return "socket"
	case mode&os.ModeDevice != 0:
		return "device"
	default:
		return "file"
	}
}
