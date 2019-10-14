package config

import (
	"context"
	"fmt"
	"net/http"
	"path"

	api "github.com/beaker/client/api"
	beaker "github.com/beaker/client/client"
	"github.com/fatih/color"
)

// EnsureDefaultWorkspace uses the configured default workspace if it is available.
// Otherwise it falls back to a set of default workspaces, creating them if needed.
func EnsureDefaultWorkspace(
	client *beaker.Client,
	config *Config,
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

	// If an org isn't specified, use the "<author>/default" workspace.
	// Otherwise, use the "<org>/<author>" workspace.
	var workspaceName string
	var workspaceRef string
	if org == "" {
		workspaceName = "default"
		workspaceRef = path.Join(author.Name, workspaceName)
	} else {
		workspaceName = author.Name
		workspaceRef = path.Join(org, workspaceName)
	}

	if _, err = client.Workspace(ctx, workspaceRef); err != nil {
		if apiErr, ok := err.(api.Error); ok {
			if apiErr.Code == http.StatusNotFound {
				if _, err = client.CreateWorkspace(ctx, api.WorkspaceSpec{
					Name:         workspaceName,
					Organization: org,
				}); err != nil {
					return "", err
				}
			}
		}
		return "", err
	}

	fmt.Printf("No workspace specified; using default workspace %s.\n", color.BlueString(workspaceRef))
	return workspaceRef, nil
}
