package main

import (
	"fmt"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/allenai/bytefmt"
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
		Short: "Attach to a running session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			info, err := beaker.Session(args[0]).Get(ctx)
			if err != nil {
				return err
			}
			if info.State.Started == nil {
				return fmt.Errorf("session not started")
			}
			if info.State.Ended != nil {
				return fmt.Errorf("session already ended")
			}
			if info.State.Finalized != nil {
				return fmt.Errorf("session already finalized")
			}

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

				if args[0] == info.Labels[sessionContainerLabel] {
					container = c
					break
				}
			}
			if container == nil {
				return fmt.Errorf("container not found")
			}
			return attach(container)
		},
	}
}

func newSessionCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <command...>",
		Short: "Create a new interactive session",
		Long: `Create a new interactive session backed by a Docker container.

Arguments are passed to the Docker container as a command.
To pass flags, use "--" e.g. "create -- ls -l"`,
		Args: cobra.ArbitraryArgs,
	}

	var gpus int
	var image string
	var name string
	var node string
	cmd.Flags().IntVar(&gpus, "gpus", 0, "Number of GPUs assigned to the session")
	cmd.Flags().StringVar(
		&image,
		"image",
		"allenai/base:cuda11.2-ubuntu20.04",
		"Docker image for the session.")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Assign a name to the session")
	cmd.Flags().StringVar(&node, "node", "", "Node that the session will run on. Defaults to current node.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if node == "" {
			var err error
			node, err = getCurrentNode()
			if err != nil {
				return fmt.Errorf("failed to detect node; use --node flag: %w", err)
			}
		}

		session, err := beaker.CreateSession(ctx, api.SessionSpec{
			Name: name,
			Node: node,
			Resources: &api.TaskResources{
				GPUCount: gpus,
			},
		})
		if err != nil {
			return err
		}

		info, err := awaitSessionSchedule(session.ID)
		if err != nil {
			return err
		}

		if !quiet && info.Limits != nil {
			fmt.Printf(
				"Session assigned %d GPU, %v CPU, %.1fGiB memory\n",
				len(info.Limits.GPUs),
				info.Limits.CPUCount,
				// TODO Use friendly formatting from bytefmt when available.
				float64(info.Limits.Memory.Int64())/float64(bytefmt.GiB))
		}

		// Pass nil instead of empty slice when there are no arguments.
		var command []string
		if len(args) > 0 {
			command = args
		}
		if err := startSession(info, image, command); err != nil {
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
	if !quiet {
		fmt.Printf("Waiting for session to be scheduled")
	}
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
				if !quiet {
					fmt.Println()
				}
				return info, nil
			}
			if !quiet {
				fmt.Print(".")
			}
			delay.Reset(time.Second)
		}
	}
}

func startSession(
	session *api.Session,
	image string,
	command []string,
) error {
	labels := map[string]string{
		sessionContainerLabel: session.ID,
		sessionGPULabel:       strings.Join(session.Limits.GPUs, ","),
	}

	u, err := user.Current()
	if err != nil {
		return err
	}

	env := make(map[string]string)
	var mounts []runtime.Mount
	if u.HomeDir != "" {
		env["HOME"] = u.HomeDir
		mounts = append(mounts, runtime.Mount{
			HostPath:      u.HomeDir,
			ContainerPath: u.HomeDir,
		})
	}
	if _, err := os.Stat("/net"); !os.IsNotExist(err) {
		// Mount in /net for NFS.
		mounts = append(mounts, runtime.Mount{
			HostPath:      "/net",
			ContainerPath: "/net",
		})
	}

	opts := &runtime.ContainerOpts{
		Name: strings.ToLower("session-" + session.ID),
		Image: &runtime.DockerImage{
			Tag: image,
		},
		Command:     command,
		Labels:      labels,
		Env:         env,
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

	if !quiet {
		fmt.Println("Pulling image...")
	}
	if err := rt.PullImage(ctx, opts.Image, quiet); err != nil {
		return err
	}

	container, err := rt.CreateContainer(ctx, opts)
	if err != nil {
		return err
	}

	if err := container.Start(ctx); err != nil {
		return err
	}
	return attach(container)
}

func attach(container runtime.Container) error {
	err := container.(*docker.Container).Attach(ctx)
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
