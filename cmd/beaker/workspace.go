package main

import (
	"fmt"
	"path"

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
	cmd.AddCommand(newWorkspaceExperimentsCommand())
	cmd.AddCommand(newWorkspaceGetCommand())
	cmd.AddCommand(newWorkspaceGroupsCommand())
	cmd.AddCommand(newWorkspaceImagesCommand())
	cmd.AddCommand(newWorkspaceListCommand())
	cmd.AddCommand(newWorkspaceMoveCommand())
	cmd.AddCommand(newWorkspacePermissionsCommand())
	cmd.AddCommand(newWorkspaceRenameCommand())
	cmd.AddCommand(newWorkspaceResultsCommand())
	cmd.AddCommand(newWorkspaceUnarchiveCommand())
	return cmd
}

func newWorkspaceArchiveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "archive <workspace>",
		Short: "Archive a workspace, making it read-only",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := beaker.Workspace(args[0]).Patch(ctx, api.WorkspacePatch{
				Archive: api.BoolPtr(true),
			}); err != nil {
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

		wsRef, err := beaker.CreateWorkspace(ctx, spec)
		if err != nil {
			return err
		}
		workspace, err := wsRef.Get(ctx)
		if err != nil {
			return err
		}

		if quiet {
			fmt.Println(workspace.ID)
		} else {
			fmt.Printf("Workspace %s created. See details at %s/ws/%s\n",
				color.BlueString(workspace.FullName), beaker.Address(), workspace.FullName)
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
	var result bool
	var text string
	var uncommitted bool
	cmd.Flags().BoolVar(&all, "all", false, "Show all datasets including result, and uncommitted datasets")
	cmd.Flags().BoolVar(&result, "result", false, "Show only result datasets")
	cmd.Flags().StringVar(&text, "text", "", "Only show datasets matching the text")
	cmd.Flags().BoolVar(&uncommitted, "uncommitted", false, "Show only uncommitted datasets")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		workspace := beaker.Workspace(args[0])

		var datasets []api.Dataset
		var cursor string
		for {
			opts := &client.ListDatasetOptions{
				Cursor: cursor,
				Text:   text,
			}
			if !all {
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

func newWorkspaceExperimentsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "experiments <workspace>",
		Short: "List experiments in a workspace",
		Args:  cobra.ExactArgs(1),
	}

	var text string
	cmd.Flags().StringVar(&text, "text", "", "Only show experiments matching the text")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		workspace := beaker.Workspace(args[0])

		var experiments []api.Experiment
		var cursor string
		for {
			var page []api.Experiment
			var err error
			if page, cursor, err = workspace.Experiments(ctx, &client.ListExperimentOptions{
				Cursor: cursor,
				Text:   text,
			}); err != nil {
				return err
			}
			experiments = append(experiments, page...)
			if cursor == "" {
				break
			}
		}
		return printExperiments(experiments)
	}
	return cmd
}

func newWorkspaceGroupsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "groups <workspace>",
		Short: "List groups in a workspace",
		Args:  cobra.ExactArgs(1),
	}

	var text string
	cmd.Flags().StringVar(&text, "text", "", "Only show groups matching the text")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		workspace := beaker.Workspace(args[0])

		var groups []api.Group
		var cursor string
		for {
			var page []api.Group
			var err error
			if page, cursor, err = workspace.Groups(ctx, &client.ListGroupOptions{
				Cursor: cursor,
				Text:   text,
			}); err != nil {
				return err
			}
			groups = append(groups, page...)
			if cursor == "" {
				break
			}
		}
		return printGroups(groups)
	}
	return cmd
}

func newWorkspaceImagesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "images <workspace>",
		Short: "List images in a workspace",
		Args:  cobra.ExactArgs(1),
	}

	var text string
	cmd.Flags().StringVar(&text, "text", "", "Only show images matching the text")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		workspace := beaker.Workspace(args[0])

		var images []api.Image
		var cursor string
		for {
			opts := &client.ListImageOptions{
				Cursor: cursor,
				Text:   text,
			}

			var page []api.Image
			var err error
			page, cursor, err = workspace.Images(ctx, opts)
			if err != nil {
				return err
			}
			images = append(images, page...)
			if cursor == "" {
				break
			}
		}
		return printImages(images)
	}
	return cmd
}

func newWorkspaceGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "get <workspace...>",
		Aliases: []string{"inspect"},
		Short:   "Display detailed information about one or more workspaces",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var workspaces []api.Workspace
			for _, name := range args {
				workspace, err := beaker.Workspace(name).Get(ctx)
				if err != nil {
					return err
				}

				workspaces = append(workspaces, *workspace)
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
	var text string
	cmd.Flags().BoolVar(&archived, "archived", false, "Only show archived workspaces")
	cmd.Flags().StringVar(&text, "text", "", "Only show workspaces matching the text")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		var workspaces []api.Workspace
		var cursor string
		for {
			var page []api.Workspace
			var err error
			page, cursor, err = beaker.ListWorkspaces(ctx, args[0], &client.ListWorkspaceOptions{
				Cursor:   cursor,
				Archived: &archived,
				Text:     text,
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
	cmd.AddCommand(newWorkspacePermissionsGetCommand())
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

			workspace := beaker.Workspace(args[0])
			if err := workspace.SetPermissions(ctx, api.WorkspacePermissionPatch{
				Authorizations: map[string]api.Permission{
					args[1]: permission,
				},
			}); err != nil {
				return err
			}

			if quiet {
				return nil
			}
			permissions, err := workspace.Permissions(ctx)
			if err != nil {
				return err
			}
			return printWorkspacePermissions(permissions)
		},
	}
}

func newWorkspacePermissionsGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "get <workspace>",
		Aliases: []string{"inspect"},
		Short:   "Get workspace permissions",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			permissions, err := beaker.Workspace(args[0]).Permissions(ctx)
			if err != nil {
				return err
			}
			return printWorkspacePermissions(permissions)
		},
	}
}

func newWorkspacePermissionsRevokeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke <workspace> <account>",
		Short: "Revoke permissions on a workspace from an account",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace := beaker.Workspace(args[0])
			if err := workspace.SetPermissions(ctx, api.WorkspacePermissionPatch{
				Authorizations: map[string]api.Permission{
					args[1]: api.NoPermission,
				},
			}); err != nil {
				return err
			}

			if quiet {
				return nil
			}
			permissions, err := workspace.Permissions(ctx)
			if err != nil {
				return err
			}
			return printWorkspacePermissions(permissions)
		},
	}
}

func newWorkspacePermissionsSetVisibilityCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set-visibility <workspace> <public|private>",
		Short: "Set the visibility of a workspace to public or private",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var public bool
			switch args[1] {
			case "public":
				public = true
			case "private":
			default:
				return fmt.Errorf(`invalid visibility: %q; must be "public" or "private"`, args[1])
			}

			workspace := beaker.Workspace(args[0])
			if err := workspace.SetPermissions(ctx, api.WorkspacePermissionPatch{
				Public: &public,
			}); err != nil {
				return err
			}

			if quiet {
				return nil
			}
			permissions, err := workspace.Permissions(ctx)
			if err != nil {
				return err
			}
			return printWorkspacePermissions(permissions)
		},
	}
}

func newWorkspaceMoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "move <workspace> <items...>",
		Short: "Move items into a workspace",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := beaker.Workspace(args[0]).Transfer(ctx, args[1:]...); err != nil {
				return err
			}

			if !quiet {
				fmt.Printf("Transferred %d items into workspace %s\n", len(args)-1, color.BlueString(args[0]))
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
			oldName := args[0]
			newName := args[1]

			workspace, err := beaker.Image(oldName).Patch(ctx, api.ImagePatch{
				Name: &newName,
			})
			if err != nil {
				return err
			}

			if quiet {
				fmt.Println(workspace.ID)
			} else {
				fmt.Printf("Renamed %s to %s\n", color.BlueString(oldName), workspace.FullName)
			}
			return nil
		},
	}
}

func newWorkspaceResultsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "results <workspace>",
		Short: "Download the results of every experiment in a workspace.",
		Long: `Download the results of every experiment in a workspace.

One folder will be created for each experiment in the workspace. The name
of the folder will be the name of the experiment or its ID if the experiment
does not have a name. Within each experiment's folder, one folder will be
created for each task in the experiment. The name of the folder will be the
name of the task or its ID if the task does not have a name. If the task has
executed multiple times, the results of the latest execution will be downloaded.

Example: beaker workspace results --output workspace <workspace>
workspace/
  experiment/
    task/
      file`,
		Args: cobra.ExactArgs(1),
	}
	flags := addFetchFlags(cmd)
	var text string
	cmd.Flags().StringVar(&text, "text", "", "Only download results of experiments matching the text")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		workspace := beaker.Workspace(args[0])

		var experiments []api.Experiment
		var cursor string
		for {
			var page []api.Experiment
			var err error
			if page, cursor, err = workspace.Experiments(ctx, &client.ListExperimentOptions{
				Cursor: cursor,
				Text:   text,
			}); err != nil {
				return err
			}
			experiments = append(experiments, page...)
			if cursor == "" {
				break
			}
		}
		if !quiet {
			fmt.Printf("Found %d experiments\n", len(experiments))
		}

		for _, experiment := range experiments {
			tasks, err := beaker.Experiment(experiment.ID).Tasks(ctx)
			if err != nil {
				return err
			}
			experimentName := experiment.ID
			if experiment.Name != "" {
				experimentName = experiment.Name
			}

			for _, task := range tasks {
				taskName := task.ID
				if task.Name != "" {
					taskName = task.Name
				}
				if len(task.Jobs) == 0 {
					fmt.Printf("Task %s/%s has no executions; skipping\n", experimentName, taskName)
					continue
				}
				job := task.Jobs[len(task.Jobs)-1] // Use last job.
				outputPath := path.Join(flags.outputPath, experimentName, taskName)
				if err := fetchDataset(
					job.Execution.Result.Beaker,
					outputPath,
					flags.prefix,
					flags.concurrency,
				); err != nil {
					return fmt.Errorf("fetching result of %s: %w", job.ID, err)
				}
			}
		}
		return nil
	}
	return cmd
}

func newWorkspaceUnarchiveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "unarchive <workspace>",
		Short: "Unarchive a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := beaker.Workspace(args[0]).Patch(ctx, api.WorkspacePatch{
				Archive: api.BoolPtr(false),
			}); err != nil {
				return err
			}

			fmt.Printf("Workspace %s unarchived\n", color.BlueString(args[0]))
			return nil
		},
	}
}
