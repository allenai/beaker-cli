package client

import (
	"context"
	"net/http"
	"path"
	"strconv"

	"github.com/pkg/errors"

	"github.com/allenai/beaker/api"
)

// BlueprintHandle provides operations on a blueprint.
type BlueprintHandle struct {
	client *Client
	id     string
}

// CreateBlueprint creates a new blueprint with an optional name.
func (c *Client) CreateBlueprint(
	ctx context.Context,
	spec api.BlueprintSpec,
	name string,
) (*BlueprintHandle, error) {
	var query map[string]string
	if name != "" {
		query = map[string]string{"name": name}
	}

	resp, err := c.sendRequest(ctx, http.MethodPost, "/api/v3/blueprints", query, spec)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp.Body)

	var body api.CreateBlueprintResponse
	if err := parseResponse(resp, &body); err != nil {
		return nil, err
	}

	return &BlueprintHandle{client: c, id: body.ID}, nil
}

// Blueprint gets a handle for a blueprint by name or ID. The returned handle is
// guaranteed throughout its lifetime to refer to the same object, even if that
// object is later renamed.
func (c *Client) Blueprint(ctx context.Context, reference string) (*BlueprintHandle, error) {
	id, err := c.resolveRef(ctx, "/api/v3/blueprints", reference)
	if err != nil {
		return nil, errors.WithMessage(err, "could not resolve blueprint reference "+reference)
	}

	return &BlueprintHandle{client: c, id: id}, nil
}

// ID returns a blueprint's stable, unique ID.
func (h *BlueprintHandle) ID() string {
	return h.id
}

// Get retrieves a blueprint's details.
func (h *BlueprintHandle) Get(ctx context.Context) (*api.Blueprint, error) {
	uri := path.Join("/api/v3/blueprints", h.id)
	resp, err := h.client.sendRequest(ctx, http.MethodGet, uri, nil, nil)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp.Body)

	var body api.Blueprint
	if err := parseResponse(resp, &body); err != nil {
		return nil, err
	}
	return &body, nil
}

// Repository returns information required to push a blueprint's Docker image.
func (h *BlueprintHandle) Repository(
	ctx context.Context,
	upload bool,
) (*api.BlueprintRepository, error) {
	path := path.Join("/api/v3/blueprints", h.id, "repository")
	query := map[string]string{"upload": strconv.FormatBool(upload)}
	resp, err := h.client.sendRequest(ctx, http.MethodPost, path, query, nil)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp.Body)

	var body api.BlueprintRepository
	if err := parseResponse(resp, &body); err != nil {
		return nil, err
	}
	return &body, nil
}

// SetName sets a blueprint's name.
func (h *BlueprintHandle) SetName(ctx context.Context, name string) error {
	path := path.Join("/api/v3/blueprints", h.id)
	body := api.BlueprintPatchSpec{Name: &name}
	resp, err := h.client.sendRequest(ctx, http.MethodPatch, path, nil, body)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}

// SetDescription sets a blueprint's description.
func (h *BlueprintHandle) SetDescription(ctx context.Context, description string) error {
	path := path.Join("/api/v3/blueprints", h.id)
	body := api.BlueprintPatchSpec{Description: &description}
	resp, err := h.client.sendRequest(ctx, http.MethodPatch, path, nil, body)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}

// Commit finalizes a blueprint, unblocking usage and locking it for further
// writes. The blueprint is guaranteed to remain uncommitted on failure.
func (h *BlueprintHandle) Commit(ctx context.Context) error {
	path := path.Join("/api/v3/blueprints", h.id)
	body := api.BlueprintPatchSpec{Commit: true}
	resp, err := h.client.sendRequest(ctx, http.MethodPatch, path, nil, body)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}

// PatchPermissions ammends a blueprint's permissions.
func (h *BlueprintHandle) PatchPermissions(
	ctx context.Context,
	permissionPatch api.PermissionPatch,
) error {
	path := path.Join("/api/v3/blueprints", h.id, "auth")
	resp, err := h.client.sendRequest(ctx, http.MethodPatch, path, nil, permissionPatch)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}

func (c *Client) SearchBlueprints(
	ctx context.Context,
	searchOptions api.BlueprintSearchOptions,
	page int,
) ([]api.Blueprint, error) {
	query := map[string]string{"page": strconv.Itoa(page)}
	resp, err := c.sendRequest(ctx, http.MethodPost, "/api/v3/blueprints/search", query, searchOptions)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp.Body)

	var body []api.Blueprint
	if err := parseResponse(resp, &body); err != nil {
		return nil, err
	}

	return body, nil
}
