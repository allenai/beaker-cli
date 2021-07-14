package main

import (
	"fmt"
	"strconv"
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
				formatted = t.Format(time.Stamp)
			}
		} else if d, ok := cell.(time.Duration); ok {
			// Format duration as HH:MM:SS.
			second := d % time.Minute
			minute := (d - second) % time.Hour
			hour := d - minute - second
			formatted = fmt.Sprintf(
				"%02d:%02d:%02d",
				hour/time.Hour,
				minute/time.Minute,
				second/time.Second)
		} else {
			formatted = fmt.Sprintf("%v", cell)
		}
		if formatted == "" {
			formatted = "N/A"
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
				if cluster.NodeShape.Memory != nil {
					memory = cluster.NodeShape.Memory.String()
				}
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
			"SOURCE EXECUTION",
		); err != nil {
			return err
		}
		for _, dataset := range datasets {
			name := dataset.ID
			if dataset.Name != "" {
				name = dataset.Name
			}
			if err := printTableRow(
				name,
				dataset.Workspace.Name,
				dataset.Author.Name,
				dataset.Committed,
				dataset.SourceExecution,
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
			"NAME",
			"AUTHOR",
			"STATUS",
			"SCHEDULED",
			"DURATION",
			"GPUS",
			"NODE",
		); err != nil {
			return err
		}
		for _, execution := range executions {
			var duration time.Duration
			if execution.State.Scheduled != nil {
				end := time.Now()
				if execution.State.Finalized != nil {
					end = *execution.State.Finalized
				}
				duration = end.Sub(*execution.State.Scheduled)
			}

			var scheduled time.Time
			if execution.State.Scheduled != nil {
				scheduled = *execution.State.Scheduled
			}

			if err := printTableRow(
				execution.ID,
				execution.Spec.Name,
				execution.Author.Name,
				executionStatus(execution.State),
				scheduled,
				duration,
				len(execution.Limits.GPUs),
				execution.Node,
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
		); err != nil {
			return err
		}
		for _, experiment := range experiments {
			name := experiment.ID
			if experiment.Name != "" {
				name = experiment.Name
			}
			var executions []api.Execution
			for _, execution := range experiment.Executions {
				executions = append(executions, *execution)
			}
			if err := printTableRow(
				name,
				experiment.Workspace.Name,
				experiment.Author.Name,
				experiment.Created,
				executionsStatus(executions),
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
		); err != nil {
			return err
		}
		for _, group := range groups {
			name := group.ID
			if group.Name != "" {
				name = group.Name
			}
			if err := printTableRow(
				name,
				group.Workspace.Name,
				group.Author.Name,
				group.Created,
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

func printSessions(sessions []api.Session) error {
	switch format {
	case formatJSON:
		return printJSON(sessions)
	default:
		if err := printTableRow(
			"ID",
			"NAME",
			"AUTHOR",
			"STATUS",
			"SCHEDULED",
			"DURATION",
			"GPUS",
			"NODE",
		); err != nil {
			return err
		}
		for _, session := range sessions {
			var duration time.Duration
			if session.State.Scheduled != nil {
				end := time.Now()
				if session.State.Finalized != nil {
					end = *session.State.Finalized
				}
				duration = end.Sub(*session.State.Scheduled)
			}

			var scheduled time.Time
			if session.State.Scheduled != nil {
				scheduled = *session.State.Scheduled
			}

			var gpus string
			if session.Limits != nil {
				gpus = strconv.Itoa(len(session.Limits.GPUs))
			}

			if err := printTableRow(
				session.ID,
				session.Name,
				session.Author.Name,
				executionStatus(session.State),
				scheduled,
				duration,
				gpus,
				session.Node,
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
			"STATUS",
		); err != nil {
			return err
		}
		for _, task := range tasks {
			if err := printTableRow(
				task.ID,
				task.ExperimentID,
				task.Name,
				task.Author.Name,
				executionsStatus(task.Executions),
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
			"ROLE",
		); err != nil {
			return err
		}
		for _, user := range users {
			if err := printTableRow(
				user.ID,
				user.Name,
				user.DisplayName,
				user.Role,
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
			accountInfo, err := beaker.User(account).Get(ctx)
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
	case state.Failed != nil:
		return "failed"
	case state.Finalized != nil:
		if state.ExitCode != nil && *state.ExitCode == 0 {
			return "succeeded"
		}
		return "failed"
	case state.Exited != nil:
		return "uploading"
	case state.Started != nil:
		return "running"
	case state.Scheduled != nil:
		return "starting"
	default:
		return "pending"
	}
}

func executionsStatus(executions []api.Execution) string {
	counts := make(map[string]int)
	for _, execution := range executions {
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
