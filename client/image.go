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

// ImageHandle provides operations on a image.
type ImageHandle struct {
	client *Client
	id     string
}

// CreateImage creates a new image with an optional name.
func (c *Client) CreateImage(
	ctx context.Context,
	spec api.ImageSpec,
	name string,
) (*ImageHandle, error) {
	query := url.Values{}
	if name != "" {
		query.Set("name", name)
	}

	resp, err := c.sendRequest(ctx, http.MethodPost, "/api/v3/images", query, spec)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp.Body)

	var body api.CreateImageResponse
	if err := parseResponse(resp, &body); err != nil {
		return nil, err
	}

	return &ImageHandle{client: c, id: body.ID}, nil
}

// Image gets a handle for a image by name or ID. The returned handle is
// guaranteed throughout its lifetime to refer to the same object, even if that
// object is later renamed.
func (c *Client) Image(ctx context.Context, reference string) (*ImageHandle, error) {
	id, err := c.resolveRef(ctx, "/api/v3/images", reference)
	if err != nil {
		return nil, errors.WithMessage(err, "could not resolve image reference "+reference)
	}

	return &ImageHandle{client: c, id: id}, nil
}

// ID returns a image's stable, unique ID.
func (h *ImageHandle) ID() string {
	return h.id
}

// Get retrieves a image's details.
func (h *ImageHandle) Get(ctx context.Context) (*api.Image, error) {
	uri := path.Join("/api/v3/images", h.id)
	resp, err := h.client.sendRequest(ctx, http.MethodGet, uri, nil, nil)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp.Body)

	var body api.Image
	if err := parseResponse(resp, &body); err != nil {
		return nil, err
	}
	return &body, nil
}

// Repository returns information required to push a image's Docker image.
func (h *ImageHandle) Repository(
	ctx context.Context,
	upload bool,
) (*api.ImageRepository, error) {
	path := path.Join("/api/v3/images", h.id, "repository")
	query := url.Values{"upload": {strconv.FormatBool(upload)}}
	resp, err := h.client.sendRequest(ctx, http.MethodPost, path, query, nil)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp.Body)

	var body api.ImageRepository
	if err := parseResponse(resp, &body); err != nil {
		return nil, err
	}
	return &body, nil
}

// SetName sets a image's name.
func (h *ImageHandle) SetName(ctx context.Context, name string) error {
	path := path.Join("/api/v3/images", h.id)
	body := api.ImagePatchSpec{Name: &name}
	resp, err := h.client.sendRequest(ctx, http.MethodPatch, path, nil, body)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}

// SetDescription sets a image's description.
func (h *ImageHandle) SetDescription(ctx context.Context, description string) error {
	path := path.Join("/api/v3/images", h.id)
	body := api.ImagePatchSpec{Description: &description}
	resp, err := h.client.sendRequest(ctx, http.MethodPatch, path, nil, body)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}

// Commit finalizes an image, unblocking usage and locking it for further
// writes. The image is guaranteed to remain uncommitted on failure.
func (h *ImageHandle) Commit(ctx context.Context) error {
	path := path.Join("/api/v3/images", h.id)
	body := api.ImagePatchSpec{Commit: true}
	resp, err := h.client.sendRequest(ctx, http.MethodPatch, path, nil, body)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}

// GetPermissions gets a summary of the user's permissions on the image.
func (h *ImageHandle) GetPermissions(ctx context.Context) (*api.PermissionSummary, error) {
	return getPermissions(ctx, h.client, path.Join("/api/v3/images", h.ID(), "auth"))
}

// PatchPermissions ammends an image's permissions.
func (h *ImageHandle) PatchPermissions(
	ctx context.Context,
	permissionPatch api.PermissionPatch,
) error {
	path := path.Join("/api/v3/images", h.id, "auth")
	resp, err := h.client.sendRequest(ctx, http.MethodPatch, path, nil, permissionPatch)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}

func (c *Client) SearchImages(
	ctx context.Context,
	searchOptions api.ImageSearchOptions,
	page int,
) ([]api.Image, error) {
	query := url.Values{"page": {strconv.Itoa(page)}}
	resp, err := c.sendRequest(ctx, http.MethodPost, "/api/v3/images/search", query, searchOptions)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp.Body)

	var body []api.Image
	if err := parseResponse(resp, &body); err != nil {
		return nil, err
	}

	return body, nil
}
