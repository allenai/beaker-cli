package client

import (
	"context"
	"net/http"
	"path"

	"github.com/pkg/errors"
)

// UserHandle provides operations on a user.
type OrgHandle struct {
	client *Client
	id     string
}

// Org gets a handle for an org by name or ID. The returned handle is
// guaranteed throughout its lifetime to refer to the same object, even if that
// object is later renamed.
func (c *Client) Org(ctx context.Context, reference string) (*OrgHandle, error) {
	resp, err := c.sendRequest(ctx, http.MethodGet, path.Join("/api/v3/orgs", reference), nil, nil)
	defer safeClose(resp.Body)
	if err != nil {
		return nil, errors.WithMessage(err, "could not resolve org reference "+reference)
	}

	type idResult struct {
		ID string `json:"id"`
	}

	var body idResult
	if err := parseResponse(resp, &body); err != nil {
		return nil, errors.WithMessage(err, "could not parse org response "+reference)
	}

	return &OrgHandle{client: c, id: body.ID}, nil
}
