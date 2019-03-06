package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
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

// Download gets a file from a datastore.
func (h *FileHandle) Download(ctx context.Context) (io.ReadCloser, error) {
	if h.dataset.Storage != nil {
		return h.dataset.Storage.ReadFile(ctx, h.file)
	}

	path := path.Join("/api/v3/datasets", h.dataset.id, "files", h.file)
	req, err := h.dataset.client.newRetryableRequest(http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	client := newRetryableClient(&http.Client{CheckRedirect: copyRedirectHeader})
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if err := errorFromResponse(resp); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// DownloadRange reads a range of bytes from a file.
// If length is negative, the file is read until the end.
func (h *FileHandle) DownloadRange(ctx context.Context, offset, length int64) (io.ReadCloser, error) {
	if h.dataset.Storage != nil {
		return h.dataset.Storage.ReadFileRange(ctx, h.file, offset, length)
	}

	path := path.Join("/api/v3/datasets", h.dataset.id, "files", h.file)
	req, err := h.dataset.client.newRetryableRequest(http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	if length < 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset))
	} else {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", offset, offset+length-1))
	}

	client := newRetryableClient(&http.Client{CheckRedirect: copyRedirectHeader})
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if err := errorFromResponse(resp); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// DownloadTo downloads a file and writes it to disk.
func (h *FileHandle) DownloadTo(ctx context.Context, filePath string) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return errors.WithStack(err)
	}

	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return errors.WithStack(err)
	}
	defer safeClose(f)

	var written int64
	for {
		var r io.ReadCloser
		var err error
		if written == 0 {
			r, err = h.Download(ctx)
		} else {
			r, err = h.DownloadRange(ctx, written, -1)
		}
		if err != nil {
			return err
		}

		n, err := io.Copy(f, r)
		safeClose(r)
		if err == nil {
			return nil
		}
		written += n
	}
}

// Upload creates or overwrites a file.
func (h *FileHandle) Upload(ctx context.Context, source io.Reader, length int64) error {
	if h.dataset.Storage != nil {
		return h.dataset.Storage.WriteFile(ctx, h.file, source, length)
	}

	path := path.Join("/api/v3/datasets", h.dataset.id, "files", h.file)
	req, err := h.dataset.client.newRequest(http.MethodPut, path, nil, source)
	if err != nil {
		return err
	}
	req.ContentLength = length

	client := &http.Client{CheckRedirect: copyRedirectHeader}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return errors.WithStack(err)
	}
	return errorFromResponse(resp)
}

// Delete removes a file from an uncommitted datastore.
func (h *FileHandle) Delete(ctx context.Context) error {
	if h.dataset.Storage != nil {
		return h.dataset.Storage.DeleteFile(ctx, h.file)
	}

	path := path.Join("/api/v3/datasets", h.dataset.id, "files", h.file)
	resp, err := h.dataset.client.sendRequest(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}
