package main

import (
	"context"
	"fmt"
	"os"

	"github.com/allenai/beaker/config"
	"github.com/beaker/client/client"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var beaker *client.Client
var ctx context.Context
var quiet bool

func main() {
	var cancel context.CancelFunc
	ctx, cancel = withSignal(context.Background())
	defer cancel()

	errorPrefix := color.RedString("Error:")

	config, err := config.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %+v\n", errorPrefix, err)
		os.Exit(1)
	}

	beaker, err = client.NewClient(config.BeakerAddress, config.UserToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %+v\n", errorPrefix, err)
		os.Exit(1)
	}

	root := &cobra.Command{
		Use:   "beaker",
		Short: "Beaker is a tool for running machine learning experiments.",
		// TODO What do these do?
		// SilenceUsage: true,
		// SilenceErrors: true,
	}

	root.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode")

	root.AddCommand(newClusterCommand())
	root.AddCommand(newDatasetCommand())

	root.Execute()
}
