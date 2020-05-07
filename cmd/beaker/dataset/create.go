package dataset

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/beaker/client/api"
	beaker "github.com/beaker/client/client"
	"github.com/beaker/fileheap/cli"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	configCmd "github.com/allenai/beaker/cmd/beaker/config"
	"github.com/allenai/beaker/config"
)

type createOptions struct {
	description string
	name        string
	quiet       bool
	sources     []string
	workspace   string
}

func newCreateCmd(
	parent *kingpin.CmdClause,
	parentOpts *datasetOptions,
	cfg *config.Config,
) {
	o := &createOptions{}
	cmd := parent.Command("create", "Create a new dataset")
	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, cfg.UserToken)
		if err != nil {
			return err
		}
		if o.workspace == "" {
			o.workspace, err = configCmd.EnsureDefaultWorkspace(beaker, cfg)
			if err != nil {
				return err
			}
			if !o.quiet {
				fmt.Printf("Using workspace %s\n", color.BlueString(o.workspace))
			}
		}
		return o.run(beaker)
	})

	cmd.Flag("desc", "Assign a description to the dataset").StringVar(&o.description)
	cmd.Flag("name", "Assign a name to the dataset").Short('n').StringVar(&o.name)
	cmd.Flag("quiet", "Only display created dataset's ID").Short('q').BoolVar(&o.quiet)
	cmd.Flag("workspace", "Workspace where the dataset will be placed").Short('w').StringVar(&o.workspace)
	cmd.Arg("sources", "List of globs resolving files or directories that should be uploaded").
		Required().StringsVar(&o.sources)
}

func (o *createOptions) run(beaker *beaker.Client) error {
	ctx := context.TODO()

	locations := make(map[string]os.FileInfo)
	for _, pattern := range o.sources {
		sources, err := filepath.Glob(pattern)
		if err != nil {
			return errors.Errorf("resolving glob pattern: %w", err)
		}
		for _, source := range sources {
			info, err := os.Stat(source)
			if err != nil {
				return err
			}
			if info.Mode()&(os.ModeSymlink|os.ModeNamedPipe|os.ModeSocket|os.ModeDevice) != 0 {
				return errors.Errorf("%s is a %s", source, modeToString(info.Mode()))
			}
			locations[source] = info
		}
	}

	spec := api.DatasetSpec{
		Description: o.description,
		Workspace:   o.workspace,
		FileHeap:    true,
	}
	dataset, err := beaker.CreateDataset(ctx, spec, o.name)
	if err != nil {
		return err
	}

	hasDirs := false
	for source, info := range locations {
		if !o.quiet {
			if o.name == "" {
				fmt.Printf("Uploading %s to %s\n", color.GreenString(source), color.CyanString(dataset.ID()))
			} else {
				fmt.Printf("Uploading %s to %s (%s)\n", color.GreenString(source), color.CyanString(o.name), dataset.ID())
			}
		}

		if info.IsDir() {
			hasDirs = true
			var tracker cli.ProgressTracker = cli.NoTracker
			if !o.quiet {
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
	}

	if err := dataset.Commit(ctx); err != nil {
		return errors.WithMessage(err, "failed to commit dataset")
	}

	if o.quiet {
		fmt.Println(dataset.ID())
	} else if !hasDirs {
		fmt.Println("Done.")
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
