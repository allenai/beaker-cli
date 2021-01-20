package main

import (
	"io"
	"os"

	"github.com/beaker/client/api"
	"github.com/spf13/cobra"
)

func newExecutionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execution <command>",
		Short: "Manage executions",
	}
	cmd.AddCommand(newExecutionInspectCommand())
	cmd.AddCommand(newExecutionLogsCommand())
	return cmd
}

func newExecutionInspectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect <execution...>",
		Short: "Display detailed information about one or more executions",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var executions []api.Execution
			for _, id := range args {
				info, err := beaker.Execution(id).Get(ctx)
				if err != nil {
					return err
				}
				executions = append(executions, *info)
			}
			return printExecutions(executions)
		},
	}
}

func newExecutionLogsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logs <execution>",
		Short: "Fetch execution logs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return printExecutionLogs(args[0])
		},
	}
}

func printExecutions(executions []api.Execution) error {
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

func printExecutionLogs(executionID string) error {
	logs, err := beaker.Execution(executionID).GetLogs(ctx)
	if err != nil {
		return err
	}

	_, err = io.Copy(os.Stdout, logs)
	return err
}
