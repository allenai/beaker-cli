package api

import (
	"time"
)

// Task is a full description of a task specification and its status.
type Task struct {
	// Identity
	ID   string `json:"id"`
	User User   `json:"user"`

	// Status
	Status  TaskStatus `json:"status"`
	Created time.Time  `json:"created"`
	Started time.Time  `json:"started"`
	Ended   time.Time  `json:"ended"`

	// Creation parameters
	Spec TaskSpec `json:"spec"`

	// Results
	ResultID string `json:"result_id"`
}

type TaskLogUploadLink struct {
	TaskID      string `json:"task_id"`
	TaskAttempt string `json:"task_attempt"`
	LogChunk    string `json:"log_chunk"`
	URL         string `json:"url"`
}

type TaskResults struct {
	Metrics map[string]interface{} `json:"metrics"`
}

// TaskSpec contains all information necessary to create a new task.
type TaskSpec struct {
	// (required) Blueprint describing the code to be run.
	Blueprint string `json:"blueprint"`

	// (required) Container path in which the task will save results. Files
	// written to this location will be persisted as a dataset upon task
	// completion.
	ResultPath string `json:"result_path"`

	// (optional) Text description of the task.
	Description string `json:"desc"` // TODO: Rename to "description"

	// (optional) Command-line arguments to pass to the task's container.
	Arguments []string `json:"arguments"`

	// (optional) Environment variables to pass into the task's container.
	Env map[string]string `json:"env"`

	// TODO: Replace both mount lists with TaskMount.

	// (optional) Data sources to mount as read-only in the task's container.
	// In the event that mounts overlap partially or in full, they will be
	// applied in order. Later mounts will overlay earlier ones (last wins).
	Mounts []DatasetMount `json:"sources"` // TODO: Rename to "mounts"

	// (optional) Task resource requirements for scheduling.
	Requirements TaskRequirements `json:"requirements"`
}

// TaskRequirements describes the runtime hardware requirements for a task.
type TaskRequirements struct {
	// (optional) Minimum required memory, in bytes.
	Memory int64 `json:"memory"`

	// (optional) Minimum CPUs to allocate in millicpus (1 CPU = 1000 millicpus).
	MilliCPU int `json:"cpu"`

	// (optional) GPUs required in increments of one full core.
	GPUCount int `json:"gpu_count"`
}

// DatasetMount describes a read-only data source for a task.
type DatasetMount struct {
	// (required) Name or Unique ID of a dataset to mount.
	DatasetID string `json:"dataset_id"` // TODO: Make this "dataset" which can be name or ID.

	// (required) Path within a task container to which file(s) will be mounted.
	ContainerPath string `json:"container_path"`
}

// TaskPatchSpec describes a patch to apply to a task's editable fields. Only
// one field may be set in a single request.
type TaskPatchSpec struct {
	// (optional) Description to assign to the task or empty string to delete an
	// existing description.
	Description *string `json:"description,omitempty"`

	// (optional) Whether the task should be canceled. Ignored if false.
	Cancel bool `json:"cancel,omitempty"`
}
