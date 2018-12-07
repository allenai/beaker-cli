package client

import (
	"strings"
	"time"

	fileheap "github.com/allenai/fileheap-client/client"
	"github.com/pkg/errors"

	"github.com/allenai/beaker/api"
)

// FileIterator is an iterator over files within a dataset.
type FileIterator interface {
	Next() (*FileHandle, *FileInfo, error)
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
