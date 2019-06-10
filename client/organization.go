package client

import (
	"context"
	"net/http"
	"path"

	"github.com/pkg/errors"
)

// VerifyOrgExists validates existence of an organization.
func (c *Client) VerifyOrgExists(ctx context.Context, org string) error {
	resp, err := c.sendRequest(ctx, http.MethodGet, path.Join("/api/v3/orgs", org), nil, nil)
	defer safeClose(resp.Body)
	if err != nil {
		return errors.WithMessage(err, "could not resolve organization "+org)
	}

	return errors.WithMessage(errorFromResponse(resp), org)
}
