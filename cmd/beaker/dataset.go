package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/beaker/client/api"
	"github.com/beaker/client/client"
	"github.com/beaker/fileheap/bytefmt"
	"github.com/beaker/fileheap/cli"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newDatasetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dataset <command>",
		Short: "Manage datasets",
	}
	cmd.AddCommand(newDatasetCommitCommand())
	cmd.AddCommand(newDatasetCreateCommand())
	cmd.AddCommand(newDatasetDeleteCommand())
	cmd.AddCommand(newDatasetFetchCommand())
	cmd.AddCommand(newDatasetInspectCommand())
	cmd.AddCommand(newDatasetLsCommand())
	cmd.AddCommand(newDatasetRenameCommand())
	cmd.AddCommand(newDatasetStreamFileCommand())
	return cmd
}

func newDatasetCommitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "commit <dataset>",
		Short: "Commit a dataset preventing further modification",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dataset, err := beaker.Dataset(ctx, args[0])
			if err != nil {
				return err
			}

			if err := dataset.Commit(ctx); err != nil {
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

		workspace, err = ensureWorkspace(workspace)
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

func newDatasetDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <dataset>",
		Short: "Permanently delete a dataset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dataset, err := beaker.Dataset(ctx, args[0])
			if err != nil {
				return err
			}

			if err := dataset.Delete(ctx); err != nil {
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

	var outputPath string
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Target path for fetched data")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		dataset, err := beaker.Dataset(ctx, args[0])
		if err != nil {
			return err
		}

		fmt.Printf("Downloading %s to %s\n",
			color.CyanString(dataset.ID()),
			color.GreenString(outputPath+"/"))

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
		return cli.Download(ctx, dataset.Storage, "", outputPath, tracker, 32)
	}
	return cmd
}

func newDatasetInspectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect <dataset...>",
		Short: "Display detailed information about one or more datasets",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var datasets []api.Dataset
			for _, name := range args {
				dataset, err := beaker.Dataset(ctx, name)
				if err != nil {
					return err
				}

				info, err := dataset.Get(ctx)
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
	type fileInfo struct {
		Path    string    `json:"path"`
		Size    int64     `json:"size"`
		Updated time.Time `json:"updated"`
	}

	return &cobra.Command{
		Use:   "ls <dataset> <prefix?>",
		Short: "List files in a dataset",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			dataset, err := beaker.Dataset(ctx, args[0])
			if err != nil {
				return err
			}

			var totalFiles, totalBytes int64
			var prefix string
			if len(args) > 1 {
				prefix = args[1]
			}
			files, err := dataset.Files(ctx, prefix)
			if err != nil {
				return err
			}
			for {
				_, info, err := files.Next()
				if err == client.ErrDone {
					break
				}
				if err != nil {
					return err
				}
				totalFiles++
				totalBytes += info.Size

				switch format {
				case formatJSON:
					buf, err := json.Marshal(fileInfo{
						Path:    info.Path,
						Size:    info.Size,
						Updated: info.Updated,
					})
					if err != nil {
						return err
					}
					fmt.Println(string(buf))
				default:
					fmt.Printf(
						"%10s  %s  %s\n",
						bytefmt.FormatBytes(info.Size),
						info.Updated.Format(time.RFC3339),
						info.Path,
					)
				}
			}

			switch format {
			case formatJSON:
			default:
				fmt.Printf("Total: %d files, %s\n", totalFiles, bytefmt.FormatBytes(totalBytes))
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
			dataset, err := beaker.Dataset(ctx, args[0])
			if err != nil {
				return err
			}

			if err := dataset.SetName(ctx, args[1]); err != nil {
				return err
			}

			// TODO: This info should probably be part of the client response instead of a separate get.
			info, err := dataset.Get(ctx)
			if err != nil {
				return err
			}

			if quiet {
				fmt.Println(info.ID)
			} else {
				fmt.Printf("Renamed %s to %s\n", color.BlueString(info.ID), info.DisplayID())
			}
			return nil
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
		dataset, err := beaker.Dataset(ctx, args[0])
		if err != nil {
			return err
		}

		fileRef := dataset.FileRef(args[1])

		var r io.ReadCloser
		if offset != 0 || length != 0 {
			if length == 0 {
				// Length not specified; read the rest of the file.
				r, err = fileRef.DownloadRange(ctx, offset, -1)
			} else {
				r, err = fileRef.DownloadRange(ctx, offset, length)
			}
		} else {
			r, err = fileRef.Download(ctx)
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
