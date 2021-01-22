package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/beaker/client/api"
)

func printJSON(v interface{}) error {
	return jsonOut.Encode(v)
}

func printTableRow(cells ...interface{}) error {
	var cellStrings []string
	for _, cell := range cells {
		var formatted string
		if t, ok := cell.(time.Time); ok {
			if !t.IsZero() {
				formatted = t.Format(time.RFC3339)
			}
		} else {
			formatted = fmt.Sprintf("%v", cell)
		}
		cellStrings = append(cellStrings, formatted)
	}
	_, err := fmt.Fprintln(tableOut, strings.Join(cellStrings, "\t"))
	return err
}

func printClusters(clusters []api.Cluster) error {
	switch format {
	case formatJSON:
		return printJSON(clusters)
	default:
		if err := printTableRow(
			"NAME",
			"GPU TYPE",
			"GPU COUNT",
			"CPU COUNT",
			"MEMORY",
			"AUTOSCALE",
		); err != nil {
			return err
		}
		for _, cluster := range clusters {
			var (
				gpuType  string
				gpuCount int
				cpuCount float64
				memory   string
			)
			if cluster.NodeShape != nil {
				gpuType = cluster.NodeShape.GPUType
				gpuCount = cluster.NodeShape.GPUCount
				cpuCount = cluster.NodeShape.CPUCount
				memory = cluster.NodeShape.Memory
			}
			if err := printTableRow(
				cluster.Name,
				gpuType,
				gpuCount,
				cpuCount,
				memory,
				cluster.Autoscale,
			); err != nil {
				return err
			}
		}
		return nil
	}
}

func printDatasets(datasets []api.Dataset) error {
	switch format {
	case formatJSON:
		return printJSON(datasets)
	default:
		if err := printTableRow(
			"ID",
			"WORKSPACE",
			"AUTHOR",
			"COMMITTED",
			"SOURCE TASK",
			"ARCHIVED",
		); err != nil {
			return err
		}
		for _, dataset := range datasets {
			name := dataset.ID
			if dataset.Name != "" {
				name = dataset.Name
			}
			var source string
			if dataset.SourceTask != nil {
				source = *dataset.SourceTask
			}
			var archived string
			if dataset.Archived {
				archived = "archived"
			}
			if err := printTableRow(
				name,
				dataset.Workspace.Name,
				dataset.Author.Name,
				dataset.Committed,
				source,
				archived,
			); err != nil {
				return err
			}
		}
		return nil
	}
}

func printExecutions(executions []api.Execution) error {
	switch format {
	case formatJSON:
		return printJSON(executions)
	default:
		if err := printTableRow(
			"ID",
			"TASK",
			"NAME",
			"NODE",
			"CPU COUNT",
			"GPU COUNT",
			"MEMORY",
			"PRIORITY",
			"STATUS",
		); err != nil {
			return err
		}
		for _, execution := range executions {
			if err := printTableRow(
				execution.ID,
				execution.Task,
				execution.Spec.Name,
				execution.Node,
				execution.Limits.CPUCount,
				execution.Limits.GPUCount,
				execution.Limits.Memory,
				execution.Priority,
				executionStatus(execution.State),
			); err != nil {
				return err
			}
		}
		return nil
	}
}

func printExperiments(experiments []api.Experiment) error {
	switch format {
	case formatJSON:
		return printJSON(experiments)
	default:
		if err := printTableRow(
			"ID",
			"WORKSPACE",
			"AUTHOR",
			"CREATED",
			"STATUS",
			"ARCHIVED",
		); err != nil {
			return err
		}
		for _, experiment := range experiments {
			name := experiment.ID
			if experiment.Name != "" {
				name = experiment.Name
			}
			var archived string
			if experiment.Archived {
				archived = "archived"
			}
			if err := printTableRow(
				name,
				experiment.Workspace.Name,
				experiment.Author.Name,
				experiment.Created,
				experimentStatus(experiment),
				archived,
			); err != nil {
				return err
			}
		}
		return nil
	}
}

func printGroups(groups []api.Group) error {
	switch format {
	case formatJSON:
		return printJSON(groups)
	default:
		if err := printTableRow(
			"ID",
			"WORKSPACE",
			"AUTHOR",
			"CREATED",
			"ARCHIVED",
		); err != nil {
			return err
		}
		for _, group := range groups {
			name := group.ID
			if group.Name != "" {
				name = group.Name
			}
			var archived string
			if group.Archived {
				archived = "archived"
			}
			if err := printTableRow(
				name,
				group.Workspace.Name,
				group.Author.Name,
				group.Created,
				archived,
			); err != nil {
				return err
			}
		}
		return nil
	}
}

func printImages(images []api.Image) error {
	switch format {
	case formatJSON:
		return printJSON(images)
	default:
		if err := printTableRow(
			"ID",
			"WORKSPACE",
			"AUTHOR",
			"CREATED",
		); err != nil {
			return err
		}
		for _, image := range images {
			name := image.ID
			if image.Name != "" {
				name = image.Name
			}
			if err := printTableRow(
				name,
				image.Workspace.Name,
				image.Author.Name,
				image.Created,
			); err != nil {
				return err
			}
		}
		return nil
	}
}

func printMembers(members []api.OrgMembership) error {
	switch format {
	case formatJSON:
		return printJSON(members)
	default:
		if err := printTableRow(
			"ID",
			"NAME",
			"DISPLAY NAME",
			"ROLE",
		); err != nil {
			return err
		}
		for _, member := range members {
			if err := printTableRow(
				member.User.ID,
				member.User.Name,
				member.User.DisplayName,
				member.Role,
			); err != nil {
				return err
			}
		}
		return nil
	}
}

func printNodes(nodes []api.Node) error {
	switch format {
	case formatJSON:
		return printJSON(nodes)
	default:
		if err := printTableRow(
			"ID",
			"HOSTNAME",
			"CPU COUNT",
			"GPU COUNT",
			"GPU TYPE",
			"MEMORY",
			"STATUS",
		); err != nil {
			return err
		}
		for _, node := range nodes {
			status := "ok"
			if node.Cordoned != nil {
				status = "cordoned"
			}
			if err := printTableRow(
				node.ID,
				node.Hostname,
				node.Limits.CPUCount,
				node.Limits.GPUCount,
				node.Limits.GPUType,
				node.Limits.Memory,
				status,
			); err != nil {
				return err
			}
		}
		return nil
	}
}

func printOrganizations(orgs []api.Organization) error {
	switch format {
	case formatJSON:
		return printJSON(orgs)
	default:
		if err := printTableRow(
			"ID",
			"NAME",
			"DISPLAY NAME",
		); err != nil {
			return err
		}
		for _, org := range orgs {
			if err := printTableRow(
				org.ID,
				org.Name,
				org.DisplayName,
			); err != nil {
				return err
			}
		}
		return nil
	}
}

func printSecrets(secrets []api.Secret) error {
	switch format {
	case formatJSON:
		return printJSON(secrets)
	default:
		if err := printTableRow("NAME", "CREATED", "UPDATED"); err != nil {
			return err
		}
		for _, secret := range secrets {
			if err := printTableRow(
				secret.Name,
				secret.Created,
				secret.Updated,
			); err != nil {
				return err
			}
		}
		return nil
	}
}

func printTasks(tasks []api.Task) error {
	switch format {
	case formatJSON:
		return printJSON(tasks)
	default:
		if err := printTableRow(
			"ID",
			"EXPERIMENT",
			"NAME",
			"AUTHOR",
		); err != nil {
			return err
		}
		for _, task := range tasks {
			if err := printTableRow(
				task.ID,
				task.ExperimentID,
				task.Name,
				task.Author.Name,
			); err != nil {
				return err
			}
		}
		return nil
	}
}

func printUsers(users []api.UserDetail) error {
	switch format {
	case formatJSON:
		return printJSON(users)
	default:
		if err := printTableRow(
			"ID",
			"NAME",
			"DISPLAY NAME",
		); err != nil {
			return err
		}
		for _, user := range users {
			if err := printTableRow(
				user.ID,
				user.Name,
				user.DisplayName,
			); err != nil {
				return err
			}
		}
		return nil
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

func printWorkspacePermissions(permissions *api.WorkspacePermissionSummary) error {
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
}

func executionStatus(state api.ExecutionState) string {
	switch {
	case state.Scheduled == nil:
		return "pending"
	case state.Started == nil:
		return "starting"
	case state.Ended == nil:
		return "running"
	case state.Finalized == nil:
		return "finalizing"
	default:
		return "finished"
	}
}

func experimentStatus(experiment api.Experiment) string {
	counts := make(map[string]int)
	for _, execution := range experiment.Executions {
		status := executionStatus(execution.State)
		count, ok := counts[status]
		if ok {
			counts[status] = count + 1
		} else {
			counts[status] = 1
		}
	}
	var parts []string
	for status, count := range counts {
		parts = append(parts, fmt.Sprintf("%d %s", count, status))
	}
	return strings.Join(parts, ", ")
}
