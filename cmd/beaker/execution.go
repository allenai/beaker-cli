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
	cmd.AddCommand(newExecutionGetCommand())
	cmd.AddCommand(newExecutionLogsCommand())
	cmd.AddCommand(newExecutionResultsCommand())
	cmd.AddCommand(newExecutionStopCommand())
	return cmd
}

func newExecutionGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "get <execution...>",
		Aliases: []string{"inspect"},
		Short:   "Display detailed information about one or more executions",
		Args:    cobra.MinimumNArgs(1),
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

func newExecutionResultsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "results <execution>",
		Short: "Get execution results",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := beaker.Job(args[0]).GetResults(ctx)
			if err != nil {
				return err
			}

			switch format {
			case formatJSON:
				return printJSON(results)
			default:
				if err := printTableRow("METRIC", "VALUE"); err != nil {
					return err
				}
				for metric, value := range results.Metrics {
					if err := printTableRow(metric, value); err != nil {
						return err
					}
				}
				return nil
			}
		},
	}
}

func newExecutionStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop <execution>",
		Short: "Stop an execution, optionally running it again",
		Args:  cobra.ExactArgs(1),
	}

	var requeue bool
	cmd.Flags().BoolVar(&requeue, "requeue", false, "Run the execution again")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return beaker.Job(args[0]).Preempt(ctx)
	}
	return cmd
}

func printExecutionLogs(executionID string) error {
	logs, err := beaker.Job(executionID).GetLogs(ctx)
	if err != nil {
		return err
	}

	_, err = io.Copy(os.Stdout, logs)
	return err
}
