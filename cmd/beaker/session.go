package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/user"
	"strconv"
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
	cmd.AddCommand(newSessionExecCommand())
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
			container, err := findRunningContainer(args[0])
			if err != nil {
				return err
			}
			return handleAttachErr(container.(*docker.Container).Attach(ctx))
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
			if node, err = getCurrentNode(); err != nil {
				return fmt.Errorf("failed to detect node; use --node flag: %w", err)
			}
		}

		session, err := beaker.CreateSession(ctx, api.SessionSpec{
			Name: name,
			Node: node,
			Requests: &api.TaskResources{
				GPUCount: gpus,
			},
		})
		if err != nil {
			return err
		}

		sessionID := session.ID
		cancel := func() {
			// If we fail to start the session, cancel it so that the executor
			// can immediately reclaim the resources allocated to it.
			//
			// Use context.Background() since ctx may already be canceled.
			_, _ = beaker.Session(sessionID).Patch(context.Background(), api.SessionPatch{
				State: &api.ExecStatusUpdate{Canceled: true},
			})
		}

		if !quiet {
			fmt.Print("Scheduling session")
			if req := resourceRequestString(session.Requests); req != "" {
				fmt.Print(" with at least ", req)
			}
			fmt.Println("... (Press Ctrl+C to cancel)")
		}

		if session, err = awaitSessionSchedule(*session); err != nil {
			cancel()
			return err
		}

		if lim := resourceLimitString(session.Limits); !quiet && lim != "" {
			fmt.Println("Reserved", lim)
		}

		// Pass nil instead of empty slice when there are no arguments.
		var command []string
		if len(args) > 0 {
			command = args
		}

		if err := startSession(*session, image, command); err != nil {
			cancel()
			return err
		}
		return nil
	}
	return cmd
}

func resourceRequestString(req *api.TaskResources) string {
	if req == nil {
		return ""
	}
	return resourceString(req.GPUCount, req.CPUCount, req.Memory)
}

func resourceLimitString(limits *api.SessionResources) string {
	if limits == nil {
		return ""
	}
	return resourceString(len(limits.GPUs), limits.CPUCount, limits.Memory)
}

func resourceString(gpuCount int, cpuCount float64, memory *bytefmt.Size) string {
	var requests []string
	if gpuCount == 1 {
		requests = append(requests, "1 GPU")
	} else if gpuCount != 0 {
		requests = append(requests, fmt.Sprintf("%d GPUs", gpuCount))
	}

	if cpuCount == 1 {
		requests = append(requests, "1 CPU")
	} else if cpuCount > 0 {
		// Format with FormatFloat instead of Printf so we can use -1 precision.
		requests = append(requests, strconv.FormatFloat(cpuCount, 'f', -1, 64)+" CPUs")
	}

	if memory != nil {
		requests = append(requests, memory.String()+" memory")
	}

	return strings.Join(requests, ", ")
}

func newSessionExecCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "exec <session> <command> <args...>",
		Short: "Execute a command in a session",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			container, err := findRunningContainer(args[0])
			if err != nil {
				return err
			}

			// Pass nil instead of empty slice when there are no arguments.
			var command []string
			if len(args) > 1 {
				command = args[1:]
			}

			err = container.(*docker.Container).Exec(ctx, &docker.ExecOpts{
				Command: command,
			})
			return handleAttachErr(err)
		},
	}
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

	var all bool
	var cluster string
	var node string
	var finalized bool
	cmd.Flags().BoolVar(&all, "all", false, "List all sessions.")
	cmd.Flags().StringVar(&cluster, "cluster", "", "Cluster to list sessions.")
	cmd.Flags().StringVar(&node, "node", "", "Node to list sessions. Defaults to current node.")
	cmd.Flags().BoolVar(&finalized, "finalized", false, "Show only finalized sessions")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		var opts client.ListSessionOpts
		if !all {
			opts.Finalized = &finalized

			if cluster != "" {
				opts.Cluster = &cluster
			}

			if !cmd.Flag("node").Changed && cluster == "" {
				var err error
				if node, err = getCurrentNode(); err != nil {
					return fmt.Errorf("failed to detect node; use --node flag: %w", err)
				}
			}
			if node != "" {
				opts.Node = &node
			}
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
			State: &api.ExecStatusUpdate{Canceled: cancel},
		}

		session, err := beaker.Session(args[0]).Patch(ctx, patch)
		if err != nil {
			return err
		}
		return printSessions([]api.Session{*session})
	}
	return cmd
}

func awaitSessionSchedule(session api.Session) (*api.Session, error) {
	s := beaker.Session(session.ID)
	cl := beaker.Cluster(session.Cluster)

	nodes, err := cl.ListClusterNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("couldn't list cluster nodes: %w", err)
	}

	nodesByID := make(map[string]*api.Node, len(nodes))
	for _, node := range nodes {
		node := node
		nodesByID[node.ID] = &node
	}

	execs, err := cl.ListExecutions(ctx, &client.ExecutionFilters{
		Scheduled: api.BoolPtr(true),
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't get cluster job queue: %w", err)
	}

	// Calculate the available capacity of each node in the cluster.
	for _, exec := range execs {
		node, ok := nodesByID[exec.Node]
		if !ok || node.Limits == nil {
			continue
		}

		node.Limits.CPUCount -= exec.Limits.CPUCount
		node.Limits.GPUCount -= exec.Limits.GPUCount
		if node.Limits.Memory != nil && exec.Limits.Memory != nil {
			node.Limits.Memory.Sub(*exec.Limits.Memory)
		}
	}

	var capacityErr string
	if node, ok := nodesByID[session.Node]; !ok {
		capacityErr = "the node has been deleted"
	} else if err := checkNodeCapacity(node, session.Requests); err != nil {
		capacityErr = err.Error()
	}

	if capacityErr != "" {
		// Find all nodes which could schedule this session.
		var hosts []string
		for _, node := range nodesByID {
			// Don't bother checking this node again.
			if node.ID == session.Node {
				continue
			}

			// Skip nodes where the session won't fit.
			if checkNodeCapacity(node, session.Requests) != nil {
				continue
			}

			hosts = append(hosts, node.Hostname)
		}

		fmt.Printf("This session is unlikely to to start because %s.\n", capacityErr)
		fmt.Println("You may continue waiting to hold your place in the queue.")
		if len(hosts) == 0 {
			fmt.Println("There are no other nodes on this cluster with sufficient capacity.")
		} else {
			fmt.Println("You could also try one of the following available nodes:")
			fmt.Println("    " + strings.Join(hosts, "\n    "))
		}
	}

	delay := time.NewTimer(0) // When to poll session status.
	for attempt := 0; ; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		case <-delay.C:
			session, err := s.Get(ctx)
			if err != nil {
				return nil, err
			}

			if session.State.Scheduled != nil {
				return session, nil
			}

			delay.Reset(3 * time.Second)
		}
	}
}

func checkNodeCapacity(node *api.Node, request *api.TaskResources) error {
	switch {
	case node.Limits == nil:
		// Node has unbounded capacity.
		return nil

	case node.Cordoned != nil:
		return errors.New("the node is cordoned")

	case request == nil:
		// No request means it'll fit anywhere.
		return nil

	case node.Limits.CPUCount < request.CPUCount:
		return errors.New("there are not enough available CPUs")

	case node.Limits.GPUCount < request.GPUCount:
		return errors.New("there are not enough available GPUs")

	case node.Limits.Memory != nil && request.Memory != nil &&
		node.Limits.Memory.Cmp(*request.Memory) < 0:
		return errors.New("there is not enough available memory")

	default:
		return nil // All checks passed.
	}
}

func startSession(session api.Session, image string, command []string) error {
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
	return handleAttachErr(container.(*docker.Container).Attach(ctx))
}

func handleAttachErr(err error) error {
	if err != nil && strings.HasPrefix(err.Error(), "exited with code ") {
		// Ignore errors coming from the container.
		// If the user exits using Ctrl-C, attach will return an error like:
		// "exited with code 130".
		return nil
	}
	return err
}

func findRunningContainer(session string) (runtime.Container, error) {
	info, err := beaker.Session(session).Get(ctx)
	if err != nil {
		return nil, err
	}
	if info.State.Started == nil {
		return nil, fmt.Errorf("session not started")
	}
	if info.State.Exited != nil || info.State.Failed != nil {
		return nil, fmt.Errorf("session already ended")
	}
	if info.State.Finalized != nil {
		return nil, fmt.Errorf("session already finalized")
	}

	rt, err := docker.NewRuntime()
	if err != nil {
		return nil, err
	}

	containers, err := rt.ListContainers(ctx)
	if err != nil {
		return nil, err
	}

	var container runtime.Container
	for _, c := range containers {
		info, err := c.Info(ctx)
		if err != nil {
			return nil, err
		}

		if session == info.Labels[sessionContainerLabel] {
			container = c
			break
		}
	}
	if container == nil {
		return nil, fmt.Errorf("container not found")
	}
	return container, nil
}
