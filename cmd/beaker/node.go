package main

import (
	"github.com/beaker/client/api"
	"github.com/beaker/client/client"
	"github.com/spf13/cobra"
)

func newNodeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node <command>",
		Short: "Manage nodes",
	}
	cmd.AddCommand(newNodeCordonCommand())
	cmd.AddCommand(newNodeDeleteCommand())
	cmd.AddCommand(newNodeExecutionsCommand())
	cmd.AddCommand(newNodeGetCommand())
	cmd.AddCommand(newNodeUncordonCommand())
	return cmd
}

func newNodeCordonCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "cordon <node>",
		Short: "Cordon a node preventing it from running new executions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cordoned := true
			return beaker.Node(args[0]).Patch(ctx, &api.NodePatchSpec{
				Cordoned: &cordoned,
			})
		},
	}
}

func newNodeDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <node>",
		Short: "Permanently delete a node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return beaker.Node(args[0]).Delete(ctx)
		},
	}
}

func newNodeExecutionsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "executions <node>",
		Short: "List the executions of a node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			kind := api.JobKindExecution
			jobs, err := listJobs(client.ListJobOpts{
				Kind: &kind,
				Node: &args[0],
			})
			if err != nil {
				return err
			}
			return printJobs(jobs)
		},
	}
}

func newNodeGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "get <node...>",
		Aliases: []string{"inspect"},
		Short:   "Display detailed information about one or more nodes",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var nodes []api.Node
			for _, id := range args {
				node, err := beaker.Node(id).Get(ctx)
				if err != nil {
					return err
				}
				nodes = append(nodes, *node)
			}
			return printNodes(nodes)
		},
	}
}

func newNodeUncordonCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "uncordon <node>",
		Short: "Uncordon a node allowing it to run new executions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cordoned := false
			return beaker.Node(args[0]).Patch(ctx, &api.NodePatchSpec{
				Cordoned: &cordoned,
			})
		},
	}
}
