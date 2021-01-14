package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/allenai/beaker/config"
	"github.com/beaker/client/api"
	"github.com/beaker/client/client"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// These variables are set externally by the linker.
var (
	version = "dev"
	commit  = "unknown"
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

	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the Beaker CLI version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Beaker %s ('%s')\n", version, commit)
			return nil
		},
	})

	root.AddCommand(newClusterCommand())
	root.AddCommand(newConfigCommand())
	root.AddCommand(newDatasetCommand())
	root.AddCommand(newExperimentCommand())
	root.AddCommand(newGroupCommand())
	root.AddCommand(newImageCommand())
	root.AddCommand(newSecretCommand())
	root.AddCommand(newTaskCommand())
	root.AddCommand(newWorkspaceCommand())

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

// Return a cancelable context which ends on signal interrupt.
//
// The first interrupt cancels the context, allowing callers to terminate
// gracefully. Upon receiving a second interrupt the process is terminated with
// exit code 130 (128 + SIGINT)
func withSignal(parent context.Context) (context.Context, context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	// In most cases this routine will leak due to the lack of a second signal.
	// That's OK since this is expected to last for the life of the process.
	go func() {
		select {
		case <-sigChan:
			cancel()
		case <-ctx.Done():
			// Do nothing.
		}
		<-sigChan
		os.Exit(130)
	}()

	return ctx, func() {
		signal.Stop(sigChan)
		cancel()
	}
}
