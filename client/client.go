package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/goware/urlx"
	"github.com/pkg/errors"

	"github.com/allenai/beaker/api"
)

// We encode the version as a manually-assigned constant for now. This must be
// updated with each material change to how a client makes requests, and is
// assumed to be monotonically increasing.
const version = "v20181026"

var idPattern = regexp.MustCompile(`^\w\w_[a-z0-9]{12}$`)

// Client is a Beaker HTTP client.
type Client struct {
	baseURL   url.URL
	userToken string

	// Maximum number of attempts to make for each request.
	maxAttempts int
	// Maximum number of milliseconds to wait between attempts.
	maxBackoff float64
	// Maximum number of milliseconds to wait after the first failed attempt.
	backoffBase float64
	// Source of randomness for generating jitter.
	random *rand.Rand
}

// NewClient creates a new Beaker client bound to a single user.
func NewClient(address string, userToken string) (*Client, error) {
	u, err := urlx.ParseWithDefaultScheme(address, "https")
	if err != nil {
		return nil, err
	}

	if u.Path != "" || u.Opaque != "" || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
		return nil, errors.New("address must be base server address in the form [scheme://]host[:port]")
	}

	return &Client{
		baseURL:     *u,
		userToken:   userToken,
		maxAttempts: 10,
		maxBackoff:  30000,
		backoffBase: 5,
		random:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

// Address returns a client's host and port.
func (c *Client) Address() string {
	return c.baseURL.String()
}

// resolveRef resolves a given name or ID to its stable ID. On success, the
// object is guaranteed to exist at the time of call.
func (c *Client) resolveRef(
	ctx context.Context,
	basePath string,
	reference string,
) (string, error) {
	canonicalRef, err := c.canonicalizeRef(ctx, reference)
	if err != nil {
		return "", err
	}

	resp, err := c.sendRequest(ctx, http.MethodGet, path.Join(basePath, canonicalRef), nil, nil)
	if err != nil {
		return "", err
	}
	defer safeClose(resp.Body)

	// Note: This depends on all root-level objects having an "id" field.
	type idResult struct {
		ID string `json:"id"`
	}

	var body idResult
	if err := parseResponse(resp, &body); err != nil {
		return "", err
	}
	return body.ID, nil
}

// canonicalizeRef ensures a given name or ID is in canonical form.
func (c *Client) canonicalizeRef(ctx context.Context, reference string) (string, error) {
	if idPattern.MatchString(reference) {
		return reference, nil
	}

	var userPart string
	var namePart string

	refParts := strings.SplitN(reference, "/", 2)
	if len(refParts) == 1 {
		// User is implicitly scoped; get the user name.
		user, err := c.WhoAmI(ctx)
		if err != nil {
			return "", errors.WithMessage(err, "failed to resolve current user")
		}

		userPart = user.Name
		namePart = refParts[0]
	} else {
		userPart = refParts[0]
		namePart = refParts[1]
	}
	return path.Join(userPart, namePart), nil
}

// doWithRetry sends a request and retries if the client returns an error
// or if a 5xx status code is received. Up to maxAttempts are made,
// waiting up to maxBackoff milliseconds between each attempt.
//
// Uses exponential backoff with full jitter as described here:
// https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/
func (c *Client) doWithRetry(
	client *http.Client,
	req *http.Request,
) (resp *http.Response, err error) {
	for i := 0; i < c.maxAttempts; i++ {
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode != 0 && resp.StatusCode < 500 {
			return
		}

		backoff := math.Min(c.maxBackoff, c.backoffBase*math.Exp2(float64(i))) * c.random.Float64()
		time.Sleep(time.Duration(backoff * float64(time.Millisecond)))
	}
	return
}

func (c *Client) sendRequest(
	ctx context.Context,
	method string,
	path string,
	query map[string]string,
	body interface{},
) (*http.Response, error) {
	b := new(bytes.Buffer)
	if body != nil {
		if err := json.NewEncoder(b).Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := c.newRequest(ctx, method, path, query, b)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout:       30 * time.Second,
		CheckRedirect: copyRedirectHeader,
	}

	return c.doWithRetry(client, req.WithContext(ctx))
}

func (c *Client) newRequest(
	ctx context.Context,
	method string,
	path string,
	query map[string]string,
	body io.Reader,
) (*http.Request, error) {
	var q url.Values
	if len(query) != 0 {
		q = url.Values{}
		for k, v := range query {
			q.Add(k, v)
		}
	}

	u := url.URL{Scheme: c.baseURL.Scheme, Host: c.baseURL.Host, Path: path, RawQuery: q.Encode()}
	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set(api.VersionHeader, version)
	if len(c.userToken) > 0 {
		req.Header.Set("Authorization", "Bearer "+c.userToken)
	}

	return req.WithContext(ctx), nil
}

func copyRedirectHeader(req *http.Request, via []*http.Request) error {
	if len(via) == 0 {
		return nil
	}
	for key, val := range via[0].Header {
		req.Header[key] = val
	}
	return nil
}

// errorFromResponse creates an error from an HTTP response, or nil on success.
func errorFromResponse(resp *http.Response) error {
	// Anything less than 400 isn't an error, so don't produce one.
	if resp.StatusCode < 400 {
		return nil
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Errorf("failed to read response: %v", err)
	}

	var apiErr api.Error
	if err := json.Unmarshal(bytes, &apiErr); err != nil {
		return errors.Errorf("failed to parse response: %v", err)
	}

	return apiErr
}

// responseValue parses the response body and stores the result in the given value.
// The value parameter should be a pointer to the desired structure.
func parseResponse(resp *http.Response, value interface{}) error {
	if err := errorFromResponse(resp); err != nil {
		return err
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, value)
}

// safeClose closes an object while safely handling nils.
func safeClose(closer io.Closer) {
	if closer == nil {
		return
	}
	_ = closer.Close()
}
