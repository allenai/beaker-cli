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
				cluster.FullName,
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
			if dataset.FullName != "" {
				name = dataset.FullName
			}
			if err := printTableRow(
				name,
				dataset.Workspace.FullName,
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
			if experiment.FullName != "" {
				name = experiment.FullName
			}
			var executions []api.Execution
			for _, execution := range experiment.Executions {
				executions = append(executions, *execution)
			}
			if err := printTableRow(
				name,
				experiment.Workspace.FullName,
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
			if group.FullName != "" {
				name = group.FullName
			}
			if err := printTableRow(
				name,
				group.Workspace.FullName,
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
			if image.FullName != "" {
				name = image.FullName
			}
			if err := printTableRow(
				name,
				image.Workspace.FullName,
				image.Author.Name,
				image.Created,
			); err != nil {
				return err
			}
		}
		return nil
	}
}

func printJobs(jobs []api.Job) error {
	switch format {
	case formatJSON:
		return printJSON(jobs)
	default:
		if err := printTableRow(
			"ID",
			"KIND",
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
		for _, job := range jobs {
			var duration time.Duration
			if job.Status.Scheduled != nil {
				end := time.Now()
				if job.Status.Finalized != nil {
					end = *job.Status.Finalized
				}
				duration = end.Sub(*job.Status.Scheduled)
			}

			var scheduled time.Time
			if job.Status.Scheduled != nil {
				scheduled = *job.Status.Scheduled
			}

			var gpus string
			if job.Limits != nil {
				gpus = strconv.Itoa(len(job.Limits.GPUs))
			}

			if err := printTableRow(
				job.ID,
				job.Kind,
				job.Name,
				job.Author.Name,
				jobStatus(job.Status),
				scheduled,
				duration,
				gpus,
				job.Node,
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
				workspace.FullName,
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

func jobStatus(state api.JobStatus) string {
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
		status := jobStatus(execution.State)
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
