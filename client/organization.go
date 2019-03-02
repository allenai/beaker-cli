package client

import (
	"context"
	"net/http"
	"path"

	"github.com/pkg/errors"
)

// Org gets a handle for an org by name or ID. The returned handle is
// guaranteed throughout its lifetime to refer to the same object, even if that
// object is later renamed.
func (c *Client) Org(ctx context.Context, reference string) error {
	resp, err := c.sendRequest(ctx, http.MethodGet, path.Join("/api/v3/orgs", reference), nil, nil)
	defer safeClose(resp.Body)
	if err != nil {
		return errors.WithMessage(err, "could not resolve org reference "+reference)
	}

	type idResult struct {
		ID string `json:"id"`
	}

	var body idResult
	if err := parseResponse(resp, &body); err != nil {
		return errors.WithMessage(err, "could not parse org response "+reference)
	}

	return nil
}
