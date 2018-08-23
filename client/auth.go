package client

import (
	"context"
	"net/http"
	"path"

	"github.com/allenai/beaker/api"
)

// WhoAmI returns a client's active user.
func (c *Client) WhoAmI(ctx context.Context) (*api.User, error) {
	uri := path.Join("/api/v3/auth/whoami")
	resp, err := c.sendRequest(ctx, http.MethodGet, uri, nil, nil)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp.Body)

	if err := errorFromResponse(resp); err != nil {
		return nil, err
	}

	var user api.User
	if err := parseResponse(resp, &user); err != nil {
		return nil, err
	}

	return &user, nil
}
