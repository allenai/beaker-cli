package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/allenai/bytefmt"
	"github.com/beaker/client/api"
	"github.com/beaker/client/client"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newClusterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster <command>",
		Short: "Manage clusters",
	}
	cmd.AddCommand(newClusterCreateCommand())
	cmd.AddCommand(newClusterDeleteCommand())
	cmd.AddCommand(newClusterGetCommand())
	cmd.AddCommand(newClusterListCommand())
	cmd.AddCommand(newClusterNodesCommand())
	cmd.AddCommand(newClusterUpdateCommand())
	return cmd
}

func newClusterCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <type>",
		Short: "Create a new cluster",
	}
	cmd.AddCommand(newClusterCreateCloudCommand())
	cmd.AddCommand(newClusterCreateOnPremCommand())
	return cmd
}

func newClusterCreateCloudCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloud <name>",
		Short: "Create a new cloud cluster with autoscaling",
		Args:  cobra.ExactArgs(1),
	}

	var maxSize int
	var preemptible bool
	var cpuCount float64
	var gpuCount int
	var gpuType string
	var memory string

	cmd.Flags().IntVar(&maxSize, "max-size", 1, "Maximum number of nodes")
	cmd.Flags().BoolVar(&preemptible, "preemptible", false, "Enable cheaper but more volatile nodes")
	cmd.Flags().Float64Var(&cpuCount, "cpus", 0, "Minimum CPU cores per node, e.g. 7.5")
	cmd.Flags().IntVar(&gpuCount, "gpus", 0, "Number of GPUs per node: 1, 2, 4, or 8")
	cmd.Flags().StringVar(&gpuType, "gpu-type", "", "Type of GPU: a100, k80, p100, p4, t4, or v100")
	cmd.Flags().StringVar(&memory, "memory", "", "Minimum memory per node, e.g. 6.5GiB")

	// Deprecated flags are replaced by the above.
	cmd.Flags().Float64Var(&cpuCount, "cpu-count", 0, "")
	cmd.Flags().MarkDeprecated("cpu-count", "please use --cpus instead")
	cmd.Flags().IntVar(&gpuCount, "gpu-count", 0, "")
	cmd.Flags().MarkDeprecated("gpu-count", "please use --gpus instead")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		parts := strings.Split(args[0], "/")
		if len(parts) != 2 {
			return fmt.Errorf("cluster names must be fully scoped in the form %s", color.GreenString("account/cluster"))
		}
		account, clusterName := parts[0], parts[1]

		if cpuCount == 0 && gpuCount == 0 && gpuType == "" && memory == "" {
			return fmt.Errorf("cloud clusters must specify at least 1 resource")
		}

		var memorySize *bytefmt.Size
		if memory != "" {
			var err error
			if memorySize, err = bytefmt.Parse(memory); err != nil {
				return err
			}
		}
		spec := api.ClusterSpec{
			Name:        clusterName,
			Capacity:    maxSize,
			Preemptible: preemptible,
			Spec: &api.NodeResources{
				CPUCount: cpuCount,
				GPUCount: gpuCount,
				GPUType:  gpuType,
				Memory:   memorySize,
			},
		}
		cluster, err := beaker.CreateCluster(ctx, account, spec)
		if err != nil {
			return err
		}

		if !quiet {
			fmt.Printf("Cluster %s created. See details at %s/cl/%s\n",
				color.BlueString(cluster.FullName), beaker.Address(), cluster.FullName)
		}

		validated := func(ctx context.Context) (bool, error) {
			var err error
			cluster, err = beaker.Cluster(cluster.ID).Get(ctx)
			if err != nil {
				return false, err
			}
			return cluster.Status != api.ClusterPending, nil
		}
		err = await(ctx, "Validating cluster", validated, 0)
		if err != nil {
			return fmt.Errorf("await cluster validation: %w", err)
		}
		switch cluster.Status {
		case api.ClusterActive:
			return printClusters([]api.Cluster{*cluster})
		case api.ClusterFailed:
			return fmt.Errorf("cluster validation failed: %s", cluster.StatusMessage)
		default:
			return fmt.Errorf("unexpected cluster state: %s", cluster.Status)
		}
	}
	return cmd
}

func newClusterCreateOnPremCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "on-prem <name>",
		Short: "Create a new on-premise cluster without autoscaling",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			parts := strings.Split(args[0], "/")
			if len(parts) != 2 {
				return fmt.Errorf("cluster names must be fully scoped in the form %s", color.GreenString("account/cluster"))
			}
			account, name := parts[0], parts[1]
			cluster, err := beaker.CreateCluster(ctx, account, api.ClusterSpec{Name: name})
			if err != nil {
				return err
			}

			if !quiet {
				fmt.Printf("Cluster %s created. See details at %s/cl/%s\n",
					color.BlueString(cluster.FullName), beaker.Address(), cluster.FullName)
			}
			return printClusters([]api.Cluster{*cluster})
		},
	}
}

func newClusterDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <cluster>",
		Short: "Permanently remove a cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := beaker.Cluster(args[0]).Terminate(ctx); err != nil {
				return err
			}

			if !quiet {
				fmt.Printf("Deleted %s\n", color.BlueString(args[0]))
			}
			return nil
		},
	}
}

func newClusterGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "get <cluster...>",
		Aliases: []string{"inspect"},
		Short:   "Display detailed information about one or more clusters",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var clusters []api.Cluster
			for _, id := range args {
				info, err := beaker.Cluster(id).Get(ctx)
				if err != nil {
					return err
				}
				clusters = append(clusters, *info)
			}
			return printClusters(clusters)
		},
	}
}

func newClusterListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <account>",
		Short: "List clusters under an account",
		Args:  cobra.ExactArgs(1),
	}

	var cloud bool
	var onPrem bool
	cmd.Flags().BoolVar(&cloud, "cloud", false, "Only show cloud (autoscaling) clusters")
	cmd.Flags().BoolVar(&onPrem, "on-prem", false, "Only show on-premise (non-autoscaling) clusters")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if cloud && onPrem {
			return fmt.Errorf("only one of --cloud and --on-prem may be set")
		}

		var clusters []api.Cluster
		var cursor string
		for {
			var page []api.Cluster
			var err error
			page, cursor, err = beaker.ListClusters(ctx, args[0], &client.ListClusterOptions{
				Cursor: cursor,
			})
			if err != nil {
				return err
			}

			for _, cluster := range page {
				if cloud {
					if !cluster.Autoscale {
						continue
					}
				}
				if onPrem {
					if cluster.Autoscale {
						continue
					}
				}
				clusters = append(clusters, cluster)
			}
			if cursor == "" {
				break
			}
		}
		return printClusters(clusters)
	}
	return cmd
}

func newClusterNodesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "nodes <cluster>",
		Short: "List nodes in a cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nodes, err := beaker.Cluster(args[0]).ListClusterNodes(ctx)
			if err != nil {
				return err
			}
			return printNodes(nodes)
		},
	}
}

func newClusterUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <cluster>",
		Short: "Modify a cluster",
		Args:  cobra.ExactArgs(1),
	}

	var maxSize int
	cmd.Flags().IntVar(&maxSize, "max-size", -1, "Maximum number of nodes")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		patch := api.ClusterPatch{}
		if maxSize >= 0 {
			patch.Capacity = &maxSize
		}

		if (patch == api.ClusterPatch{}) {
			fmt.Println("Nothing to update.")
			return nil
		}

		cluster, err := beaker.Cluster(args[0]).Patch(ctx, &patch)
		if err != nil {
			return err
		}

		fmt.Println(cluster.ID)
		return nil
	}
	return cmd
}
