package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/allenai/beaker/config"
	"github.com/allenai/bytefmt"
	"github.com/beaker/client/api"
	"github.com/beaker/client/client"
	"github.com/beaker/runtime"
	"github.com/beaker/runtime/docker"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const defaultImage = "beaker://ai2/cuda11.2-ubuntu20.04"

func newSessionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session <command>",
		Short: "Manage sessions",
	}
	cmd.AddCommand(newSessionAttachCommand())
	cmd.AddCommand(newSessionCreateCommand())
	cmd.AddCommand(newSessionExecCommand())
	cmd.AddCommand(newSessionGetCommand())
	cmd.AddCommand(newSessionDescribeCommand())
	cmd.AddCommand(newSessionImagesCommand())
	cmd.AddCommand(newSessionListCommand())
	cmd.AddCommand(newSessionStopCommand())
	return cmd
}

func newSessionAttachCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attach",
		Short: "Attach to a running session",
		Args:  cobra.NoArgs,
	}

	var session string
	cmd.Flags().StringVar(&session, "session", "", "Target session. Defaults to the running session.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		container, _, err := findRunningSessionContainer(session)
		if err != nil {
			return err
		}

		resp, err := container.Attach(ctx)
		if err != nil {
			return err
		}
		defer resp.Close()

		return handleAttachErr(container.Stream(ctx, resp))
	}
	return cmd
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

	var image string
	var name string
	var node string
	var workspace string
	var saveImage bool
	var noUpdateDefaultImage bool
	cmd.Flags().StringVarP(
		&image,
		"image",
		"i",
		defaultImage,
		"Base image to run, may be a Beaker or Docker image. Uses 'default_image' from the Beaker configuration if set.")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Assign a name to the session")
	cmd.Flags().StringVar(&node, "node", "", "Node that the session will run on. Defaults to current node.")
	cmd.Flags().StringVarP(&workspace, "workspace", "w", "", "Workspace where the session will be placed")
	cmd.Flags().BoolVarP(
		&saveImage,
		"save-image",
		"s",
		false,
		"Save the result image of the session. A new image will be created in the session's workspace.")
	cmd.Flags().BoolVar(
		&noUpdateDefaultImage,
		"no-update-default-image",
		false,
		"Do not update the default image when using --save-image.")

	var secretEnv map[string]string
	var secretMount map[string]string
	cmd.Flags().StringToStringVar(
		&secretEnv,
		"secret-env",
		map[string]string{},
		"Secret environment variables in the format <variable>=<secret name>")
	cmd.Flags().StringToStringVar(
		&secretMount,
		"secret-mount",
		map[string]string{},
		"Secret file mounts in the format <secret name>=<file path> e.g. SECRET=/secret")

	var cpus float64
	var gpus int
	var memory string
	var sharedMemory string
	var ports []string
	cmd.Flags().Float64Var(&cpus, "cpus", 0, "Minimum CPU cores to reserve, e.g. 7.5")
	cmd.Flags().IntVar(&gpus, "gpus", 0, "Minimum number of GPUs to reserve")
	cmd.Flags().StringVar(&memory, "memory", "", "Minimum memory to reserve, e.g. 6.5GiB")
	cmd.Flags().StringVar(&sharedMemory, "shared-memory", "", "Shared memory (size of /dev/shm), e.g. 1GiB")
	cmd.Flags().StringSliceVar(
		&ports,
		"port",
		[]string{},
		"TCP container ports to expose. Each will be assigned a random, ephemeral port on the host.",
	)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		rt, err := docker.NewRuntime()
		if err != nil {
			return fmt.Errorf("couldn't initialize container runtime: %w", err)
		}

		if node == "" {
			if node, err = getCurrentNode(); err != nil {
				return fmt.Errorf("failed to detect node; use --node flag: %w", err)
			}
		}

		var memSize *bytefmt.Size
		if memory != "" {
			if memSize, err = bytefmt.Parse(memory); err != nil {
				return fmt.Errorf("invalid value for --memory: %w", err)
			}
		}

		var sharedMemSize *bytefmt.Size
		if sharedMemory != "" {
			if sharedMemSize, err = bytefmt.Parse(sharedMemory); err != nil {
				return fmt.Errorf("invalid value for --shared-memory: %w", err)
			}
		}

		if image == defaultImage && beakerConfig.DefaultImage != "" {
			fmt.Printf("Defaulting to image %s\n", color.BlueString(beakerConfig.DefaultImage))
			image = beakerConfig.DefaultImage
		}
		imageSource, err := getImageSource(image)
		if err != nil {
			return err
		}

		if workspace, err = ensureWorkspace(workspace); err != nil {
			return err
		}

		var envVars []api.EnvironmentVariable
		for k, v := range secretEnv {
			envVars = append(envVars, api.EnvironmentVariable{
				Name:   k,
				Secret: v,
			})
		}

		var mounts []api.DataMount
		for k, v := range secretMount {
			mounts = append(mounts, api.DataMount{
				MountPath: v,
				Source: api.DataSource{
					Secret: k,
				},
			})
		}

		var tcpPorts []api.TCPPort
		for _, p := range ports {
			pp, err := strconv.ParseInt(p, 10, 32)
			if err != nil {
				return fmt.Errorf("invalid port: %d", pp)
			}
			// We know the conversion to int32 is safe because of the bitsize
			// passed to strconv.ParseInt() above.
			tcpPorts = append(tcpPorts, int32(pp))
		}

		session, err := beaker.CreateJob(ctx, api.JobSpec{
			Session: &api.SessionJobSpec{
				Workspace: workspace,
				Name:      name,
				Node:      node,
				Requests: &api.ResourceRequest{
					CPUCount:     cpus,
					GPUCount:     gpus,
					Memory:       memSize,
					SharedMemory: sharedMemSize,
				},
				Command:   args,
				EnvVars:   envVars,
				Datasets:  mounts,
				Image:     *imageSource,
				SaveImage: saveImage,
				TCPPorts:  tcpPorts,
			},
		})
		if err != nil {
			return err
		}

		verificationFile, err := os.Create(session.SessionVerificationFile())
		if err != nil {
			return fmt.Errorf("failed to create session verification file")
		}
		defer verificationFile.Close()
		defer os.Remove(verificationFile.Name())

		shouldCancel, sessionID := true, session.ID
		defer func() {
			// If we fail to start the session, cancel it so that the executor
			// can immediately reclaim the resources allocated to it.
			//
			// Use context.Background() since ctx may already be canceled.
			if !shouldCancel {
				return
			}
			_, _ = beaker.Job(sessionID).Patch(context.Background(), api.JobPatch{
				Status: &api.JobStatusUpdate{Canceled: true},
			})
		}()

		if !quiet {
			fmt.Printf("Starting session %s", color.BlueString(session.ID))
			if req := resourceRequestString(session.Requests); req != "" {
				fmt.Print(" with at least ", req)
			}
			fmt.Println("... (Press Ctrl+C to cancel)")
		}

		if session, err = awaitSessionStart(*session); err != nil {
			return err
		}

		if lim := resourceLimitString(session.Limits); !quiet && lim != "" {
			fmt.Println("Reserved", lim)
		}

		container := rt.Container(session.ContainerName()).(*docker.Container)
		resp, err := container.Attach(ctx)
		if err != nil {
			return fmt.Errorf("attach: %w", err)
		}
		defer resp.Close()

		if err := container.Start(ctx); err != nil {
			return fmt.Errorf("start: %w", err)
		}

		if saveImage && !quiet {
			fmt.Println(color.YellowString(`
WARNING: The root filesystem of this session will be saved.
Do not write sensitive information outside of the home directory.
`))
		}

		info, err := container.Info(ctx)
		if err != nil {
			return fmt.Errorf("getting container info: %w", err)
		}

		if len(info.TCPPorts) > 0 {
			bindings := []string{}
			for _, pb := range info.TCPPorts {
				bindings = append(bindings, color.BlueString(
					"0.0.0.0:%d->%d/tcp",
					pb.HostPort,
					pb.ContainerPort,
				))
			}
			fmt.Printf("Exposed Ports: %s\n", strings.Join(bindings, ", "))
		}

		if err := handleAttachErr(container.Stream(ctx, resp)); err != nil {
			return fmt.Errorf("stream: %w", err)
		}
		shouldCancel = false

		if !saveImage {
			return nil
		}
		var job *api.Job
		started := func(ctx context.Context) (bool, error) {
			var err error
			job, err = beaker.Job(session.ID).Get(ctx)
			if err != nil {
				return false, err
			}
			return job.Status.Finalized != nil, nil
		}
		if err := await(ctx, "Waiting for image capture to complete", started, 0); err != nil {
			return fmt.Errorf("waiting for image capture to complete: %w", err)
		}
		if job.Status.Failed != nil {
			return fmt.Errorf("session failed: %s", job.Status.Message)
		}
		images, err := beaker.Job(job.ID).GetImages(ctx)
		if err != nil {
			return err
		}
		if len(images) == 0 {
			return fmt.Errorf("job has no result images")
		}
		if !quiet {
			fmt.Printf("Image saved to %s: %s/im/%s\n",
				color.BlueString(images[0].ID),
				beaker.Address(),
				images[0].ID)
		}
		if noUpdateDefaultImage {
			if !quiet {
				fmt.Printf(`Default image not updated.
Resume this session with: beaker session create --image beaker://%s
`, images[0].ID)
			}
			return nil
		}
		beakerConfig.DefaultImage = "beaker://" + images[0].ID
		if err := config.WriteConfig(beakerConfig, config.GetFilePath()); err != nil {
			return fmt.Errorf("setting default image: %w", err)
		}
		if !quiet {
			fmt.Printf(`Default image updated in your config file: %s
Resume this session with: beaker session create
`, config.GetFilePath())
		}
		return nil
	}
	return cmd
}

func resourceRequestString(req *api.ResourceRequest) string {
	if req == nil {
		return ""
	}
	return resourceString(req.GPUCount, req.CPUCount, req.Memory)
}

func resourceLimitString(limits *api.ResourceLimits) string {
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
		requests = append(requests, fmt.Sprintf("%v memory", memory))
	}

	return strings.Join(requests, ", ")
}

func getImageSource(name string) (*api.ImageSource, error) {
	parts := strings.SplitN(name, "://", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("image must include scheme such as beaker:// or docker://")
	}
	scheme, image := parts[0], parts[1]

	switch strings.ToLower(scheme) {
	case "beaker":
		return &api.ImageSource{Beaker: image}, nil

	case "docker":
		return &api.ImageSource{Docker: image}, nil

	default:
		return nil, fmt.Errorf("%q is not a supported image type", scheme)
	}
}

func newSessionExecCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec [command...]",
		Short: "Execute a command in a session",
		Long: `Execute a command in a session

If no command is provided, exec will run 'bash -l'`,
	}

	var session string
	cmd.Flags().StringVar(&session, "session", "", "Target session. Defaults to the running session.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		container, _, err := findRunningSessionContainer(session)
		if err != nil {
			return err
		}

		command := []string{"bash", "-l"}
		if len(args) > 0 {
			command = args
		}

		err = container.Exec(ctx, &docker.ExecOpts{
			Command: command,
		})
		return handleAttachErr(err)
	}
	return cmd
}

func newSessionGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <session...>",
		Short: "Display basic information about one or more sessions",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var jobs []api.Job
			for _, id := range args {
				info, err := beaker.Job(id).Get(ctx)
				if err != nil {
					return err
				}
				jobs = append(jobs, *info)
			}
			return printJobs(jobs)
		},
	}
}

// These structs are local to the client because they represent the combination
// of details the API knows about a session with those derived from the container
// runtime. This is the information that's output via the `describe` command.
type tcpPortBinding struct {
	HostPort      runtime.TCPPort `json:"host_port"`
	ContainerPort runtime.TCPPort `json:"container_port"`
}
type runtimeInfo struct {
	TCPPorts []tcpPortBinding `json:"tcp_ports"`
}
type sessionDetails struct {
	*api.Job
	Runtime *runtimeInfo `json:"runtime"`
}

func newDetails(session *api.Job, info *runtime.ContainerInfo) *sessionDetails {
	ports := []tcpPortBinding{}
	for _, pb := range info.TCPPorts {
		ports = append(ports, tcpPortBinding{
			HostPort:      pb.HostPort,
			ContainerPort: pb.ContainerPort,
		})
	}

	ri := runtimeInfo{TCPPorts: ports}
	return &sessionDetails{
		Job:     session,
		Runtime: &ri,
	}
}

func newSessionDescribeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "describe <session...>",
		Short: "Display detailed information about a single session",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := ""
			if len(args) == 1 {
				ref = args[0]
			}
			container, session, err := findRunningSessionContainer(ref)
			if err != nil {
				return err
			}

			info, err := container.Info(ctx)
			if err != nil {
				return err
			}

			node, err := beaker.Node(session.Node).Get(ctx)
			if err != nil {
				return err
			}

			details := newDetails(session, info)

			switch format {
			case formatJSON:
				printJSON(details)
			default:
				printTableRow("ID", details.ID)
				printTableRow("Name", details.Name)
				printTableRow(
					"User",
					fmt.Sprintf(
						"%s (%s)",
						details.Author.DisplayName,
						details.Author.Name,
					),
				)

				var img string
				if details.Session.Image.Beaker != "" {
					img = fmt.Sprintf("beaker://%s", details.Session.Image.Beaker)
				} else if details.Session.Image.Docker != "" {
					img = fmt.Sprintf("docker://%s", details.Session.Image.Docker)
				}
				printTableRow("Image", img)

				// Print some extra information for Beaker images, since we have it.
				if details.Session.Image.Beaker != "" {
					bkrImg, err := beaker.Image(details.Session.Image.Beaker).Get(ctx)
					if err != nil {
						return err
					}
					url := fmt.Sprintf("%s/im/%s", beaker.Address(), bkrImg.ID)
					// HACK printTableRow prints "N/A" instead of an empty string,
					// which we don't want, so we pass a single space instead.
					printTableRow(" ", url)
				}

				var start time.Time
				if details.Status.Scheduled != nil {
					start = *details.Status.Scheduled
				}
				printTableRow("Started", start)
				printTableRow("Elapsed", jobDuration(*details.Job))
				printTableRow("Status", jobStatus(details.Job.Status))

				var gpus string
				if details.Limits != nil {
					gpus = strconv.Itoa(len(details.Limits.GPUs))
				}
				printTableRow("GPUS", gpus)

				for i, pb := range details.Runtime.TCPPorts {
					// HACK printTableRow prints "N/A" instead of an empty string,
					// which we don't want, so we pass a single space instead.
					title := " "
					if i == 0 {
						title = "TCP Ports"
					}
					url := fmt.Sprintf("http://%s:%d", node.Hostname, pb.HostPort)
					p := fmt.Sprintf(
						"%s:%d->%d/tcp (%s)",
						node.Hostname,
						pb.HostPort,
						pb.ContainerPort,
						url,
					)
					printTableRow(title, p)
				}
			}

			return nil
		},
	}
}

func newSessionImagesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "images <session>",
		Short: "List result images of a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			images, err := beaker.Job(args[0]).GetImages(ctx)
			if err != nil {
				return err
			}
			return printImages(images)
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
		kind := api.JobKindSession
		opts := client.ListJobOpts{Kind: &kind}
		if !all {
			opts.Finalized = &finalized
			opts.Cluster = cluster

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

		jobs, err := listJobs(opts)
		if err != nil {
			return err
		}
		return printJobs(jobs)
	}
	return cmd
}

func newSessionStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop a pending or running session",
		Args:  cobra.NoArgs,
	}

	var session string
	cmd.Flags().StringVar(&session, "session", "", "Target session. Defaults to the running session.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if session == "" {
			info, err := findRunningSession()
			if err != nil {
				return err
			}
			session = info.ID
		}

		job, err := beaker.Job(session).Patch(ctx, api.JobPatch{
			Status: &api.JobStatusUpdate{Canceled: true},
		})
		if err != nil {
			return err
		}
		return printJobs([]api.Job{*job})
	}
	return cmd
}

func awaitSessionStart(session api.Job) (*api.Job, error) {
	s := beaker.Job(session.ID)
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

	jobs, err := listJobs(client.ListJobOpts{
		Cluster:   session.Cluster,
		Finalized: api.BoolPtr(false),
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't list cluster jobs: %w", err)
	}

	// Subtract each running job from its node's capacity.
	for _, job := range jobs {
		node, ok := nodesByID[job.Node]
		if !ok || node.Limits == nil {
			continue
		}

		// Ignore jobs which haven't fully scheduled yet, including the one we're starting.
		if job.ID == session.ID || job.Limits == nil {
			continue
		}

		node.Limits.CPUCount -= job.Limits.CPUCount
		node.Limits.GPUCount -= len(job.Limits.GPUs)
		if node.Limits.Memory != nil && job.Limits.Memory != nil {
			node.Limits.Memory.Sub(*job.Limits.Memory)
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

		if !quiet {
			fmt.Printf("This session is unlikely to to start because %s.\n", capacityErr)
			fmt.Println("You may continue waiting to hold your place in the queue.")
			if len(hosts) == 0 {
				fmt.Println("There are no other nodes on this cluster with sufficient capacity.")
			} else {
				fmt.Println("You could also try one of the following available nodes:")
				fmt.Println("    " + strings.Join(hosts, "\n    "))
			}
			fmt.Println()
		}
	}

	var job *api.Job
	started := func(ctx context.Context) (bool, error) {
		var err error
		job, err = s.Get(ctx)
		if err != nil {
			return false, err
		}
		if job.Status.Finalized != nil {
			return false, fmt.Errorf("session finalized: %s", job.Status.Message)
		}
		return job.Status.Started != nil, nil
	}
	return job, await(ctx, "Waiting for session to start", started, 0)
}

func checkNodeCapacity(node *api.Node, request *api.ResourceRequest) error {
	switch {
	case node.Limits == nil:
		// Node has unknown capacity. Treat it as unbounded.
		return nil

	case node.Cordoned != nil:
		return errors.New("the node is cordoned")

	case request.IsEmpty():
		// No request means it'll fit anywhere.
		return nil

	case node.Limits.CPUCount < request.CPUCount:
		return errors.New("there are not enough available CPUs")

	case node.Limits.GPUCount < request.GPUCount:
		return errors.New("there are not enough available GPUs")

	case node.Limits.Memory != nil && request.Memory != nil &&
		node.Limits.Memory.Cmp(*request.Memory) < 0:
		return errors.New("there is not enough available memory")

	case node.Limits.CPUCount == 0 &&
		node.Limits.GPUCount == 0 &&
		(node.Limits.Memory == nil || node.Limits.Memory.IsZero()):
		return errors.New("the node has no space left")

	default:
		return nil // All checks passed.
	}
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

// Find the container of the given session or the running session if no
// session reference is provided. Returns an error if there is not exactly
// one running session or if the container is not in a running state.
func findRunningSessionContainer(ref string) (*docker.Container, *api.Job, error) {
	var session *api.Job
	var err error
	if ref == "" {
		session, err = findRunningSession()
	} else {
		session, err = beaker.Job(ref).Get(ctx)
	}
	if err != nil {
		return nil, nil, err
	}
	c, err := findRunningContainer(*session)
	if err != nil {
		return nil, session, err
	}
	return c, session, nil
}

// Find a running session. Returns an error if there are no running sessions
// or if there are multiple running sessions.
func findRunningSession() (*api.Job, error) {
	node, err := getCurrentNode()
	if err != nil {
		return nil, fmt.Errorf("failed to detect current node, ensure executor is running: %w", err)
	}
	kind := api.JobKindSession
	sessions, err := beaker.ListJobs(ctx, &client.ListJobOpts{
		Kind:      &kind,
		Node:      &node,
		Scheduled: api.BoolPtr(true),
		Finalized: api.BoolPtr(false),
	})
	if err != nil {
		return nil, fmt.Errorf("listing sessions: %w", err)
	}
	user, err := beaker.WhoAmI(ctx)
	if err != nil {
		return nil, fmt.Errorf("whoami: %w", err)
	}
	// The list jobs API does not support filtering by user so we filter client-side.
	// TODO: https://github.com/allenai/beaker-service/issues/1872
	var userSessions []api.Job
	for _, session := range sessions.Data {
		if session.Author.ID != user.Identity.ID {
			continue
		}
		userSessions = append(userSessions, session)
	}
	if len(userSessions) == 0 {
		return nil, fmt.Errorf("no running sessions found")
	}
	if len(userSessions) > 1 {
		if !quiet {
			if err := printJobs(userSessions); err != nil {
				return nil, err
			}
		}
		return nil, fmt.Errorf("multiple running sessions found, select one with --session")
	}
	session := userSessions[0]
	if !quiet {
		fmt.Printf("Found running session: %s\n", color.BlueString(session.ID))
	}
	return &session, nil
}

// Find a running container for a session.
func findRunningContainer(session api.Job) (*docker.Container, error) {
	if session.Status.Started == nil {
		return nil, fmt.Errorf("session not started")
	}
	if session.Status.Exited != nil || session.Status.Failed != nil {
		return nil, fmt.Errorf("session already ended")
	}
	if session.Status.Finalized != nil {
		return nil, fmt.Errorf("session already finalized")
	}

	rt, err := docker.NewRuntime()
	if err != nil {
		return nil, err
	}
	return rt.Container(session.ContainerName()).(*docker.Container), nil
}
