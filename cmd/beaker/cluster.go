package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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
	cmd.AddCommand(newClusterExecutionsCommand())
	cmd.AddCommand(newClusterInspectCommand())
	cmd.AddCommand(newClusterListCommand())
	cmd.AddCommand(newClusterNodesCommand())
	cmd.AddCommand(newClusterTerminateCommand())
	cmd.AddCommand(newClusterUpdateCommand())
	return cmd
}

func newClusterCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new cluster",
		Args:  cobra.ExactArgs(1),
	}

	var maxSize int
	var preemptible bool
	var protected bool
	var cpuCount float64
	var gpuCount int
	var gpuType string
	var memory string

	cmd.Flags().IntVar(&maxSize, "max-size", 0, "Maximum number of nodes")
	cmd.Flags().BoolVar(&preemptible, "preemptible", false, "Enable cheaper but more volatile nodes")
	cmd.Flags().BoolVar(&protected, "protected", false, "Mark cluster as protected")
	cmd.Flags().Float64Var(&cpuCount, "cpu-count", 0, "Number of CPUs per node")
	cmd.Flags().IntVar(&gpuCount, "gpu-count", 0, "Number of GPUs per node")
	cmd.Flags().StringVar(&gpuType, "gpu-type", "", "Type of GPU, e.g. p100")
	cmd.Flags().StringVar(&memory, "memory", "", "Memory limit per node, e.g. 6.5GiB")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		parts := strings.Split(args[0], "/")
		if len(parts) != 2 {
			return fmt.Errorf("cluster names must be fully scoped in the form %s", color.GreenString("account/cluster"))
		}

		account, clusterName := parts[0], parts[1]
		spec := api.ClusterSpec{
			Name:        clusterName,
			Capacity:    maxSize,
			Preemptible: preemptible,
			Protected:   protected,
			Spec: api.NodeResources{
				CPUCount: cpuCount,
				GPUCount: gpuCount,
				GPUType:  gpuType,
				Memory:   memory,
			},
		}

		cluster, err := beaker.CreateCluster(ctx, account, spec)
		if err != nil {
			return err
		}

		fmt.Printf("Cluster %s created (ID %s)\n", color.BlueString(cluster.Name), color.BlueString(cluster.ID))
		if !cluster.Autoscale {
			return nil
		}

		fmt.Printf("Preparing cluster...")
		ticker := time.NewTicker(3 * time.Second)
		for {
			select {
			case <-ctx.Done():
				fmt.Println(" canceled")
				os.Exit(1)

			case <-ticker.C:
				cluster, err = beaker.Cluster(cluster.ID).Get(ctx)
				if err != nil {
					fmt.Println(" failed")
					return err
				}

				switch cluster.Status {
				case api.ClusterPending:
					continue

				case api.ClusterActive:
					fmt.Println("Success!")

					gpuStr := "none"
					if gpuCount := cluster.NodeShape.GPUCount; gpuCount != 0 {
						gpuStr = strconv.FormatInt(int64(gpuCount), 10)
						if gpuType := cluster.NodeShape.GPUType; gpuType != "" {
							gpuStr += " " + gpuType
						}
					}

					fmt.Print("\nEstimated cost per node: ")
					color.Green("$%v/hour", cluster.NodeCost.Round(2))
					fmt.Println("Nodes may exceed requested parameters to optimize cost:")
					fmt.Println("    CPUs:      ", cluster.NodeShape.CPUCount)
					fmt.Println("    CPU Memory:", cluster.NodeShape.Memory)
					fmt.Println("    GPUs:      ", gpuStr)
					return nil

				case api.ClusterFailed:
					fmt.Println(" failed")
					return errors.New(cluster.StatusMessage)

				default:
					fmt.Println(" failed")
					return fmt.Errorf("unrecognized cluster state: %s", cluster.Status)
				}
			}
		}
	}
	return cmd
}

func newClusterExecutionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "executions <cluster>",
		Short: "List executions in a cluster",
		Args:  cobra.ExactArgs(1),
	}

	var scheduled bool
	var unscheduled bool
	cmd.Flags().BoolVar(&scheduled, "scheduled", false, "Only show scheduled executions")
	cmd.Flags().BoolVar(&unscheduled, "unscheduled", false, "Only show unscheduled executions")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if scheduled && unscheduled {
			return fmt.Errorf("only one of --scheduled and --unscheduled may be set")
		}

		var scheduledOpt *bool
		if scheduled || unscheduled {
			scheduledOpt = &scheduled
		}
		executions, err := beaker.Cluster(args[0]).ListExecutions(ctx, &client.ExecutionFilters{
			Scheduled: scheduledOpt,
		})
		if err != nil {
			return err
		}

		switch format {
		case formatJSON:
			return printJSON(executions)
		default:
			if err := printTableRow(
				"ID",
				"TASK",
				"NAME",
				"NODE",
				"CPU COUNT",
				"GPU COUNT",
				"MEMORY",
				"PRIORITY",
				"STATUS",
			); err != nil {
				return err
			}
			for _, execution := range executions {
				if err := printTableRow(
					execution.ID,
					execution.Task,
					execution.Spec.Name,
					execution.Node,
					execution.Limits.CPUCount,
					execution.Limits.GPUCount,
					execution.Limits.Memory,
					execution.Priority,
					executionStatus(execution.State),
				); err != nil {
					return err
				}
			}
			return nil
		}
	}
	return cmd
}

func executionStatus(state api.ExecutionState) string {
	if state.Scheduled == nil {
		return "pending"
	}
	if state.Started == nil {
		return "starting"
	}
	if state.Ended == nil {
		return "running"
	}
	if state.Finalized == nil {
		return "finalizing"
	}
	return "finished"
}

func newClusterInspectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect <cluster...>",
		Short: "Display detailed information about one or more clusters",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var clusters []*api.Cluster
			for _, id := range args {
				info, err := beaker.Cluster(id).Get(ctx)
				if err != nil {
					return err
				}
				clusters = append(clusters, info)
			}

			return printJSON(clusters)
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

		terminated := false
		var clusters []api.Cluster
		var cursor string
		for {
			var page []api.Cluster
			var err error
			page, cursor, err = beaker.ListClusters(ctx, args[0], &client.ListClusterOptions{
				Cursor:     cursor,
				Terminated: &terminated,
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

		switch format {
		case formatJSON:
			return printJSON(clusters)
		default:
			if err := printTableRow(
				"NAME",
				"GPU TYPE",
				"GPU COUNT",
				"CPU COUNT",
				"MEMORY",
			); err != nil {
				return err
			}
			for _, cluster := range clusters {
				var (
					gpuType  string
					gpuCount string
					cpuCount string
					memory   string
				)
				if cluster.NodeShape != nil {
					gpuType = fmt.Sprintf("%v", cluster.NodeShape.GPUType)
					gpuCount = fmt.Sprintf("%v", cluster.NodeShape.GPUCount)
					cpuCount = fmt.Sprintf("%v", cluster.NodeShape.CPUCount)
					memory = fmt.Sprintf("%v", cluster.NodeShape.Memory)
				}
				if err := printTableRow(
					cluster.Name,
					gpuType,
					gpuCount,
					cpuCount,
					memory,
				); err != nil {
					return err
				}
			}
			return nil
		}
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

			switch format {
			case formatJSON:
				return printJSON(nodes)
			default:
				if err := printTableRow(
					"ID",
					"HOSTNAME",
					"CREATED",
					"CPU COUNT",
					"GPU COUNT",
					"GPU TYPE",
					"MEMORY",
				); err != nil {
					return err
				}
				for _, node := range nodes {
					if err := printTableRow(
						node.ID,
						node.Hostname,
						node.Created,
						node.Limits.CPUCount,
						node.Limits.GPUCount,
						node.Limits.GPUType,
						node.Limits.Memory,
					); err != nil {
						return err
					}
				}
				return nil
			}
		},
	}
}

func newClusterTerminateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "terminate <cluster>",
		Short: "Permanently expire a cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := beaker.Cluster(args[0]).Terminate(ctx); err != nil {
				return err
			}

			fmt.Printf("Terminated %s\n", color.BlueString(args[0]))
			return nil
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
