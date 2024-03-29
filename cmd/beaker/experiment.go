package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/beaker/client/api"
	"github.com/beaker/client/client"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newExperimentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "experiment <command>",
		Short: "Manage experiments",
	}
	cmd.AddCommand(newExperimentAwaitCommand())
	cmd.AddCommand(newExperimentCreateCommand())
	cmd.AddCommand(newExperimentDeleteCommand())
	cmd.AddCommand(newExperimentGroupsCommand())
	cmd.AddCommand(newExperimentGetCommand())
	cmd.AddCommand(newExperimentRenameCommand())
	cmd.AddCommand(newExperimentResultsCommand())
	cmd.AddCommand(newExperimentResumeCommand())
	cmd.AddCommand(newExperimentSpecCommand())
	cmd.AddCommand(newExperimentStopCommand())
	cmd.AddCommand(newExperimentTasksCommand())
	return cmd
}

func newExperimentAwaitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "await <experiment> <task> <status>",
		Short: "Wait until a task in an experiment reach the given status",
		Args:  cobra.ExactArgs(3),
	}

	var index bool
	var interval time.Duration
	var timeout time.Duration
	cmd.Flags().BoolVar(&index, "index", false, "Interpret task reference as an index.")
	cmd.Flags().DurationVar(&interval, "interval", 5*time.Second, "Interval to poll status.")
	cmd.Flags().DurationVar(&timeout, "timeout", 5*time.Minute, "Maximum time to poll for status.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		experimentID := args[0]
		taskRef := args[1]
		status := args[2]

		var taskIndex int
		if index {
			var err error
			taskIndex, err = strconv.Atoi(taskRef)
			if err != nil {
				return fmt.Errorf("invalid task index %s: %w", taskRef, err)
			}
		}

		var taskID string
		tasks, err := beaker.Experiment(experimentID).Tasks(ctx)
		if err != nil {
			return err
		}
		for i, task := range tasks {
			if task.Name == taskRef || task.ID == taskRef || (index && i == taskIndex) {
				taskID = task.ID
				break
			}
		}
		if taskID == "" {
			return fmt.Errorf("task not found: %s", taskRef)
		}

		var job api.Job
		experimentAtStatus := func(ctx context.Context) (bool, error) {
			task, err := beaker.Task(taskID).Get(ctx)
			if err != nil {
				return false, err
			}
			if len(task.Jobs) == 0 {
				return false, nil // Controller has not created any jobs yet.
			}
			// Use status of last job.
			job = task.Jobs[len(task.Jobs)-1]
			return isAtStatus(job.Status, status)
		}
		ctx, cancel := context.WithDeadline(ctx, time.Now().Add(timeout))
		defer cancel()
		message := "Waiting for experiment to reach status " + status
		if err := await(ctx, message, experimentAtStatus, interval); err != nil {
			return err
		}
		return printJobs([]api.Job{job})
	}
	return cmd
}

func newExperimentCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <spec-file>",
		Short: "Create a new experiment",
		Args:  cobra.ExactArgs(1),
	}

	var name string
	var workspace string
	var priority string
	cmd.Flags().StringVarP(&name, "name", "n", "", "Assign a name to the experiment")
	cmd.Flags().StringVarP(&workspace, "workspace", "w", "", "Workspace where the experiment will be placed")
	cmd.Flags().StringVarP(&priority, "priority", "p", "", "Assign an execution priority to the experiment")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		specFile, err := openPath(args[0])
		if err != nil {
			return err
		}

		if workspace, err = ensureWorkspace(workspace); err != nil {
			return err
		}

		rawSpec, err := readSpec(specFile)
		if err != nil {
			return err
		}

		experiment, err := beaker.Workspace(workspace).CreateExperimentRaw(
			ctx,
			"application/x-yaml",
			bytes.NewReader(rawSpec),
			&client.ExperimentOpts{Name: name})
		if err != nil {
			return err
		}

		if format == formatJSON {
			return printJSON([]api.Experiment{*experiment})
		}

		if quiet {
			fmt.Println(experiment.ID)
		} else {
			fmt.Printf("Experiment %s submitted. See progress at %s/ex/%s\n",
				color.BlueString(experiment.ID), beaker.Address(), experiment.ID)
		}
		return nil
	}
	return cmd
}

func newExperimentDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <experiment>",
		Short: "Permanently delete an experiment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := beaker.Experiment(args[0]).Delete(ctx); err != nil {
				return err
			}

			fmt.Printf("Deleted %s\n", color.BlueString(args[0]))
			return nil
		},
	}
}

func newExperimentGroupsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "groups <experiment>",
		Short: "List the groups that the experiments belongs to",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupIDs, err := beaker.Experiment(args[0]).Groups(ctx)
			if err != nil {
				return err
			}

			var groups []api.Group
			for _, id := range groupIDs {
				group, err := beaker.Group(id).Get(ctx)
				if err != nil {
					return err
				}
				groups = append(groups, *group)
			}
			return printGroups(groups)
		},
	}
}

func newExperimentGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "get <experiment...>",
		Aliases: []string{"inspect"},
		Short:   "Display detailed information about one or more experiments",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var experiments []api.Experiment
			for _, name := range args {
				exp, err := beaker.Experiment(name).Get(ctx)
				if err != nil {
					return err
				}

				experiments = append(experiments, *exp)
			}
			return printExperiments(experiments)
		},
	}
}

func newExperimentRenameCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <experiment> <name>",
		Short: "Rename an experiment",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			oldName := args[0]
			newName := args[1]
			experiment, err := beaker.Experiment(oldName).Patch(ctx, api.ExperimentPatch{
				Name: &newName,
			})
			if err != nil {
				return err
			}

			if quiet {
				fmt.Println(experiment.ID)
			} else {
				fmt.Printf("Renamed %s to %s\n", color.BlueString(oldName), experiment.FullName)
			}
			return nil
		},
	}
}

func newExperimentResultsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "results <experiment>",
		Short: "Download the results of an experiment",
		Long: `Download the results of an experiment.

One folder will be created for each task in the experiment. The name of the
folder will be the name of the task or its ID if the task does not have a name.
If the task has executed multiple times, the results of the latest execution
will be downloaded.

Example: beaker experiment results --output experiment <experiment>
experiment/
  task/
    file`,
		Args: cobra.ExactArgs(1),
	}
	flags := addFetchFlags(cmd)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		tasks, err := beaker.Experiment(args[0]).Tasks(ctx)
		if err != nil {
			return err
		}

		for _, task := range tasks {
			name := task.ID
			if task.Name != "" {
				name = task.Name
			}
			if len(task.Jobs) == 0 {
				fmt.Printf("Task %s has no executions; skipping\n", name)
				continue
			}
			job := task.Jobs[len(task.Jobs)-1] // Use last job.
			outputPath := path.Join(flags.outputPath, name)
			if err := fetchDataset(
				job.Execution.Result.Beaker,
				outputPath,
				flags.prefix,
				flags.concurrency,
			); err != nil {
				return fmt.Errorf("fetching result of %s: %w", job.ID, err)
			}
		}
		return nil
	}
	return cmd
}

func newExperimentResumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume <experiment>",
		Short: "Resume a preempted experiment",
		Args:  cobra.ExactArgs(1),
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := beaker.Experiment(args[0]).Resume(ctx); err != nil {
			return err
		}

		return nil
	}
	return cmd
}

func newExperimentSpecCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spec <experiment>",
		Short: "Get the spec of an experiment as YAML",
		Args:  cobra.ExactArgs(1),
	}

	var version string
	cmd.Flags().StringVar(&version, "version", "v2-alpha", "Spec version: v1 or v2-alpha")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		spec, err := beaker.Experiment(args[0]).Spec(ctx, version, format == formatJSON)
		if err != nil {
			return err
		}
		defer spec.Close()

		_, err = io.Copy(os.Stdout, spec)
		return err
	}
	return cmd
}

func newExperimentStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop <experiment...>",
		Short: "Stop one or more running experiments",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, name := range args {
				if err := beaker.Experiment(name).Stop(ctx); err != nil {
					// We want to stop as many of the requested experiments as possible.
					// Therefore we print to STDERR here instead of returning.
					fmt.Fprintln(os.Stderr, color.RedString("Error:"), err)
				}

				fmt.Println(name)
			}
			return nil
		},
	}
}

func newExperimentTasksCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "tasks <experiment>",
		Short: "List the tasks in an experiment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tasks, err := beaker.Experiment(args[0]).Tasks(ctx)
			if err != nil {
				return err
			}
			return printTasks(tasks)
		},
	}
}

// readSpec reads an experiment spec from YAML.
func readSpec(r io.Reader) ([]byte, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	specTemplate, err := template.New("spec").Parse(string(b))
	if err != nil {
		return nil, err
	}

	envVars := map[string]string{}
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		envVars[parts[0]] = parts[1]
	}

	type templateParams struct {
		Env map[string]string
	}
	buf := &bytes.Buffer{}
	if err := specTemplate.Execute(buf, templateParams{Env: envVars}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func openPath(p string) (io.Reader, error) {
	// Special case: "-" means read from STDIN.
	if p == "-" {
		return os.Stdin, nil
	}
	return os.Open(p)
}
