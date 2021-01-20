package main

import (
	"fmt"

	"github.com/beaker/client/api"
	"github.com/beaker/client/client"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newWorkspaceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace <command>",
		Short: "Manage workspaces",
	}
	cmd.AddCommand(newWorkspaceArchiveCommand())
	cmd.AddCommand(newWorkspaceCreateCommand())
	cmd.AddCommand(newWorkspaceDatasetsCommand())
	cmd.AddCommand(newWorkspaceInspectCommand())
	cmd.AddCommand(newWorkspaceListCommand())
	cmd.AddCommand(newWorkspacePermissionsCommand())
	cmd.AddCommand(newWorkspaceMoveCommand())
	cmd.AddCommand(newWorkspaceRenameCommand())
	cmd.AddCommand(newWorkspaceUnarchiveCommand())
	return cmd
}

func newWorkspaceArchiveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "archive <workspace>",
		Short: "Archive a workspace, making it read-only",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := beaker.Workspace(ctx, args[0])
			if err != nil {
				return err
			}

			if err := workspace.SetArchived(ctx, true); err != nil {
				return err
			}

			fmt.Printf("Workspace %s archived\n", color.BlueString(args[0]))
			return nil
		},
	}
}

func newWorkspaceCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new workspace",
		Args:  cobra.ExactArgs(1),
	}

	var description string
	var org string
	cmd.Flags().StringVar(&description, "description", "", "Workspace description")
	cmd.Flags().StringVarP(&org, "org", "o", "", "Workpace organization")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		spec := api.WorkspaceSpec{
			Name:         args[0],
			Description:  description,
			Organization: org,
		}

		workspace, err := beaker.CreateWorkspace(ctx, spec)
		if err != nil {
			return err
		}

		if quiet {
			fmt.Println(workspace.ID())
		} else {
			fmt.Printf("Workspace %s created (ID %s)\n", color.BlueString(spec.Name), color.BlueString(workspace.ID()))
		}
		return nil
	}
	return cmd
}

func newWorkspaceDatasetsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "datasets <workspace>",
		Short: "List datasets in a workspace",
		Args:  cobra.ExactArgs(1),
	}

	var all bool
	var archived bool
	var result bool
	var uncommitted bool
	cmd.Flags().BoolVar(&all, "all", false, "Show all datasets including archived, result, and uncommitted datasets")
	cmd.Flags().BoolVar(&archived, "archived", false, "Show only archived datasets")
	cmd.Flags().BoolVar(&result, "result", false, "Show only result datasets")
	cmd.Flags().BoolVar(&uncommitted, "uncommitted", false, "Show only uncommitted datasets")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		workspace, err := beaker.Workspace(ctx, args[0])
		if err != nil {
			return err
		}

		var datasets []api.Dataset
		var cursor string
		for {
			opts := &client.ListDatasetOptions{
				Cursor: cursor,
			}
			if !all {
				opts.Archived = &archived
				opts.ResultsOnly = &result
				committed := !uncommitted
				opts.CommittedOnly = &committed
			}

			var page []api.Dataset
			var err error
			page, cursor, err = workspace.Datasets(ctx, opts)
			if err != nil {
				return err
			}
			datasets = append(datasets, page...)
			if cursor == "" {
				break
			}
		}
		return printDatasets(datasets)
	}
	return cmd
}

func newWorkspaceInspectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect <workspace...>",
		Short: "Display detailed information about one or more workspaces",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var workspaces []api.Workspace
			for _, name := range args {
				workspace, err := beaker.Workspace(ctx, name)
				if err != nil {
					return err
				}

				workspaceInfo, err := workspace.Get(ctx)
				if err != nil {
					return err
				}

				workspaces = append(workspaces, *workspaceInfo)
			}
			return printWorkspaces(workspaces)
		},
	}
}

func newWorkspaceListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <account>",
		Short: "List workspaces in an account",
		Args:  cobra.ExactArgs(1),
	}

	var archived bool
	cmd.Flags().BoolVar(&archived, "archived", false, "Only show archived workspaces")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		var workspaces []api.Workspace
		var cursor string
		for {
			var page []api.Workspace
			var err error
			page, cursor, err = beaker.ListWorkspaces(ctx, args[0], &client.ListWorkspaceOptions{
				Cursor:   cursor,
				Archived: &archived,
			})
			if err != nil {
				return err
			}
			workspaces = append(workspaces, page...)
			if cursor == "" {
				break
			}
		}
		return printWorkspaces(workspaces)
	}
	return cmd
}

func newWorkspacePermissionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "permissions <command>",
		Short: "Manage workspace permissions",
	}
	cmd.AddCommand(newWorkspacePermissionsGrantCommand())
	cmd.AddCommand(newWorkspacePermissionsInspectCommand())
	cmd.AddCommand(newWorkspacePermissionsRevokeCommand())
	cmd.AddCommand(newWorkspacePermissionsSetVisibilityCommand())
	return cmd
}

func newWorkspacePermissionsGrantCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "grant <workspace> <account> <read|write|all>",
		Short: "Grant permissions on a workspace to an account",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := beaker.Workspace(ctx, args[0])
			if err != nil {
				return err
			}

			var permission api.Permission
			switch args[2] {
			case "read":
				permission = api.Read
			case "write":
				permission = api.Write
			case "all":
				permission = api.FullControl
			default:
				return errors.Errorf(`invalid permission: %q; must be "read", "write", or "all"`, args[2])
			}

			return workspace.SetPermissions(ctx, api.WorkspacePermissionPatch{
				Authorizations: map[string]api.Permission{
					args[1]: permission,
				},
			})
		},
	}
}

func newWorkspacePermissionsInspectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect <workspace>",
		Short: "Inspect workspace permissions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := beaker.Workspace(ctx, args[0])
			if err != nil {
				return err
			}

			permissions, err := workspace.Permissions(ctx)
			if err != nil {
				return err
			}

			switch format {
			case formatJSON:
				return printJSON(permissions)
			default:
				visibility := "private"
				if permissions.Public {
					visibility = "public"
				}
				fmt.Printf("Visibility: %s\n", visibility)
				if len(permissions.Authorizations) == 0 {
					return nil
				}

				fmt.Println()
				if err := printTableRow("ACCOUNT", "PERMISSION"); err != nil {
					return err
				}
				for account, permission := range permissions.Authorizations {
					user, err := beaker.User(ctx, account)
					if err != nil {
						return err
					}

					accountInfo, err := user.Get(ctx)
					if err != nil {
						return err
					}

					if err := printTableRow(accountInfo.Name, permission); err != nil {
						return err
					}
				}
				return nil
			}
		},
	}
}

func newWorkspacePermissionsRevokeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke <workspace> <account>",
		Short: "Revoke permissions on a workspace from an account",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := beaker.Workspace(ctx, args[0])
			if err != nil {
				return err
			}

			return workspace.SetPermissions(ctx, api.WorkspacePermissionPatch{
				Authorizations: map[string]api.Permission{
					args[1]: api.NoPermission,
				},
			})
		},
	}
}

func newWorkspacePermissionsSetVisibilityCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set-visibility <workspace> <public|private>",
		Short: "Set the visibility of a workspace to public or private",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := beaker.Workspace(ctx, args[0])
			if err != nil {
				return err
			}

			var public bool
			switch args[1] {
			case "public":
				public = true
			case "private":
			default:
				return fmt.Errorf(`invalid visibility: %q; must be "public" or "private"`, args[1])
			}
			return workspace.SetPermissions(ctx, api.WorkspacePermissionPatch{
				Public: &public,
			})
		},
	}
}

func newWorkspaceMoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "move <workspace> <items...>",
		Short: "Move items into a workspace",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := beaker.Workspace(ctx, args[0])
			if err != nil {
				return err
			}

			if err := workspace.Transfer(ctx, args[1:]...); err != nil {
				return err
			}

			if !quiet {
				fmt.Printf("Transferred %d items into workspace %s\n", len(args)-1, color.BlueString(workspace.ID()))
			}
			return nil
		},
	}
}

func newWorkspaceRenameCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <workspace> <name>",
		Short: "Rename an workspace",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := beaker.Workspace(ctx, args[0])
			if err != nil {
				return err
			}

			if err := workspace.SetName(ctx, args[1]); err != nil {
				return err
			}

			// TODO: This info should probably be part of the client response instead of a separate get.
			info, err := workspace.Get(ctx)
			if err != nil {
				return err
			}

			if quiet {
				fmt.Println(info.ID)
			} else {
				fmt.Printf("Renamed %s to %s\n", color.BlueString(args[0]), args[1])
			}
			return nil
		},
	}
}

func newWorkspaceUnarchiveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "unarchive <workspace>",
		Short: "Unarchive a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := beaker.Workspace(ctx, args[0])
			if err != nil {
				return err
			}

			if err := workspace.SetArchived(ctx, false); err != nil {
				return err
			}

			fmt.Printf("Workspace %s unarchived\n", color.BlueString(args[0]))
			return nil
		},
	}
}

func printWorkspaces(workspaces []api.Workspace) error {
	switch format {
	case formatJSON:
		return printJSON(workspaces)
	default:
		if err := printTableRow(
			"NAME",
			"AUTHOR",
			"DATASETS",
			"EXPERIMENTS",
			"GROUPS",
			"IMAGES",
		); err != nil {
			return err
		}
		for _, workspace := range workspaces {
			if err := printTableRow(
				workspace.Name,
				workspace.Author.Name,
				workspace.Size.Datasets,
				workspace.Size.Experiments,
				workspace.Size.Groups,
				workspace.Size.Images,
			); err != nil {
				return err
			}
		}
		return nil
	}
}
