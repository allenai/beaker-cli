package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

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
	cmd.AddCommand(newJobAwaitCommand())
	cmd.AddCommand(newJobFinalizeCommand())
	cmd.AddCommand(newJobGetCommand())
	cmd.AddCommand(newJobListCommand())
	cmd.AddCommand(newJobLogsCommand())
	cmd.AddCommand(newJobResultsCommand())
	cmd.AddCommand(newJobStopCommand())
	return cmd
}

func newJobAwaitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "await <job> <status>",
		Short: "Wait until a job reaches a state",
		Args:  cobra.ExactArgs(2),
	}

	var interval time.Duration
	var timeout time.Duration
	cmd.Flags().DurationVar(&interval, "interval", 5*time.Second, "Interval to poll status.")
	cmd.Flags().DurationVar(&timeout, "timeout", 5*time.Minute, "Maximum time to poll for status.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		jobID := args[0]
		status := args[1]

		var job *api.Job
		jobAtStatus := func(ctx context.Context) (bool, error) {
			var err error
			job, err = beaker.Job(jobID).Get(ctx)
			if err != nil {
				return false, err
			}
			return isAtStatus(job.Status, status)
		}
		ctx, cancel := context.WithDeadline(ctx, time.Now().Add(timeout))
		defer cancel()
		message := "Waiting for job to reach status " + status
		if err := await(ctx, message, jobAtStatus, interval); err != nil {
			return err
		}
		return printJobs([]api.Job{*job})
	}
	return cmd
}

func isAtStatus(status api.JobStatus, target string) (bool, error) {
	switch target {
	case "scheduled":
		if status.Scheduled != nil {
			return true, nil
		}
	case "started":
		if status.Started != nil {
			return true, nil
		}
	case "exited":
		if status.Exited != nil {
			return true, nil
		}
	case "failed":
		if status.Failed != nil {
			return true, nil
		}
	case "finalized":
		if status.Finalized != nil {
			return true, nil
		}
	case "canceled":
		if status.Canceled != nil {
			return true, nil
		}
	case "idle":
		if status.IdleSince != nil {
			return true, nil
		}
	default:
		return false, fmt.Errorf("invalid status: %s", target)
	}
	return false, nil
}

func newJobFinalizeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "finalize <job>",
		Short: "Finalize a job",
		Args:  cobra.ExactArgs(1),
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		job, err := beaker.Job(args[0]).Patch(ctx, api.JobPatch{
			Status: &api.JobStatusUpdate{Finalized: true},
		})
		if err != nil {
			return err
		}
		return printJobs([]api.Job{*job})
	}
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

		if node == "" {
			if cluster == "" && len(experiments) == 0 {
				var err error
				if node, err = getCurrentNode(); err != nil {
					return fmt.Errorf("failed to detect node; use --node flag: %w", err)
				}
				opts.Node = &node
			}
		} else { // some node is specified
			if cluster != "" {
				return fmt.Errorf("you cannot specify node and cluster")
			}
			if len(experiments) > 0 {
				return fmt.Errorf("you cannot specify node and experiments")
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
	cmd := &cobra.Command{
		Use:   "logs <job>",
		Short: "Print job logs",
		Args:  cobra.ExactArgs(1),
	}

	var noTimestamps bool
	cmd.Flags().BoolVar(&noTimestamps, "no-timestamps", false, "Don't include timestamps in logs.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		logs, err := beaker.Job(args[0]).GetLogs(ctx)
		if err != nil {
			return err
		}

		if !noTimestamps {
			_, err = io.Copy(os.Stdout, logs)
			return err
		}

		scanner := bufio.NewScanner(logs)
		for scanner.Scan() {
			line := scanner.Text()
			i := strings.Index(line, " ")
			if i < 0 {
				return fmt.Errorf("timestamp not found: %q", line)
			}
			if _, err := fmt.Println(line[i+1:]); err != nil {
				return err
			}
		}
		return nil
	}
	return cmd
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
