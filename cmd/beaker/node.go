package main

import (
	"github.com/beaker/client/api"
	"github.com/spf13/cobra"
)

func newNodeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node <command>",
		Short: "Manage nodes",
	}
	cmd.AddCommand(newNodeCordonCommand())
	cmd.AddCommand(newNodeExecutionsCommand())
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

func newNodeExecutionsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "executions <node>",
		Short: "List the executions of a node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			executions, err := beaker.Node(args[0]).ListExecutions(ctx)
			if err != nil {
				return err
			}
			return printExecutions(executions.Data)
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
