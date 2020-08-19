package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/cmd/beaker/cluster"
	configCmd "github.com/allenai/beaker/cmd/beaker/config"
	"github.com/allenai/beaker/cmd/beaker/dataset"
	"github.com/allenai/beaker/cmd/beaker/experiment"
	"github.com/allenai/beaker/cmd/beaker/group"
	"github.com/allenai/beaker/cmd/beaker/image"
	"github.com/allenai/beaker/cmd/beaker/options"
	"github.com/allenai/beaker/cmd/beaker/task"
	"github.com/allenai/beaker/cmd/beaker/workspace"
	"github.com/allenai/beaker/config"
)

func main() {
	errorPrefix := color.RedString("Error:")

	config, err := config.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %+v\n", errorPrefix, err)
		os.Exit(1)
	}

	if opts, err := newApp(config); err != nil {
		if opts.Debug {
			fmt.Fprintf(os.Stderr, "%s %+v\n", errorPrefix, err)
		} else {
			fmt.Fprintf(os.Stderr, "%s %v\n", errorPrefix, err)
		}
		os.Exit(1)
	}
}

// newApp creates a root application containing all Beaker subcommands.
func newApp(config *config.Config) (*options.AppOptions, error) {
	o := &options.AppOptions{}
	app := kingpin.New("beaker", "Beaker is a lab assistant to run and view experiments.")

	// Set a usage template to print better help messages.
	app.UsageTemplate(usageTemplate)

	// Disable interspersing flags with positional args.
	app.Interspersed(false)

	// Add global flags. These flags will also be available to sub-commands.
	app.HelpFlag.Short('h')
	app.Version(makeVersion())
	app.VersionFlag.Short('v')
	app.Flag("debug", "Print verbose stack traces on error.").BoolVar(&o.Debug)

	// Build out sub-command groups.
	cluster.NewClusterCmd(app, o, config)
	configCmd.NewConfigCmd(app, o, config)
	dataset.NewDatasetCmd(app, o, config)
	experiment.NewExperimentCmd(app, o, config)
	group.NewGroupCmd(app, o, config)
	image.NewImageCmd(app, o, config)
	task.NewTaskCmd(app, o, config)
	workspace.NewWorkspaceCmd(app, o, config)

	// Attach sub-commands.
	NewVersionCmd(app)

	// Parse command line input.
	_, err := app.Parse(os.Args[1:])
	return o, err
}
