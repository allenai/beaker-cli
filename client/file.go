package client

import (
	"context"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"

	"github.com/allenai/beaker/api"
)

// FileHandle provides operations on a file within a dataset.
type FileHandle struct {
	dataset *DatasetHandle
	file    string
}

// FileRef creates an actor for an existing file within a dataset.
// This call doesn't perform any network operations.
func (h *DatasetHandle) FileRef(filePath string) *FileHandle {
	return &FileHandle{h, filePath}
}

// PresignLink creates a pre-signed URL link to a file.
func (h *FileHandle) PresignLink(ctx context.Context, forWrite bool) (*api.DatasetFileLink, error) {
	path := path.Join("/api/v3/datasets", h.dataset.id, "links", h.file)
	query := map[string]string{"upload": strconv.FormatBool(forWrite)}
	resp, err := h.dataset.client.sendRequest(ctx, http.MethodPost, path, query, nil)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp.Body)

	var body api.DatasetFileLink
	if err = parseResponse(resp, &body); err != nil {
		return nil, err
	}
	return &body, nil
}

// Download gets a file from a datastore.
func (h *FileHandle) Download(ctx context.Context) (io.ReadCloser, error) {
	link, err := h.PresignLink(ctx, false)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, link.URL, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	httpClient := http.Client{}
	resp, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return resp.Body, nil
}

// DownloadTo downloads a file and writes it to disk.
func (h *FileHandle) DownloadTo(ctx context.Context, filePath string) error {
	r, err := h.Download(ctx)
	if err != nil {
		return err
	}
	defer safeClose(r)

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return errors.WithStack(err)
	}

	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return errors.WithStack(err)
	}
	defer safeClose(f)

	_, err = io.Copy(f, r)
	return errors.WithStack(err)
}

// Delete removes a file from an uncommitted datastore.
func (h *FileHandle) Delete(ctx context.Context) error {
	path := path.Join("/api/v3/datasets", h.dataset.id, "files", h.file)
	resp, err := h.dataset.client.sendRequest(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}
