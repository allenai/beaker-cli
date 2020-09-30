package config

// TODO: Re-assess where this file should go when refactoring the client's package structure.

import (
	"context"
	"net/http"
	"strings"

	api "github.com/beaker/client/api"
	beaker "github.com/beaker/client/client"
	"github.com/pkg/errors"

	"github.com/allenai/beaker/config"
)

// EnsureWorkspace ensures that workspaceRef exists or that the default workspace
// exists if workspaceRef is empty.
// Returns errWorkspaceNotProvided if workspaceRef and the default workspace are empty.
func EnsureWorkspace(
	client *beaker.Client,
	config *config.Config,
	workspaceRef string,
) (string, error) {
	ctx := context.TODO()

	if workspaceRef == "" {
		if config.DefaultWorkspace == "" {
			return "", errors.New(`workspace not provided, either:
1. Pass the --workspace flag
2. Configure a default workspace with 'beaker config set default_workspace <workspace>'`)
		}
		workspaceRef = config.DefaultWorkspace
	}

	// Create the workspace if it doesn't exist.
	if _, err := client.Workspace(ctx, workspaceRef); err != nil {
		if apiErr, ok := err.(api.Error); ok && apiErr.Code == http.StatusNotFound {
			parts := strings.Split(workspaceRef, "/")
			if len(parts) != 2 {
				return "", errors.New("workspace must be formatted like '<account>/<name>'")
			}

			if _, err = client.CreateWorkspace(ctx, api.WorkspaceSpec{
				Organization: parts[0],
				Name:         parts[1],
			}); err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}

	return workspaceRef, nil
}
