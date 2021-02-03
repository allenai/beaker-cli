package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"text/tabwriter"

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

var jsonOut *json.Encoder
var tableOut *tabwriter.Writer

func main() {
	jsonOut = json.NewEncoder(os.Stdout)
	jsonOut.SetIndent("", "    ")

	tableOut = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer tableOut.Flush()

	var cancel context.CancelFunc
	ctx, cancel = withSignal(context.Background())
	defer cancel()

	root := &cobra.Command{
		Use:           "beaker <command>",
		Short:         "Beaker is a tool for running machine learning experiments.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       fmt.Sprintf("Beaker %s (%q)", version, commit),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			if beakerConfig, err = config.New(); err != nil {
				return err
			}
			if beakerConfig.UserToken == "" {
				if err := login(); err != nil {
					return err
				}
			}

			beaker, err = client.NewClient(
				beakerConfig.BeakerAddress,
				beakerConfig.UserToken,
			)
			return err
		},
	}

	root.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode")
	root.PersistentFlags().StringVar(&format, "format", "", "Output format")

	root.AddCommand(newAccountCommand())
	root.AddCommand(newClusterCommand())
	root.AddCommand(newConfigCommand())
	root.AddCommand(newDatasetCommand())
	root.AddCommand(newExecutionCommand())
	root.AddCommand(newExperimentCommand())
	root.AddCommand(newGroupCommand())
	root.AddCommand(newImageCommand())
	root.AddCommand(newNodeCommand())
	root.AddCommand(newOrganizationCommand())
	root.AddCommand(newSecretCommand())
	root.AddCommand(newTaskCommand())
	root.AddCommand(newWorkspaceCommand())

	err := root.Execute()
	if err != nil && err.Error() == "invalid authentication token" {
		err = login()
		if err == nil {
			err = root.Execute()
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %+v\n", color.RedString("Error:"), err)
		os.Exit(1)
	}
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

// login prompts the user for an authentication token, validates it,
// and writes it to the configuration file.
func login() error {
	loginURL, err := url.Parse(beakerConfig.BeakerAddress)
	if err != nil {
		return err
	}
	loginURL.Path = path.Join(loginURL.Path, "user")

	fmt.Println(
		"You are not logged in. To log in, find your user token here:",
		color.BlueString(loginURL.String()),
	)
	fmt.Print("Enter your user token: ")
	reader := bufio.NewReader(os.Stdin)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		beakerConfig.UserToken = strings.TrimSpace(input)

		beaker, err = client.NewClient(
			beakerConfig.BeakerAddress,
			beakerConfig.UserToken,
		)
		if err != nil {
			return err
		}
		user, err := beaker.WhoAmI(ctx)
		if err != nil {
			fmt.Print("Invalid user token, please try again: ")
			continue
		}

		fmt.Printf("Successfully logged in as %q\n\n", user.Name)
		break
	}
	return config.WriteConfig(beakerConfig, config.GetFilePath())
}

// confirm prompts the user for a yes/no answer and defaults to no.
// Returns true, nil if the user enters yes.
func confirm(prompt string) (bool, error) {
	fmt.Print(prompt, " [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			return false, err
		}
		input = strings.TrimSuffix(input, "\n")
		input = strings.ToLower(input)
		switch input {
		case "y", "yes":
			return true, nil
		case "", "n", "no":
			return false, nil
		default:
			fmt.Print("Please type 'yes' or 'no': ")
		}
	}
}
