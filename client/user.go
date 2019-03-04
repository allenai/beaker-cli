package client

import (
	"context"
	"net/http"
	"path"

	"github.com/pkg/errors"

	"github.com/allenai/beaker/api"
)

// UserHandle provides operations on a user.
type UserHandle struct {
	client *Client
	id     string
}

// User gets a handle for a user by name or ID. The returned handle is
// guaranteed throughout its lifetime to refer to the same object, even if that
// object is later renamed.
func (c *Client) User(ctx context.Context, reference string) (*UserHandle, error) {
	id, err := c.resolveRef(ctx, "/api/v3/users", reference)
	if err != nil {
		return nil, errors.WithMessage(err, "could not resolve user reference "+reference)
	}

	return &UserHandle{client: c, id: id}, nil
}

// ID returns a user's stable, unique ID.
func (h *UserHandle) ID() string {
	return h.id
}

// Get retrieves a user's details.
func (h *UserHandle) Get(ctx context.Context) (*api.UserDetail, error) {
	uri := path.Join("/api/v3/users", h.id)
	resp, err := h.client.sendRequest(ctx, http.MethodGet, uri, nil, nil)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp.Body)

	var body api.UserDetail
	if err := parseResponse(resp, &body); err != nil {
		return nil, err
	}
	return &body, nil
}
