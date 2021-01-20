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
	cmd.AddCommand(newGroupInspectCommand())
	cmd.AddCommand(newGroupRemoveCommand())
	cmd.AddCommand(newGroupRenameCommand())
	return cmd
}

func newGroupAddCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add <group> <experiment...>",
		Short: "Add experiments to a group",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			group, err := beaker.Group(ctx, args[0])
			if err != nil {
				return err
			}

			ids := trimAndUnique(args[1:])
			if err := group.AddExperiments(ctx, ids); err != nil {
				return err
			}

			if quiet {
				fmt.Println(group.ID())
			} else {
				fmt.Printf("Added experiments to %s: %s\n", color.BlueString(group.ID()), ids)
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
			fmt.Println(group.ID())
		} else {
			fmt.Println("Created group " + color.BlueString(group.ID()))
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
			group, err := beaker.Group(ctx, args[0])
			if err != nil {
				return err
			}

			if err := group.Delete(ctx); err != nil {
				return err
			}

			if quiet {
				fmt.Println(group.ID())
			} else {
				fmt.Println("Deleted group " + color.BlueString(group.ID()))
			}
			return nil
		},
	}
}

func newGroupInspectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect <group...>",
		Short: "Display detailed information about one or more groups",
		Args:  cobra.MinimumNArgs(1),
	}

	var contents bool
	cmd.Flags().BoolVar(&contents, "contents", false, "Include group contents in output")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		type detail struct {
			api.Group
			Experiments []string `json:"experiments,omitempty"`
		}

		var groups []detail
		for _, name := range args {
			group, err := beaker.Group(ctx, name)
			if err != nil {
				return err
			}

			info, err := group.Get(ctx)
			if err != nil {
				return err
			}

			var experiments []string
			if contents {
				if experiments, err = group.Experiments(ctx); err != nil {
					return err
				}
			}

			groups = append(groups, detail{*info, experiments})
		}
		return printJSON(groups)
	}
	return cmd
}

func newGroupRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <group> <experiment...>",
		Short: "Remove experiments from a group",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			group, err := beaker.Group(ctx, args[0])
			if err != nil {
				return err
			}

			ids := trimAndUnique(args[1:])
			if err := group.RemoveExperiments(ctx, ids); err != nil {
				return err
			}

			if quiet {
				fmt.Println(group.ID())
			} else {
				fmt.Printf("Removed experiments from %s: %s\n", color.BlueString(group.ID()), ids)
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
			group, err := beaker.Group(ctx, args[0])
			if err != nil {
				return err
			}

			if err := group.SetName(ctx, args[1]); err != nil {
				return err
			}

			// TODO: This info should probably be part of the client response instead of a separate get.
			info, err := group.Get(ctx)
			if err != nil {
				return err
			}

			if quiet {
				fmt.Println(info.ID)
			} else {
				fmt.Printf("Renamed %s to %s\n", color.BlueString(info.ID), info.DisplayID())
			}
			return nil
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
