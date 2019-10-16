package config

// TODO: Re-assess where this file should go when refactoring the client's package structure.

import (
	"context"
	"net/http"
	"path"

	api "github.com/beaker/client/api"
	beaker "github.com/beaker/client/client"

	"github.com/allenai/beaker/config"
)

// EnsureDefaultWorkspace uses the configured default workspace if it is available.
// Otherwise it falls back to a set of default workspaces, creating them if needed.
func EnsureDefaultWorkspace(
	client *beaker.Client,
	config *config.Config,
	org string,
) (string, error) {
	ctx := context.TODO()

	// If the user configured a default workspace, use it.
	if config.DefaultWorkspace != "" {
		return config.DefaultWorkspace, nil
	}

	author, err := client.WhoAmI(ctx)
	if err != nil {
		return "", err
	}

	// If an org is specified, use the "<org>/<author>-default" workspace.
	// Otherwise, use the "<author>/default" workspace.
	var workspaceName string
	var workspaceRef string
	if org == "" {
		workspaceName = "default"
		workspaceRef = path.Join(author.Name, workspaceName)
	} else {
		workspaceName = author.Name + "-default"
		workspaceRef = path.Join(org, workspaceName)
	}

	if _, err = client.Workspace(ctx, workspaceRef); err != nil {
		if apiErr, ok := err.(api.Error); ok && apiErr.Code == http.StatusNotFound {
			if _, err = client.CreateWorkspace(ctx, api.WorkspaceSpec{
				Name:         workspaceName,
				Organization: org,
			}); err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}

	return workspaceRef, nil
}
