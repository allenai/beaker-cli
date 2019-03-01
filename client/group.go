package client

import (
	"context"
	"net/http"
	"path"

	"github.com/pkg/errors"

	"github.com/allenai/beaker/api"
)

// GroupHandle provides operations on a group.
type GroupHandle struct {
	client *Client
	id     string
}

// CreateGroup creates a new group with an optional name.
func (c *Client) CreateGroup(ctx context.Context, spec api.GroupSpec, defaultOrg string) (*GroupHandle, error) {
	if spec.Org == "" && defaultOrg != "" {
		spec.Org = defaultOrg
	}

	resp, err := c.sendRequest(ctx, http.MethodPost, "/api/v3/groups", nil, spec)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp.Body)

	var body api.CreateGroupResponse
	if err = parseResponse(resp, &body); err != nil {
		return nil, err
	}
	return &GroupHandle{client: c, id: body.ID}, nil
}

// Group gets a handle for a group by name or ID. The returned handle is
// guaranteed throughout its lifetime to refer to the same object, even if that
// object is later renamed.
func (c *Client) Group(ctx context.Context, reference string) (*GroupHandle, error) {
	id, err := c.resolveRef(ctx, "/api/v3/groups", reference)
	if err != nil {
		return nil, errors.WithMessage(err, "could not resolve group reference "+reference)
	}

	return &GroupHandle{client: c, id: id}, nil
}

// ID returns a group's stable, unique ID.
func (h *GroupHandle) ID() string {
	return h.id
}

// Get retrieves a group's details.
func (h *GroupHandle) Get(ctx context.Context) (*api.Group, error) {
	path := path.Join("/api/v3/groups", h.id)
	resp, err := h.client.sendRequest(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp.Body)

	var body api.Group
	if err = parseResponse(resp, &body); err != nil {
		return nil, err
	}
	return &body, nil
}

// SetName sets a group's name.
func (h *GroupHandle) SetName(ctx context.Context, name string) error {
	path := path.Join("/api/v3/groups", h.id)
	body := api.GroupPatchSpec{Name: &name}
	resp, err := h.client.sendRequest(ctx, http.MethodPatch, path, nil, body)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}

// SetDescription sets a group's description.
func (h *GroupHandle) SetDescription(ctx context.Context, description string) error {
	path := path.Join("/api/v3/groups", h.id)
	body := api.GroupPatchSpec{Description: &description}
	resp, err := h.client.sendRequest(ctx, http.MethodPatch, path, nil, body)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}

// Experiments returns the IDs of all experiments within a group.
func (h *GroupHandle) Experiments(ctx context.Context) ([]string, error) {
	path := path.Join("/api/v3/groups", h.id, "experiments")
	resp, err := h.client.sendRequest(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp.Body)

	var body []string
	if err = parseResponse(resp, &body); err != nil {
		return nil, err
	}
	return body, nil
}

// AddExperiments adds experiments by name or ID to a group.
func (h *GroupHandle) AddExperiments(ctx context.Context, experiments []string) error {
	if len(experiments) == 0 {
		return nil
	}

	path := path.Join("/api/v3/groups", h.id)
	body := api.GroupPatchSpec{AddExperiments: experiments}
	resp, err := h.client.sendRequest(ctx, http.MethodPatch, path, nil, body)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)

	return errorFromResponse(resp)
}

// RemoveExperiments removes experiments by name or ID from a group.
func (h *GroupHandle) RemoveExperiments(ctx context.Context, experiments []string) error {
	if len(experiments) == 0 {
		return nil
	}

	path := path.Join("/api/v3/groups", h.id)
	body := api.GroupPatchSpec{RemoveExperiments: experiments}
	resp, err := h.client.sendRequest(ctx, http.MethodPatch, path, nil, body)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)

	return errorFromResponse(resp)
}

// Delete removes a group and its contents.
func (h *GroupHandle) Delete(ctx context.Context) error {
	path := path.Join("/api/v3/groups", h.id)
	resp, err := h.client.sendRequest(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)

	return errorFromResponse(resp)
}

// GetPermissions gets a summary of the user's permissions on the group.
func (h *GroupHandle) GetPermissions(ctx context.Context) (*api.PermissionSummary, error) {
	return getPermissions(ctx, h.client, path.Join("/api/v3/groups", h.ID(), "auth"))
}

// PatchPermissions ammends a group's permissions.
func (h *GroupHandle) PatchPermissions(
	ctx context.Context,
	permissionPatch api.PermissionPatch,
) error {
	path := path.Join("/api/v3/groups", h.id, "auth")
	resp, err := h.client.sendRequest(ctx, http.MethodPatch, path, nil, permissionPatch)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}
