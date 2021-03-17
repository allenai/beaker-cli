package main

import (
	"fmt"
	"os/user"
	"strings"
	"time"

	"github.com/beaker/client/api"
	"github.com/beaker/client/client"
	"github.com/beaker/runtime"
	"github.com/beaker/runtime/docker"
	"github.com/spf13/cobra"
)

const (
	// Label containing the session ID on session containers.
	sessionContainerLabel = "beaker.org/session"

	// Label containing a list of the GPUs assigned to the container e.g. "1,2".
	sessionGPULabel = "beaker.org/gpus"
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
	cmd.AddCommand(newSessionUpdateCommand())
	return cmd
}

func newSessionAttachCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "attach <session>",
		Short: "Attach to a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO What if the session is created but not started?
			// How can we recover in that case?
			return attachSession(args[0])
		},
	}
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
	cmd.Flags().StringVar(&node, "node", "", "Node that the session will run on. Defaults to current node.")

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

		info, err := awaitSessionSchedule(session.ID)
		if err != nil {
			return err
		}

		return startSession(info)
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
		Short: "List sessions",
		Args:  cobra.NoArgs,
	}

	var cluster string
	var node string
	var finalized bool
	cmd.Flags().StringVar(&cluster, "cluster", "", "Cluster to list sessions.")
	cmd.Flags().StringVar(&node, "node", "", "Node to list sessions. Defaults to current node.")
	cmd.Flags().BoolVar(&finalized, "finalized", false, "Show only finalized sessions")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		opts := client.ListSessionOpts{
			Finalized: &finalized,
		}

		if cluster != "" {
			opts.Cluster = &cluster
		}

		if !cmd.Flag("node").Changed && cluster == "" {
			var err error
			node, err = getCurrentNode()
			if err != nil {
				return fmt.Errorf("failed to detect node; use --node flag: %w", err)
			}
		}
		if node != "" {
			opts.Node = &node
		}

		sessions, err := beaker.ListSessions(ctx, &opts)
		if err != nil {
			return err
		}
		return printSessions(sessions)
	}
	return cmd
}

func newSessionUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a session",
		Args:  cobra.ExactArgs(1),
	}

	var cancel bool
	cmd.Flags().BoolVar(&cancel, "cancel", false, "Cancel a session")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		patch := api.SessionPatch{
			State: &api.ExecutionState{},
		}
		if cancel {
			now := time.Now()
			patch.State.Canceled = &now
		}

		session, err := beaker.Session(args[0]).Patch(ctx, patch)
		if err != nil {
			return err
		}
		return printSessions([]api.Session{*session})
	}
	return cmd
}

func awaitSessionSchedule(session string) (*api.Session, error) {
	fmt.Printf("Waiting for session to be scheduled")
	delay := time.NewTimer(0) // No delay on first attempt.
	for attempt := 0; ; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		case <-delay.C:
			info, err := beaker.Session(session).Get(ctx)
			if err != nil {
				return nil, err
			}

			if info.State.Scheduled != nil {
				fmt.Println()
				return info, nil
			}
			fmt.Print(".")
			delay.Reset(time.Second)
		}
	}
}

func startSession(session *api.Session) error {
	// TODO This image is a placeholder; replace it with the interactive base image.
	image := &runtime.DockerImage{
		Tag: "ubuntu:20.04",
	}

	labels := map[string]string{
		sessionContainerLabel: session.ID,
		sessionGPULabel:       strings.Join(session.Limits.GPUs, ","),
	}

	mounts := []runtime.Mount{
		// All of these mounts are for LDAP.
		{
			HostPath:      "/etc/group",
			ContainerPath: "/etc/group",
			ReadOnly:      true,
		},
		{
			HostPath:      "/etc/passwd",
			ContainerPath: "/etc/passwd",
			ReadOnly:      true,
		},
		{
			HostPath:      "/etc/shadow",
			ContainerPath: "/etc/shadow",
			ReadOnly:      true,
		},
	}

	u, err := user.Current()
	if err != nil {
		return err
	}
	if u.HomeDir != "" {
		mounts = append(mounts, runtime.Mount{
			HostPath:      u.HomeDir,
			ContainerPath: u.HomeDir,
		})
	}

	opts := &runtime.ContainerOpts{
		Name:        strings.ToLower("session-" + session.ID),
		Image:       image,
		Command:     []string{"bash"},
		Labels:      labels,
		CPUCount:    session.Limits.CPUCount,
		GPUs:        session.Limits.GPUs,
		Memory:      session.Limits.Memory.Int64(),
		Interactive: true,
		User:        u.Uid + ":" + u.Gid,
		WorkingDir:  u.HomeDir,
	}

	rt, err := docker.NewRuntime()
	if err != nil {
		return err
	}

	container, err := rt.CreateContainer(ctx, opts)
	if err != nil {
		return err
	}

	return container.(*docker.Container).Attach(ctx)
}

func attachSession(session string) error {
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

		if session == info.Labels[sessionContainerLabel] {
			container = c
			break
		}
	}
	if container == nil {
		return fmt.Errorf("container not found")
	}

	if err := container.(*docker.Container).Attach(ctx); err != nil {
		return fmt.Errorf("couldn't attach to container: %w", err)
	}
	return nil
}
