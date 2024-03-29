package main

import (
	"fmt"
	"io"
	"os"

	"github.com/allenai/bytefmt"
	"github.com/beaker/client/api"
	"github.com/beaker/client/client"
	fileheapAPI "github.com/beaker/fileheap/api"
	"github.com/beaker/fileheap/cli"
	fileheap "github.com/beaker/fileheap/client"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const defaultConcurrency = 8

func newDatasetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dataset <command>",
		Short: "Manage datasets",
	}
	cmd.AddCommand(newDatasetCommitCommand())
	cmd.AddCommand(newDatasetCreateCommand())
	cmd.AddCommand(newDatasetDeleteCommand())
	cmd.AddCommand(newDatasetFetchCommand())
	cmd.AddCommand(newDatasetGetCommand())
	cmd.AddCommand(newDatasetLsCommand())
	cmd.AddCommand(newDatasetRenameCommand())
	cmd.AddCommand(newDatasetSizeCommand())
	cmd.AddCommand(newDatasetStreamFileCommand())
	cmd.AddCommand(newDatasetSyncCommand())
	return cmd
}

func newDatasetCommitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "commit <dataset>",
		Short: "Commit a dataset preventing further modification",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := beaker.Dataset(args[0]).Patch(ctx, api.DatasetPatch{
				Commit: true,
			}); err != nil {
				return err
			}

			if !quiet {
				fmt.Printf("Committed %s\n", color.BlueString(args[0]))
			}
			return nil
		},
	}
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
	var concurrency int
	var commit bool
	cmd.Flags().StringVar(&description, "desc", "", "Assign a description to the dataset")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Assign a name to the dataset")
	cmd.Flags().StringVarP(&workspace, "workspace", "w", "", "Workspace where the dataset will be placed")
	cmd.Flags().IntVar(
		&concurrency,
		"concurrency",
		defaultConcurrency,
		"Number of files to upload at a time")
	cmd.Flags().BoolVar(&commit, "commit", true, "Commit the dataset after successful upload")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		source := args[0]

		workspace, err := ensureWorkspace(workspace)
		if err != nil {
			return err
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

		if err := syncDataset(source, dataset, concurrency); err != nil {
			return err
		}

		if commit {
			if _, err := dataset.Patch(ctx, api.DatasetPatch{
				Commit: true,
			}); err != nil {
				return errors.WithMessage(err, "failed to commit dataset")
			}
		}

		if quiet {
			fmt.Println(dataset.Ref())
		} else {
			fmt.Println("Done.")
		}
		return nil
	}
	return cmd
}

func newDatasetDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <dataset>",
		Short: "Permanently delete a dataset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := beaker.Dataset(args[0]).Delete(ctx); err != nil {
				return err
			}

			if !quiet {
				fmt.Printf("Deleted %s\n", color.BlueString(args[0]))
			}
			return nil
		},
	}
}

func newDatasetFetchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch <dataset>",
		Short: "Download a dataset",
		Args:  cobra.ExactArgs(1),
	}
	flags := addFetchFlags(cmd)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return fetchDataset(args[0], flags.outputPath, flags.prefix, flags.concurrency)
	}
	return cmd
}

type fetchFlags struct {
	outputPath  string
	prefix      string
	concurrency int
}

func addFetchFlags(cmd *cobra.Command) *fetchFlags {
	flags := &fetchFlags{}
	cmd.Flags().StringVarP(&flags.outputPath, "output", "o", ".", "Target path for fetched data")
	cmd.Flags().StringVar(&flags.prefix, "prefix", "", "Only download files that start with the given prefix")
	cmd.Flags().IntVar(
		&flags.concurrency,
		"concurrency",
		defaultConcurrency,
		"Number of files to download at a time")
	return flags
}

func fetchDataset(
	dataset string,
	outputPath string,
	prefix string,
	concurrency int,
) error {
	storage, _, err := beaker.Dataset(dataset).Storage(ctx)
	if err != nil {
		return err
	}

	info, err := storage.Info(ctx)
	if err != nil {
		return err
	}

	if !quiet {
		fmt.Printf("Downloading dataset %s to %s\n",
			color.CyanString(dataset),
			color.GreenString(outputPath))
	}
	var tracker cli.ProgressTracker
	if quiet {
		tracker = cli.NoTracker
	} else if info.Size != nil && info.Size.Final {
		tracker = cli.BoundedTracker(ctx, info.Size.Files, info.Size.Bytes)
	} else {
		tracker = cli.UnboundedTracker(ctx)
	}
	return cli.Download(ctx, storage, prefix, outputPath, tracker, concurrency)
}

func newDatasetGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "get <dataset...>",
		Aliases: []string{"inspect"},
		Short:   "Display detailed information about one or more datasets",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var datasets []api.Dataset
			for _, name := range args {
				info, err := beaker.Dataset(name).Get(ctx)
				if err != nil {
					return err
				}

				datasets = append(datasets, *info)
			}
			return printDatasets(datasets)
		},
	}
}

func newDatasetLsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ls <dataset> [prefix]",
		Short: "List files in a dataset",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			storage, _, err := beaker.Dataset(args[0]).Storage(ctx)
			if err != nil {
				return err
			}

			var files []*fileheapAPI.FileInfo
			var prefix string
			if len(args) > 1 {
				prefix = args[1]
			}

			iterator := storage.Files(ctx, &fileheap.FileIteratorOptions{Prefix: prefix})
			for {
				info, err := iterator.Next()
				if err == fileheap.ErrDone {
					break
				}
				if err != nil {
					return err
				}
				files = append(files, info)
			}

			switch format {
			case formatJSON:
				return printJSON(files)
			default:
				if err := printTableRow(
					"PATH",
					"SIZE",
					"UPDATED",
				); err != nil {
					return err
				}
				for _, file := range files {
					if err := printTableRow(
						file.Path,
						bytefmt.New(file.Size, bytefmt.Binary),
						file.Updated,
					); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}
}

func newDatasetRenameCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <dataset> <name>",
		Short: "Rename a dataset",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			oldName := args[0]
			newName := args[1]

			dataset, err := beaker.Dataset(oldName).Patch(ctx, api.DatasetPatch{
				Name: &newName,
			})
			if err != nil {
				return err
			}

			if quiet {
				fmt.Println(dataset.ID)
			} else {
				fmt.Printf("Renamed %s to %s\n", color.BlueString(oldName), dataset.FullName)
			}
			return nil
		},
	}
}

func newDatasetSizeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "size <dataset> [prefix]",
		Short: "Calculate the size of a dataset",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			storage, _, err := beaker.Dataset(args[0]).Storage(ctx)
			if err != nil {
				return err
			}

			var totalFiles, totalBytes int64
			var prefix string
			if len(args) > 1 {
				prefix = args[1]
			}

			iterator := storage.Files(ctx, &fileheap.FileIteratorOptions{Prefix: prefix})
			for {
				info, err := iterator.Next()
				if err == fileheap.ErrDone {
					break
				}
				if err != nil {
					return err
				}
				totalFiles++
				totalBytes += info.Size
			}

			switch format {
			case formatJSON:
				type size struct {
					Files int64 `json:"files"`
					Bytes int64 `json:"bytes"`
				}
				return printJSON(size{
					Files: totalFiles,
					Bytes: totalBytes,
				})
			default:
				if err := printTableRow(
					"FILES",
					"SIZE",
				); err != nil {
					return err
				}
				return printTableRow(
					totalFiles,
					bytefmt.New(totalBytes, bytefmt.Binary),
				)
			}
		},
	}
}

func newDatasetStreamFileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stream-file <dataset> <file>",
		Short: "Stream a single file from an existing dataset to stdout",
		Args:  cobra.ExactArgs(2),
	}

	var offset int64
	var length int64
	cmd.Flags().Int64Var(&offset, "offset", 0, "Offset in bytes")
	cmd.Flags().Int64Var(&length, "length", 0, "Number of bytes to read")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		fileName := args[1]
		storage, _, err := beaker.Dataset(args[0]).Storage(ctx)
		if err != nil {
			return err
		}

		var r io.ReadCloser
		if offset != 0 || length != 0 {
			if length == 0 {
				// Length not specified; read the rest of the file.
				r, err = storage.ReadFileRange(ctx, fileName, offset, -1)
			} else {
				r, err = storage.ReadFileRange(ctx, fileName, offset, length)
			}
		} else {
			r, err = storage.ReadFile(ctx, fileName)
		}
		if err != nil {
			return err
		}
		defer r.Close()

		_, err = io.Copy(os.Stdout, r)
		return err
	}
	return cmd
}

func newDatasetSyncCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync <source> <target>",
		Short: "Sync the source to an uncommitted dataset",
		Args:  cobra.ExactArgs(2),
	}

	var concurrency int
	cmd.Flags().IntVar(
		&concurrency,
		"concurrency",
		defaultConcurrency,
		"Number of files to upload at a time")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		source := args[0]
		target := args[1]

		if err := syncDataset(source, beaker.Dataset(target), concurrency); err != nil {
			return err
		}

		if quiet {
			fmt.Println(target)
		} else {
			fmt.Println("Done.")
		}
		return nil
	}
	return cmd
}

func syncDataset(source string, target *client.DatasetHandle, concurrency int) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	if info.Mode()&(os.ModeSymlink|os.ModeNamedPipe|os.ModeSocket|os.ModeDevice) != 0 {
		return errors.Errorf("%s is a %s", source, modeToString(info.Mode()))
	}

	if !quiet {
		fmt.Printf("Uploading %s to %s\n", color.GreenString(source), color.CyanString(target.Ref()))
	}

	storage, _, err := target.Storage(ctx)
	if err != nil {
		return err
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
		if err := cli.Upload(ctx, source, storage, "", tracker, concurrency); err != nil {
			return err
		}
	} else {
		file, err := os.Open(source)
		if err != nil {
			return errors.WithStack(err)
		}
		defer func() { _ = file.Close() }()

		if err := storage.WriteFile(ctx, info.Name(), file, info.Size()); err != nil {
			return err
		}
	}
	return nil
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
