package main

import (
	"fmt"
	"os"

	"github.com/allenai/beaker/config"
	"github.com/beaker/client/client"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func main() {
	errorPrefix := color.RedString("Error:")

	config, err := config.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %+v\n", errorPrefix, err)
		os.Exit(1)
	}

	client, err := client.NewClient(config.BeakerAddress, config.UserToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %+v\n", errorPrefix, err)
		os.Exit(1)
	}

	root := &cobra.Command{
		Use:   "beaker",
		Short: "Beaker is a tool for running machine learning experiments.",
		// TODO What do these do?
		SilenceUsage: true,
		// SilenceErrors: true,
	}

	root.AddCommand(newClusterCommand(client))

	root.Execute()
}
