package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/allenai/beaker/config"
	"github.com/beaker/client/api"
	"github.com/beaker/client/client"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var beaker *client.Client
var beakerConfig *config.Config
var ctx context.Context
var quiet bool
var format string

const (
	formatJSON = "json"
)

func main() {
	var cancel context.CancelFunc
	ctx, cancel = withSignal(context.Background())
	defer cancel()

	errorPrefix := color.RedString("Error:")

	var err error
	beakerConfig, err = config.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %+v\n", errorPrefix, err)
		os.Exit(1)
	}

	beaker, err = client.NewClient(beakerConfig.BeakerAddress, beakerConfig.UserToken)
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
	root.PersistentFlags().StringVarP(&format, "format", "f", "", "Output format")

	root.AddCommand(newClusterCommand())
	root.AddCommand(newDatasetCommand())
	root.AddCommand(newExperimentCommand())
	root.AddCommand(newGroupCommand())
	root.AddCommand(newImageCommand())
	root.AddCommand(newSecretCommand())

	root.Execute()
}

// ensureWorkspace ensures that workspaceRef exists or that the default workspace
// exists if workspaceRef is empty.
// Returns an error if workspaceRef and the default workspace are empty.
func ensureWorkspace(workspaceRef string) (string, error) {
	if workspaceRef == "" {
		if beakerConfig.DefaultWorkspace == "" {
			return "", errors.New(`workspace not provided, either:
1. Pass the --workspace flag
2. Configure a default workspace with 'beaker config set default_workspace <workspace>'`)
		}
		workspaceRef = beakerConfig.DefaultWorkspace
	}

	// Create the workspace if it doesn't exist.
	if _, err := beaker.Workspace(ctx, workspaceRef); err != nil {
		if apiErr, ok := err.(api.Error); ok && apiErr.Code == http.StatusNotFound {
			parts := strings.Split(workspaceRef, "/")
			if len(parts) != 2 {
				return "", errors.New("workspace must be formatted like '<account>/<name>'")
			}

			if _, err = beaker.CreateWorkspace(ctx, api.WorkspaceSpec{
				Organization: parts[0],
				Name:         parts[1],
			}); err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}

	return workspaceRef, nil
}
