package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/allenai/beaker-service/runtime"
	"github.com/allenai/beaker-service/runtime/docker"
	"github.com/beaker/client/api"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func newSessionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session <command>",
		Short: "Manage sessions",
	}
	cmd.AddCommand(newSessionAttachCommand())
	cmd.AddCommand(newSessionCreateCommand())
	cmd.AddCommand(newSessionGetCommand())
	cmd.AddCommand(newSessionListCommand())
	return cmd
}

func newSessionAttachCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attach <session>",
		Short: "Attach to an interactive session",
		Args:  cobra.ExactArgs(1),
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		rt, err := docker.NewRuntime()
		if err != nil {
			return err
		}

		containers, err := rt.ListContainers(ctx)
		if err != nil {
			return err
		}

		var container runtime.Container
		for _, c := range containers {
			info, err := c.Info(ctx)
			if err != nil {
				return err
			}

			if args[0] == info.Labels["beaker.org/session"] {
				c := c
				container = c
				break
			}
		}
		if container == nil {
			return fmt.Errorf("container not found")
		}

		if err := container.Start(ctx); err != nil {
			return fmt.Errorf("couldn't start container: %w", err)
		}

		if err := container.(*docker.Container).Attach(ctx); err != nil {
			return fmt.Errorf("couldn't attach to container: %w", err)
		}
		return nil
	}
	return cmd
}

func newSessionCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new interactive session on a node",
		Args:  cobra.NoArgs,
	}

	var name string
	var node string
	cmd.Flags().StringVarP(&name, "name", "n", "", "Assign a name to the session")
	cmd.Flags().StringVar(&node, "node", "", "Node that the session will run on")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if node == "" {
			var err error
			node, err = getCurrentNode()
			if err != nil {
				return fmt.Errorf("failed to detect node; use --node flag: %w", err)
			}
			fmt.Printf("Detected node: %q\n", node)
		}

		session, err := beaker.CreateSession(ctx, api.SessionSpec{
			Name: name,
			Node: node,
		})
		if err != nil {
			return err
		}
		return printSessions([]api.Session{*session})
	}
	return cmd
}

func newSessionGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "get <session...>",
		Aliases: []string{"inspect"},
		Short:   "Display detailed information about one or more sessions",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessions []api.Session
			for _, id := range args {
				info, err := beaker.Session(id).Get(ctx)
				if err != nil {
					return err
				}
				sessions = append(sessions, *info)
			}
			return printSessions(sessions)
		},
	}
}

func newSessionListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sessions on a node",
		Args:  cobra.NoArgs,
	}

	var node string
	cmd.Flags().StringVar(&node, "node", "", "Node that the session will run on")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if node == "" {
			var err error
			node, err = getCurrentNode()
			if err != nil {
				return fmt.Errorf("failed to detect node; use --node flag: %w", err)
			}
			fmt.Printf("Detected node: %q\n", node)
		}

		sessions, err := beaker.Node(node).ListSessions(ctx)
		if err != nil {
			return err
		}
		return printSessions(sessions)
	}
	return cmd
}

type executorConfig struct {
	StoragePath string `yaml:"storagePath"`
}

// Get the node ID of the executor running on this machine, if there is one.
func getCurrentNode() (string, error) {
	configFile, err := ioutil.ReadFile("/etc/beaker/config.yml")
	if err != nil {
		return "", err
	}
	expanded := strings.NewReader(os.ExpandEnv(string(configFile)))

	var config executorConfig
	if err := yaml.NewDecoder(expanded).Decode(&config); err != nil {
		return "", err
	}

	node, err := ioutil.ReadFile(path.Join(config.StoragePath, "node"))
	if err != nil {
		return "", err
	}
	return string(node), nil
}
