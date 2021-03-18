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
	cmd.AddCommand(newSessionCreateCommand())
	cmd.AddCommand(newSessionGetCommand())
	cmd.AddCommand(newSessionListCommand())
	cmd.AddCommand(newSessionUpdateCommand())
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

		if err := startSession(info); err != nil {
			// If we fail to create and attach to the container, send the executor
			// a cancellation signal so that it can immediately clean up after the session
			// and reclaim the resources allocated to it.
			_, _ = beaker.Session(session.ID).Patch(ctx, api.SessionPatch{
				State: &api.ExecutionState{
					Canceled: now(),
				},
			})
			return err
		}
		return nil
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
			patch.State.Canceled = now()
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
	image := &runtime.DockerImage{
		Tag: "allenai/base:cuda11.2-ubuntu20.04",
	}

	labels := map[string]string{
		sessionContainerLabel: session.ID,
		sessionGPULabel:       strings.Join(session.Limits.GPUs, ","),
	}

	mounts := []runtime.Mount{
		// These mounts are for system accounts that are not handled by LDAP.
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
		Mounts:      mounts,
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

	err = container.(*docker.Container).Attach(ctx)
	if err != nil && strings.HasPrefix(err.Error(), "exited with code ") {
		// Ignore errors coming from the container.
		// If the user exits using Ctrl-C, attach will return an error like:
		// "exited with code 130".
		return nil
	}
	return err
}

func now() *time.Time {
	now := time.Now()
	return &now
}
