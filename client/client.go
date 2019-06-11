package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/goware/urlx"
	retryable "github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"

	"github.com/allenai/beaker/api"
)

// We encode the version as a manually-assigned constant for now. This must be
// updated with each material change to how a client makes requests, and is
// assumed to be monotonically increasing.
const version = "v20190301"

var idPattern = regexp.MustCompile(`^\w\w_[a-z0-9]{12}$`)

// Client is a Beaker HTTP client.
type Client struct {
	baseURL   url.URL
	userToken string
}

// NewClient creates a new Beaker client bound to a single user.
func NewClient(address string, userToken string) (*Client, error) {
	u, err := urlx.ParseWithDefaultScheme(address, "https")
	if err != nil {
		return nil, err
	}

	// First condition permits single trailing slash in url
	if (u.Path != "" && u.Path != "/") || u.Opaque != "" || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
		return nil, errors.New("address must be base server address in the form [scheme://]host[:port]")
	}

	return &Client{
		baseURL:   *u,
		userToken: userToken,
	}, nil
}

func newRetryableClient(httpClient *http.Client) *retryable.Client {
	return &retryable.Client{
		HTTPClient:   httpClient,
		Logger:       &errorLogger{Logger: log.New(os.Stderr, "", log.LstdFlags)},
		RetryWaitMin: 100 * time.Millisecond,
		RetryWaitMax: 30 * time.Second,
		RetryMax:     9,
		CheckRetry:   retryable.DefaultRetryPolicy,
		Backoff:      exponentialJitterBackoff,
		ErrorHandler: retryable.PassthroughErrorHandler,
	}
}

type errorLogger struct {
	Logger *log.Logger
}

func (l *errorLogger) Printf(template string, args ...interface{}) {
	if strings.HasPrefix(template, "[ERR]") {
		l.Logger.Printf(template, args...)
	}
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

func (c *Client) sendRequest(
	ctx context.Context,
	method string,
	path string,
	query url.Values,
	body interface{},
) (*http.Response, error) {
	b := new(bytes.Buffer)
	if body != nil {
		if err := json.NewEncoder(b).Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := c.newRetryableRequest(method, path, query, b)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return newRetryableClient(&http.Client{
		Timeout:       30 * time.Second,
		CheckRedirect: copyRedirectHeader,
	}).Do(req.WithContext(ctx))
}

func (c *Client) newRetryableRequest(
	method string,
	path string,
	query url.Values,
	body interface{},
) (*retryable.Request, error) {
	u := c.baseURL.ResolveReference(&url.URL{Path: path, RawQuery: query.Encode()})
	req, err := retryable.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set(api.HeaderVersion, version)
	if len(c.userToken) > 0 {
		req.Header.Set("Authorization", "Bearer "+c.userToken)
	}

	return req, nil
}

func (c *Client) newRequest(
	method string,
	path string,
	query url.Values,
	body io.Reader,
) (*http.Request, error) {
	u := c.baseURL.ResolveReference(&url.URL{Path: path, RawQuery: query.Encode()})
	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set(api.HeaderVersion, version)
	if len(c.userToken) > 0 {
		req.Header.Set("Authorization", "Bearer "+c.userToken)
	}

	return req, nil
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
		return errors.Errorf("failed to parse response: %s", string(bytes))
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

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

// exponentialJitterBackoff implements exponential backoff with full jitter as described here:
// https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/
func exponentialJitterBackoff(
	minDuration, maxDuration time.Duration,
	attempt int,
	resp *http.Response,
) time.Duration {
	min := float64(minDuration)
	max := float64(maxDuration)

	backoff := min + math.Min(max-min, min*math.Exp2(float64(attempt)))*random.Float64()
	return time.Duration(backoff)
}
