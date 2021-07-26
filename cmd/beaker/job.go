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
		Use:     "job <command>",
		Short:   "Manage jobs",
		Aliases: []string{"execution"},
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
		Short: "List jobs. Defaults to listing running jobs on the current node.",
		Args:  cobra.NoArgs,
	}

	var cluster string
	var experiments []string
	var finalized bool
	var kind string
	var node string
	cmd.Flags().StringVar(&cluster, "cluster", "", "List jobs on a cluster.")
	cmd.Flags().StringArrayVar(&experiments, "experiment", nil, "List jobs in a set of experiments.")
	cmd.Flags().BoolVar(&finalized, "finalized", false, "List finalized jobs.")
	cmd.Flags().StringVar(&kind, "kind", "", "List jobs of a certain kind. Either 'execution' or 'session'.")
	cmd.Flags().StringVar(&node, "node", "", "List jobs on a node. Defaults to current node if no other filters are specified.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		opts := client.ListJobOpts{
			Cluster:     cluster,
			Experiments: experiments,
			Finalized:   &finalized,
		}
		if kind != "" {
			kind := api.JobKind(kind)
			opts.Kind = &kind
		}
		if node == "" && cluster == "" && len(experiments) == 0 {
			var err error
			if node, err = getCurrentNode(); err != nil {
				return fmt.Errorf("failed to detect node; use --node flag: %w", err)
			}
			opts.Node = &node
		}
		if len(experiments) > 0 {
			opts.Experiments = experiments
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
		Short: "Print job logs",
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
			Status: &api.JobStatusUpdate{Canceled: true},
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
