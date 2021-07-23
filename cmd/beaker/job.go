package main

import (
	"fmt"
	"io"
	"os"

	"github.com/beaker/client/api"
	"github.com/beaker/client/client"
	"github.com/spf13/cobra"
)

func newJobCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job <command>",
		Short: "Manage jobs",
	}
	cmd.AddCommand(newJobGetCommand())
	cmd.AddCommand(newJobListCommand())
	cmd.AddCommand(newJobLogsCommand())
	cmd.AddCommand(newJobResultsCommand())
	cmd.AddCommand(newJobStopCommand())
	return cmd
}

func newJobGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "get <job...>",
		Aliases: []string{"inspect"},
		Short:   "Display detailed information about one or more jobs",
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

func newJobListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List jobs",
		Args:  cobra.NoArgs,
	}

	var all bool
	var kind string
	var cluster string
	var node string
	var finalized bool
	cmd.Flags().BoolVar(&all, "all", false, "List all jobs.")
	cmd.Flags().StringVar(&cluster, "cluster", "", "Cluster to list jobs.")
	cmd.Flags().StringVar(&node, "node", "", "Node to list jobs. Defaults to current node.")
	cmd.Flags().BoolVar(&finalized, "finalized", false, "Show only finalized jobs")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		var opts client.ListJobOpts
		if kind != "" {
			kind := api.JobKind(kind)
			opts.Kind = &kind
		}
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

func newJobLogsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logs <job>",
		Short: "Fetch job logs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return printJobLogs(args[0])
		},
	}
}

func newJobResultsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "results <job>",
		Short: "Get job results",
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

func newJobStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop a pending or running job",
		Args:  cobra.ExactArgs(1),
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		job, err := beaker.Job(args[0]).Patch(ctx, api.JobPatch{
			State: &api.JobStatusUpdate{Canceled: true},
		})
		if err != nil {
			return err
		}
		return printJobs([]api.Job{*job})
	}
	return cmd
}

// listJobs follows all cursors to get a complete list of jobs.
func listJobs(opts client.ListJobOpts) ([]api.Job, error) {
	var jobs []api.Job
	for {
		var page api.Jobs
		page, err := beaker.ListJobs(ctx, &opts)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, page.Data...)
		if page.Next == "" {
			break
		}
		opts.Cursor = page.Next
	}
	return jobs, nil
}

func printJobLogs(jobID string) error {
	logs, err := beaker.Job(jobID).GetLogs(ctx)
	if err != nil {
		return err
	}

	_, err = io.Copy(os.Stdout, logs)
	return err
}
