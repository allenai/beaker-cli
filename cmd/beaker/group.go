package main

import (
	"fmt"
	"strings"

	"github.com/beaker/client/api"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newGroupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group <command>",
		Short: "Manage groups",
	}
	cmd.AddCommand(newGroupAddCommand())
	cmd.AddCommand(newGroupCreateCommand())
	cmd.AddCommand(newGroupDeleteCommand())
	cmd.AddCommand(newGroupExecutionsCommand())
	cmd.AddCommand(newGroupExperimentsCommand())
	cmd.AddCommand(newGroupGetCommand())
	cmd.AddCommand(newGroupRemoveCommand())
	cmd.AddCommand(newGroupRenameCommand())
	cmd.AddCommand(newGroupTasksCommand())
	return cmd
}

func newGroupAddCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add <group> <experiment...>",
		Short: "Add experiments to a group",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids := trimAndUnique(args[1:])
			if err := beaker.Group(args[0]).AddExperiments(ctx, ids); err != nil {
				return err
			}

			if quiet {
				fmt.Println(args[0])
			} else {
				fmt.Printf("Added experiments to %s: %s\n", color.BlueString(args[0]), ids)
			}
			return nil
		},
	}
}

func newGroupCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name> <experiment...>",
		Short: "Create a new experiment group",
		Args:  cobra.MinimumNArgs(1),
	}

	var description string
	var workspace string
	cmd.Flags().StringVar(&description, "desc", "", "Group description")
	cmd.Flags().StringVarP(&workspace, "workspace", "w", "", "Group workspace")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		var err error
		if workspace, err = ensureWorkspace(workspace); err != nil {
			return err
		}

		spec := api.GroupSpec{
			Name:        args[0],
			Description: description,
			Workspace:   workspace,
			Experiments: trimAndUnique(args[1:]),
		}
		group, err := beaker.CreateGroup(ctx, spec)
		if err != nil {
			return err
		}

		if quiet {
			fmt.Println(group.Ref())
		} else {
			fmt.Println("Created group " + color.BlueString(group.Ref()))
		}
		return nil
	}
	return cmd
}

func newGroupDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <group>",
		Short: "Permanently delete a group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := beaker.Group(args[0]).Delete(ctx); err != nil {
				return err
			}

			if quiet {
				fmt.Println(args[0])
			} else {
				fmt.Println("Deleted group " + color.BlueString(args[0]))
			}
			return nil
		},
	}
}

func newGroupExecutionsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "executions <group>",
		Short: "List executions in a group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			experimentIDs, err := beaker.Group(args[0]).Experiments(ctx)
			if err != nil {
				return err
			}

			var executions []api.Execution
			for _, experimentID := range experimentIDs {
				experiment, err := beaker.Experiment(experimentID).Get(ctx)
				if err != nil {
					return err
				}

				for _, execution := range experiment.Executions {
					executions = append(executions, *execution)
				}
			}
			return printExecutions(executions)
		},
	}
}

func newGroupExperimentsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "experiments <group>",
		Short: "List experiments in a group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			experimentIDs, err := beaker.Group(args[0]).Experiments(ctx)
			if err != nil {
				return err
			}

			var experiments []api.Experiment
			for _, experimentID := range experimentIDs {
				experiment, err := beaker.Experiment(experimentID).Get(ctx)
				if err != nil {
					return err
				}

				experiments = append(experiments, *experiment)
			}
			return printExperiments(experiments)
		},
	}
}

func newGroupGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "get <group...>",
		Aliases: []string{"inspect"},
		Short:   "Display detailed information about one or more groups",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var groups []api.Group
			for _, name := range args {
				group, err := beaker.Group(name).Get(ctx)
				if err != nil {
					return err
				}
				groups = append(groups, *group)
			}
			return printGroups(groups)
		},
	}
}

func newGroupRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <group> <experiment...>",
		Short: "Remove experiments from a group",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids := trimAndUnique(args[1:])
			if err := beaker.Group(args[0]).RemoveExperiments(ctx, ids); err != nil {
				return err
			}

			if quiet {
				fmt.Println(args[0])
			} else {
				fmt.Printf("Removed experiments from %s: %s\n", color.BlueString(args[0]), ids)
			}
			return nil
		},
	}
}

func newGroupRenameCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <group> <name>",
		Short: "Rename a group",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			oldName := args[0]
			newName := args[1]

			group, err := beaker.Group(oldName).Patch(
				ctx, api.GroupPatch{Name: &newName},
			)
			if err != nil {
				return err
			}

			if quiet {
				fmt.Println(group.ID)
			} else {
				fmt.Printf("Renamed %s to %s\n", color.BlueString(oldName), group.FullName)
			}
			return nil
		},
	}
}

func newGroupTasksCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "tasks <group>",
		Short: "List tasks in a group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			experimentIDs, err := beaker.Group(args[0]).Experiments(ctx)
			if err != nil {
				return err
			}

			var tasks []api.Task
			for _, experimentID := range experimentIDs {
				groupTasks, err := beaker.Experiment(experimentID).Tasks(ctx)
				if err != nil {
					return err
				}
				tasks = append(tasks, groupTasks...)
			}
			return printTasks(tasks)
		},
	}
}

// Trim and unique a collection of strings, typically used to pre-process IDs.
func trimAndUnique(ids []string) []string {
	if len(ids) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	var unique []string
	for _, id := range ids {
		id := strings.TrimSpace(id)
		if _, ok := seen[id]; !ok {
			seen[id] = true
			unique = append(unique, id)
		}
	}

	return unique
}
