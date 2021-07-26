package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"strings"

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
	cmd.AddCommand(newExperimentCreateCommand())
	cmd.AddCommand(newExperimentDeleteCommand())
	cmd.AddCommand(newExperimentExecutionsCommand())
	cmd.AddCommand(newExperimentGroupsCommand())
	cmd.AddCommand(newExperimentGetCommand())
	cmd.AddCommand(newExperimentRenameCommand())
	cmd.AddCommand(newExperimentResumeCommand())
	cmd.AddCommand(newExperimentSpecCommand())
	cmd.AddCommand(newExperimentStopCommand())
	cmd.AddCommand(newExperimentTasksCommand())
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

func newExperimentExecutionsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "executions <experiment>",
		Short: "List the executions in an experiment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobs, err := listJobs(client.ListJobOpts{
				Experiments: args,
			})
			if err != nil {
				return err
			}
			return printJobs(jobs)
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
