package client

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/pkg/errors"

	"github.com/allenai/beaker/api"
)

// DatasetHandle provides operations on a dataset.
type DatasetHandle struct {
	client *Client
	id     string
}

// CreateDataset creates a new dataset with an optional name.
func (c *Client) CreateDataset(
	ctx context.Context,
	spec api.DatasetSpec,
	name string,
) (*DatasetHandle, error) {
	query := url.Values{}
	if name != "" {
		query.Set("name", name)
	}

	resp, err := c.sendRequest(ctx, http.MethodPost, "/api/v3/datasets", query, spec)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp.Body)

	var body api.CreateDatasetResponse
	if err := parseResponse(resp, &body); err != nil {
		return nil, err
	}
	return &DatasetHandle{client: c, id: body.ID}, nil
}

// Dataset gets a handle for a dataset by name or ID. The returned handle is
// guaranteed throughout its lifetime to refer to the same object, even if that
// object is later renamed.
func (c *Client) Dataset(ctx context.Context, reference string) (*DatasetHandle, error) {
	id, err := c.resolveRef(ctx, "/api/v3/datasets", reference)
	if err != nil {
		return nil, errors.WithMessage(err, "could not resolve dataset reference "+reference)
	}

	return &DatasetHandle{client: c, id: id}, nil
}

// ID returns a dataset's stable, unique ID.
func (h *DatasetHandle) ID() string {
	return h.id
}

// Get retrieves a dataset's details.
func (h *DatasetHandle) Get(ctx context.Context) (*api.Dataset, error) {
	uri := path.Join("/api/v3/datasets", h.id)
	resp, err := h.client.sendRequest(ctx, http.MethodGet, uri, nil, nil)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp.Body)

	var body api.Dataset
	if err := parseResponse(resp, &body); err != nil {
		return nil, err
	}
	return &body, nil
}

// Manifest retrieves a manifest for a dataset's contents.
func (h *DatasetHandle) Manifest(ctx context.Context) (*api.DatasetManifest, error) {
	path := path.Join("/api/v3/datasets", h.id, "manifest")
	resp, err := h.client.sendRequest(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp.Body)

	var body api.DatasetManifest
	if err = parseResponse(resp, &body); err != nil {
		return nil, err
	}
	return &body, nil
}

// SetName sets a dataset's name.
func (h *DatasetHandle) SetName(ctx context.Context, name string) error {
	path := path.Join("/api/v3/datasets", h.id)
	body := api.DatasetPatchSpec{Name: &name}
	resp, err := h.client.sendRequest(ctx, http.MethodPatch, path, nil, body)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}

// SetDescription sets a dataset's description.
func (h *DatasetHandle) SetDescription(ctx context.Context, description string) error {
	path := path.Join("/api/v3/datasets", h.id)
	body := api.DatasetPatchSpec{Description: &description}
	resp, err := h.client.sendRequest(ctx, http.MethodPatch, path, nil, body)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}

// Commit finalizes a dataset, unblocking usage and locking it for further
// writes. The dataset is guaranteed to remain uncommitted on failure.
func (h *DatasetHandle) Commit(ctx context.Context) error {
	path := path.Join("/api/v3/datasets", h.id)
	body := api.DatasetPatchSpec{Commit: true}
	resp, err := h.client.sendRequest(ctx, http.MethodPatch, path, nil, body)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}

// GetPermissions gets a summary of the user's permissions on the dataset.
func (h *DatasetHandle) GetPermissions(ctx context.Context) (*api.PermissionSummary, error) {
	return getPermissions(ctx, h.client, path.Join("/api/v3/datasets", h.ID(), "auth"))
}

// PatchPermissions ammends a dataset's permissions.
func (h *DatasetHandle) PatchPermissions(
	ctx context.Context,
	permissionPatch api.PermissionPatch,
) error {
	path := path.Join("/api/v3/datasets", h.id, "auth")
	resp, err := h.client.sendRequest(ctx, http.MethodPatch, path, nil, permissionPatch)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}

func (c *Client) SearchDatasets(
	ctx context.Context,
	searchOptions api.DatasetSearchOptions,
	page int,
) ([]api.Dataset, error) {
	query := url.Values{"page": {strconv.Itoa(page)}}
	resp, err := c.sendRequest(ctx, http.MethodPost, "/api/v3/datasets/search", query, searchOptions)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp.Body)

	var body []api.Dataset
	if err := parseResponse(resp, &body); err != nil {
		return nil, err
	}

	return body, nil
}
