package main

import (
	"encoding/json"
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
			var executions []*api.Execution
			for _, id := range args {
				info, err := beaker.Execution(id).Get(ctx)
				if err != nil {
					return err
				}

				executions = append(executions, info)
			}

			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "    ")
			return encoder.Encode(executions)
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

func printExecutionLogs(executionID string) error {
	logs, err := beaker.Execution(executionID).GetLogs(ctx)
	if err != nil {
		return err
	}

	_, err = io.Copy(os.Stdout, logs)
	return err
}
