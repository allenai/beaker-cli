package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/allenai/bytefmt"
	"github.com/beaker/client/api"
	"github.com/beaker/client/client"
	"github.com/fatih/color"
	"github.com/pkg/errors"
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

		ws, err := beaker.Workspace(ctx, workspace)
		if err != nil {
			return err
		}

		experiment, err := ws.CreateExperimentRaw(
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
			experiment, err := beaker.Experiment(ctx, args[0])
			if err != nil {
				return err
			}

			if err := experiment.Delete(ctx); err != nil {
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
			experiment, err := beaker.Experiment(ctx, args[0])
			if err != nil {
				return err
			}

			info, err := experiment.Get(ctx)
			if err != nil {
				return err
			}

			var executions []api.Execution
			for _, execution := range info.Executions {
				executions = append(executions, *execution)
			}
			return printExecutions(executions)
		},
	}
}

func newExperimentGroupsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "groups <experiment>",
		Short: "List the groups that the experiments belongs to",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			experiment, err := beaker.Experiment(ctx, args[0])
			if err != nil {
				return err
			}

			groupIDs, err := experiment.Groups(ctx)
			if err != nil {
				return err
			}

			var groups []api.Group
			for _, id := range groupIDs {
				group, err := beaker.Group(ctx, id)
				if err != nil {
					return err
				}

				info, err := group.Get(ctx)
				if err != nil {
					return err
				}
				groups = append(groups, *info)
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
				experiment, err := beaker.Experiment(ctx, name)
				if err != nil {
					return err
				}

				exp, err := experiment.Get(ctx)
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
			experiment, err := beaker.Experiment(ctx, args[0])
			if err != nil {
				return err
			}

			if err := experiment.SetName(ctx, args[1]); err != nil {
				return err
			}

			// TODO: This info should probably be part of the client response instead of a separate get.
			exp, err := experiment.Get(ctx)
			if err != nil {
				return err
			}

			if quiet {
				fmt.Println(exp.ID)
			} else {
				fmt.Printf("Renamed %s to %s\n", color.BlueString(exp.ID), exp.FullName)
			}
			return nil
		},
	}
}

func newExperimentResumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume <experiment>",
		Short: "Resume a preempted experiment and return the experiment ID for the new experiment",
		Args:  cobra.ExactArgs(1),
	}

	var name string
	cmd.Flags().StringVarP(&name, "name", "n", "", "Name for the new experiment")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		experiment, err := beaker.ResumeExperiment(ctx, args[0], name)
		if err != nil {
			return err
		}

		fmt.Println(experiment.ID())
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
		experiment, err := beaker.Experiment(ctx, args[0])
		if err != nil {
			return err
		}

		spec, err := experiment.Spec(ctx, version, format == formatJSON)
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
				experiment, err := beaker.Experiment(ctx, name)
				if err != nil {
					return err
				}

				if err := experiment.Stop(ctx); err != nil {
					// We want to stop as many of the requested experiments as possible.
					// Therefore we print to STDERR here instead of returning.
					fmt.Fprintln(os.Stderr, color.RedString("Error:"), err)
				}

				fmt.Println(experiment.ID())
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
			experiment, err := beaker.Experiment(ctx, args[0])
			if err != nil {
				return err
			}

			tasks, err := experiment.Tasks(ctx)
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

// canonicalizeSpecV1 fills out JSON fields used by the API from YAML fields parsed from disk.
func canonicalizeSpecV1(spec *api.ExperimentSpecV1) error {
	// TODO: This should be unnecessary when the service accepts YAML directly.
	for i := range spec.Tasks {
		reqs := &spec.Tasks[i].Spec.Requirements
		if reqs.CPU < 0 {
			return errors.Errorf("couldn't parse cpu argument '%.2f' because it was negative", reqs.CPU)
		}
		reqs.MilliCPU = int(reqs.CPU * 1000)
		if reqs.MemoryHuman != "" {
			size, err := bytefmt.Parse(reqs.MemoryHuman)
			if err != nil {
				return errors.Wrapf(err, "invalid memory value %q", reqs.MemoryHuman)
			}
			reqs.Memory = int64(size.Int64())
		}
	}
	return nil
}
