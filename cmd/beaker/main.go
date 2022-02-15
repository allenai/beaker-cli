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
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

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

			beaker, err = client.NewClient(
				beakerConfig.BeakerAddress,
				beakerConfig.UserToken,
			)
			if beakerConfig.HTTPDiag {
				beaker.HTTPResponseHook = func(resp *http.Response, duration time.Duration) {
					durationMs := duration.Nanoseconds() / 1000000
					fmt.Fprintf(os.Stderr,
						"Beaker HTTP diagnostic: status_code = %d duration_ms = %d request = %s %s\n",
						resp.StatusCode, durationMs, resp.Request.Method, resp.Request.URL.String(),
					)
				}
			}

			if quiet && format != "" {
				return errors.New("flags --quiet and --format are mutually exclusive")
			}

			return err
		},
	}

	root.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode")
	root.PersistentFlags().StringVar(&format, "format", "", "Output format")

	root.AddCommand(newAccountCommand())
	root.AddCommand(newClusterCommand())
	root.AddCommand(newConfigCommand())
	root.AddCommand(newDatasetCommand())
	root.AddCommand(newExecutorCommand())
	root.AddCommand(newExperimentCommand())
	root.AddCommand(newGroupCommand())
	root.AddCommand(newImageCommand())
	root.AddCommand(newJobCommand())
	root.AddCommand(newNodeCommand())
	root.AddCommand(newOrganizationCommand())
	root.AddCommand(newSecretCommand())
	root.AddCommand(newSessionCommand())
	root.AddCommand(newWorkspaceCommand())

	jsonOut = json.NewEncoder(os.Stdout)
	jsonOut.SetIndent("", "    ")
	tableOut = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	err := root.Execute()
	tableOut.Flush()
	if err != nil {
		var apiErr api.Error
		if errors.As(err, &apiErr) && apiErr.Code == http.StatusUnauthorized {
			err = login()
			if err == nil {
				err = root.Execute()
			}
		}
	}
	if err != nil {
		// Don't print "context canceled" error on Ctrl-C.
		if !errors.Is(err, context.Canceled) {
			fmt.Fprintf(os.Stderr, "%s %+v\n", color.RedString("Error:"), err)
		}
		os.Exit(1)
	}
}

// ensureWorkspace ensures that workspaceRef exists or that the default workspace
// exists if workspaceRef is empty.
// Returns an error if workspaceRef and the default workspace are empty.
func ensureWorkspace(workspaceRef string) (string, error) {
	// If this flag is true, workspaceRef is written to config as the default workspace on exit.
	var updateConfig bool
	if workspaceRef == "" {
		if beakerConfig.DefaultWorkspace == "" {
			user, err := beaker.WhoAmI(ctx)
			if err != nil {
				return "", err
			}
			orgs, err := beaker.ListMyOrgs(ctx)
			if err != nil {
				return "", err
			}

			var workspaces []api.Workspace
			for _, org := range orgs {
				w, _, err := beaker.ListWorkspaces(ctx, org.Name, &client.ListWorkspaceOptions{
					Author:    user.Name,
					SortBy:    api.WorkspaceModified,
					SortOrder: api.SortDescending,
					Limit:     10,
				})
				if err != nil {
					return "", err
				}
				workspaces = append(workspaces, w...)
			}
			if len(workspaces) > 0 {
				fmt.Println("No default workspace is configured. Please select one of your workspaces:")
				for i, workspace := range workspaces {
					fmt.Printf(" %3d. %s (%d experiments)\n", i, workspace.FullName, workspace.Size.Experiments)
				}

				fmt.Println("Enter a number from the list above or the workspace name e.g. ai2/my-workspace.")
				fmt.Println("The workspace will be created if it does not exist yet.")
			} else {
				fmt.Println("No default workspace is configured. You have no existing workspaces.")
				fmt.Println("Enter the name of a new workspace e.g. ai2/my-workspace.")
			}
			workspaceRef = prompt("Default workspace")
			if i, err := strconv.Atoi(workspaceRef); err == nil {
				if i < 0 || i >= len(workspaces) {
					return "", fmt.Errorf("list index out of range: %d", i)
				}
				workspaceRef = workspaces[i].FullName
			}
			updateConfig = true
		} else {
			workspaceRef = beakerConfig.DefaultWorkspace
			if !quiet {
				fmt.Fprintf(os.Stderr, "Defaulting to workspace %s\n", color.BlueString(workspaceRef))
			}
		}
	}

	// Create the workspace if it doesn't exist.
	if _, err := beaker.Workspace(workspaceRef).Get(ctx); err != nil {
		if apiErr, ok := err.(api.Error); ok && apiErr.Code == http.StatusNotFound {
			parts := strings.Split(workspaceRef, "/")
			if len(parts) != 2 {
				return "", errors.New("workspace must be formatted like '<account>/<name>'")
			}

			if !quiet {
				fmt.Printf("Creating workspace %s\n", color.BlueString(workspaceRef))
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

	if !updateConfig {
		return workspaceRef, nil
	}
	if !quiet {
		fmt.Printf("Setting default workspace to %s\n", color.BlueString(workspaceRef))
	}
	beakerConfig.DefaultWorkspace = workspaceRef
	if err := config.WriteConfig(beakerConfig, config.GetFilePath()); err != nil {
		return "", nil
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
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		input = strings.TrimSpace(input)
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
	return false, scanner.Err()
}

// Prompt the user for input.
func prompt(prompt string) string {
	fmt.Print(prompt, ": ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := scanner.Text()
	input = strings.TrimSpace(input)
	return input
}
