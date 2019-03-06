package api

import (
	"path"
	"time"
)

// DatasetStorage is a reference to a FileHeap dataset.
type DatasetStorage struct {
	Address string `json:"address,omitempty"`
	ID      string `json:"id,omitempty"`
	Token   string `json:"token,omitempty"`
}

// CreateDatasetResponse is a service response returned when a new dataset is created.
type CreateDatasetResponse struct {
	Storage DatasetStorage `json:"storage,omitempty"`

	ID string `json:"id"`
}

// Dataset is a file or collection of files. It may be the result of a task or
// uploaded directly by a user.
type Dataset struct {
	Storage DatasetStorage `json:"storage,omitempty"`

	// Identity
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`

	// Ownership
	Owner  Identity `json:"owner"`
	Author Identity `json:"author"`
	User   Identity `json:"user"` // TODO: Deprecated.

	// Status
	Created   time.Time `json:"created"`
	Committed time.Time `json:"committed,omitempty"`

	// A plain-text description of this dataset.
	Description string `json:"description,omitempty"`

	// Task for which this dataset is a result, i.e. provenance, if any.
	SourceTask *string `json:"source_task,omitempty"`

	// Included if the dataset is a single file.
	IsFile bool `json:"is_file,omitempty"`
}

// DisplayID returns the most human-friendly name available for a dataset while
// guaranteeing that it's unique and non-empty.
func (ds *Dataset) DisplayID() string {
	if ds.Name != "" {
		return path.Join(ds.User.Name, ds.Name)
	}
	return ds.ID
}

// DatasetSpec is a specification for creating a new Dataset.
type DatasetSpec struct {
	// (optional) Organization on behalf of whom this resource is created. The
	// user issuing the request must be a member of the organization. If omitted,
	// the resource will be owned by the requestor.
	Organization string `json:"org,omitempty"`

	// (optional) Text description for the dataset.
	Description string `json:"description,omitempty"`

	// (optional) If set, the dataset will be treated as a single file with the
	// given file name. Beaker will also enforce that the dataset contains at
	// most one file.
	Filename string `json:"filename,omitempty"`

	// (optional) A token representing the user to which the object should be attributed.
	// If omitted attribution will be given to the user issuing the request.
	AuthorToken string `json:"author_token,omitempty"`

	// (optional) If set, the dataset will be stored in FileHeap.
	// This flag will eventually become the default and be removed.
	FileHeap bool `json:"fileHeap,omitempty"`
}

// DatasetFile dsecribes a file within a dataset.
type DatasetFile struct {
	// The full path/name of the file from the root of the dataset.
	File string `json:"file"`

	// The size of the file, in bytes.
	Size uint64 `json:"size"`

	TimeLastModified time.Time `json:"time_last_modified"`
}

// DatasetFileLink represents a pre-signed upload or download link to a single file.
type DatasetFileLink struct {
	ID       string `json:"dataset_id"`
	FilePath string `json:"file_path"`
	URL      string `json:"url"`
}

// DatasetManifest describes the file contents of a dataset.
type DatasetManifest struct {
	// The unique ID of the dataset.
	ID string `json:"id"`

	// Whether the dataset should be treated as a single file.
	SingleFile bool `json:"single_file,omitempty"`

	// Descriptions of files contained in the dataset.
	Files []DatasetFile `json:"files,omitempty"`
}

// DatasetUsage describes how many experiments were started using this dataset
// as a source, including some metadata about when the usage occurred.
// Intended to be used in an aggregate statistic reporting.
type DatasetUsage struct {
	DatasetID          string    `json:"id"`
	EarliestUsage      time.Time `json:"earliest_usage"`
	LatestUsage        time.Time `json:"latest_usage"`
	DatasetNames       []string  `json:"dataset_names"`
	DatasetCreator     string    `json:"dataset_creator"`
	ExperimentCreators []string  `json:"experiment_creators"`
	ExperimentCount    int64     `json:"experiment_count"`
}

// DatasetPatchSpec describes a patch to apply to a dataset's editable fields.
// Only one field may be set in a single request.
type DatasetPatchSpec struct {
	// (optional) Unqualified name to assign to the dataset. It is considered
	// a collision error if another dataset has the same creator and name.
	Name *string `json:"name,omitempty"`

	// (optional) Description to assign to the dataset or empty string to
	// delete an existing description.
	Description *string `json:"description,omitempty"`

	// (optional) Whether the dataset should be locked for writes. Ignored if false.
	Commit bool `json:"commit,omitempty"`
}
