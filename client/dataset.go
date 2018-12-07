package client

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	fileheap "github.com/allenai/fileheap-client/client"
	"github.com/pkg/errors"

	"github.com/allenai/beaker/api"
)

// DatasetHandle provides operations on a dataset.
type DatasetHandle struct {
	client *Client
	id     string
	pkg    *fileheap.PackageRef
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

	var pkg *fileheap.PackageRef
	if body.PackageAddress != "" && body.PackageID != "" {
		fileheap, err := fileheap.New(body.PackageAddress)
		if err != nil {
			return nil, err
		}
		pkg = fileheap.Package(body.PackageID)
	}

	return &DatasetHandle{client: c, id: body.ID, pkg: pkg}, nil
}

// Dataset gets a handle for a dataset by name or ID. The returned handle is
// guaranteed throughout its lifetime to refer to the same object, even if that
// object is later renamed.
func (c *Client) Dataset(ctx context.Context, reference string) (*DatasetHandle, error) {
	canonicalRef, err := c.canonicalizeRef(ctx, reference)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendRequest(ctx, http.MethodGet, path.Join("/api/v3/datasets", canonicalRef), nil, nil)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp.Body)

	var body api.Dataset
	if err := parseResponse(resp, &body); err != nil {
		return nil, err
	}

	var pkg *fileheap.PackageRef
	if body.PackageAddress != "" && body.PackageID != "" {
		fileheap, err := fileheap.New(body.PackageAddress)
		if err != nil {
			return nil, err
		}
		pkg = fileheap.Package(body.PackageID)
	}

	return &DatasetHandle{client: c, id: body.ID, pkg: pkg}, nil
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
// Deprecated. Use Files() instead.
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

// FileIterator is an iterator over files within a dataset.
type FileIterator interface {
	Next() (*FileHandle, *FileInfo, error)
}

// Files returns an iterator over all files in the dataset under the given path.
func (h *DatasetHandle) Files(ctx context.Context, path string) (FileIterator, error) {
	if h.pkg != nil {
		return &packageFileIterator{
			dataset:  h,
			iterator: h.pkg.Files(ctx, path),
		}, nil
	}

	manifest, err := h.Manifest(ctx)
	if err != nil {
		return nil, err
	}

	return &manifestFileIterator{
		dataset: h,
		files:   manifest.Files,
	}, nil
}

// FileInfo describes a single file within a dataset.
type FileInfo struct {
	// Path of the file relative to its dataset root.
	Path string `json:"path"`

	// Size of the file in bytes.
	Size int64 `json:"size"`

	// Time at which the file was last updated.
	Updated time.Time `json:"updated"`
}

// ErrDone indicates an iterator is expended.
var ErrDone = errors.New("no more items in iterator")

// packageFileIterator is an iterator over files within a FileHeap package.
type packageFileIterator struct {
	dataset  *DatasetHandle
	iterator *fileheap.FileIterator
}

func (i *packageFileIterator) Next() (*FileHandle, *FileInfo, error) {
	ref, info, err := i.iterator.Next()
	if err == fileheap.ErrDone {
		return nil, nil, ErrDone
	}
	if err != nil {
		return nil, nil, err
	}
	return &FileHandle{
			dataset: i.dataset,
			file:    info.Path,
			fileRef: ref,
		}, &FileInfo{
			Path:    info.Path,
			Size:    info.Size,
			Updated: info.Updated,
		}, nil
}

// manifestFileIterator is an iterator over files in a dataset manifest.
type manifestFileIterator struct {
	dataset *DatasetHandle
	files   []api.DatasetFile
	prefix  string
}

func (i *manifestFileIterator) Next() (*FileHandle, *FileInfo, error) {
	for len(i.files) > 0 && !strings.HasPrefix(i.files[0].File, i.prefix) {
		i.files = i.files[1:]
	}
	if len(i.files) == 0 {
		return nil, nil, ErrDone
	}

	file := i.files[0]
	i.files = i.files[1:]

	return &FileHandle{
			dataset: i.dataset,
			file:    file.File,
		}, &FileInfo{
			Path:    file.File,
			Size:    int64(file.Size),
			Updated: file.TimeLastModified,
		}, nil
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
