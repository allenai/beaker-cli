package client

import (
	"context"
	"net/http"
	"net/url"
	"path"

	"github.com/pkg/errors"

	"github.com/allenai/beaker/api"
)

// ExperimentHandle provides operations on an experiment.
type ExperimentHandle struct {
	client *Client
	id     string
}

// CreateExperiment creates a new experiment with an optional name.
func (c *Client) CreateExperiment(
	ctx context.Context,
	spec api.ExperimentSpec,
	name string,
) (*ExperimentHandle, error) {
	var query url.Values
	query.Set("name", name)
	resp, err := c.sendRequest(ctx, http.MethodPost, "/api/v3/experiments", query, spec)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp.Body)

	var id string
	if err := parseResponse(resp, &id); err != nil {
		return nil, err
	}

	return &ExperimentHandle{client: c, id: id}, nil
}

// Experiment gets a handle for an experiment by name or ID. The returned handle
// is guaranteed throughout its lifetime to refer to the same object, even if
// that object is later renamed.
func (c *Client) Experiment(ctx context.Context, reference string) (*ExperimentHandle, error) {
	id, err := c.resolveRef(ctx, "/api/v3/experiments", reference)
	if err != nil {
		return nil, errors.WithMessage(err, "could not resolve experiment reference "+reference)
	}

	return &ExperimentHandle{client: c, id: id}, nil
}

// ID returns an experiment's stable, unique ID.
func (h *ExperimentHandle) ID() string {
	return h.id
}

// Get retrieves an experiment's details, including a summary of contained tasks.
func (h *ExperimentHandle) Get(ctx context.Context) (*api.Experiment, error) {
	path := path.Join("/api/v3/experiments", h.id)
	resp, err := h.client.sendRequest(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp.Body)

	var experiment api.Experiment
	if err := parseResponse(resp, &experiment); err != nil {
		return nil, err
	}

	return &experiment, nil
}

// SetName sets an experiment's name.
func (h *ExperimentHandle) SetName(ctx context.Context, name string) error {
	path := path.Join("/api/v3/experiments", h.id)
	body := api.ExperimentPatchSpec{Name: &name}
	resp, err := h.client.sendRequest(ctx, http.MethodPatch, path, nil, body)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}

// SetDescription sets an experiment's description
func (h *ExperimentHandle) SetDescription(ctx context.Context, description string) error {
	path := path.Join("/api/v3/experiments", h.id)
	body := api.ExperimentPatchSpec{Description: &description}
	resp, err := h.client.sendRequest(ctx, http.MethodPatch, path, nil, body)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}

// Stop cancels all uncompleted tasks for an experiment. If the experiment has
// already completed, this succeeds without effect.
func (h *ExperimentHandle) Stop(ctx context.Context) error {
	path := path.Join("/api/v3/experiments", h.id, "stop")
	resp, err := h.client.sendRequest(ctx, http.MethodPut, path, nil, nil)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}

// GetPermissions gets a summary of the user's permissions on the experiment.
func (h *ExperimentHandle) GetPermissions(ctx context.Context) (*api.PermissionSummary, error) {
	return getPermissions(ctx, h.client, path.Join("/api/v3/experiments", h.ID(), "auth"))
}

// PatchPermissions ammends an experiment's permissions.
func (h *ExperimentHandle) PatchPermissions(
	ctx context.Context,
	permissionPatch api.PermissionPatch,
) error {
	path := path.Join("/api/v3/experiments", h.id, "auth")
	resp, err := h.client.sendRequest(ctx, http.MethodPatch, path, nil, permissionPatch)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}
