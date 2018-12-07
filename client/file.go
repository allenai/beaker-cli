package client

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"

	fileheap "github.com/allenai/fileheap-client/client"
	"github.com/pkg/errors"
)

// FileHandle provides operations on a file within a dataset.
type FileHandle struct {
	dataset *DatasetHandle
	file    string
	fileRef *fileheap.FileRef
}

// FileRef creates an actor for an existing file within a dataset.
// This call doesn't perform any network operations.
func (h *DatasetHandle) FileRef(filePath string) *FileHandle {
	return &FileHandle{h, filePath, h.pkg.File(filePath)}
}

// Download gets a file from a datastore.
func (h *FileHandle) Download(ctx context.Context) (io.ReadCloser, error) {
	if h.fileRef != nil {
		return h.fileRef.NewReader(ctx)
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
	if h.fileRef != nil {
		return h.fileRef.NewRangeReader(ctx, offset, length)
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
func (h *FileHandle) Upload(ctx context.Context, source io.ReadSeeker) error {
	hasher := sha256.New()
	length, err := io.Copy(hasher, source)
	if err != nil {
		return errors.Wrap(err, "failed to hash contents")
	}

	digest := hasher.Sum(nil)

	if _, err := source.Seek(0, 0); err != nil {
		return errors.WithStack(err)
	}

	// Only read as many bytes as were hashed.
	body := io.LimitReader(source, length)

	if h.fileRef != nil {
		w, err := h.fileRef.NewWriter(ctx, &fileheap.WriteOpts{
			Length: length,
			Digest: digest,
		})
		if err != nil {
			return errors.WithStack(err)
		}

		if _, err = io.Copy(w, body); err != nil {
			return errors.WithStack(err)
		}

		return errors.WithStack(w.Close())
	}

	path := path.Join("/api/v3/datasets", h.dataset.id, "files", h.file)
	req, err := h.dataset.client.newRequest(http.MethodPut, path, nil, body)
	if err != nil {
		return err
	}
	req.ContentLength = length
	req.Header.Set("Digest", "SHA256 "+base64.StdEncoding.EncodeToString(digest))

	client := &http.Client{CheckRedirect: copyRedirectHeader}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return errors.WithStack(err)
	}
	return errorFromResponse(resp)
}

// Delete removes a file from an uncommitted datastore.
func (h *FileHandle) Delete(ctx context.Context) error {
	if h.fileRef != nil {
		return h.fileRef.Delete(ctx)
	}

	path := path.Join("/api/v3/datasets", h.dataset.id, "files", h.file)
	resp, err := h.dataset.client.sendRequest(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}
