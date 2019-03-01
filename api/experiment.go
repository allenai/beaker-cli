package api

import (
	"path"
	"time"
)

// Experiment describes an experiment and its tasks.
type Experiment struct {
	ID          string           `json:"id"`
	User        User             `json:"user"`
	Name        string           `json:"name,omitempty"`
	Description string           `json:"description,omitempty"`
	Nodes       []ExperimentNode `json:"nodes"`
	Created     time.Time        `json:"created"`
}

// DisplayID returns the most human-friendly name available for an experiment
// while guaranteeing that it's unique and non-empty.
func (e *Experiment) DisplayID() string {
	if e.Name != "" {
		return path.Join(e.User.Name, e.Name)
	}
	return e.ID
}

// ExperimentSpec describes a set of tasks with optional dependencies.
// This set represents a (potentially disconnected) directed acyclic graph.
type ExperimentSpec struct {
	// (optional) Text description of the experiment.
	Description string `json:"description,omitempty"`

	// (required) Tasks to create. Tasks may be defined in any order, though all
	// dependencies must be internally resolvable within the experiment.
	Tasks []ExperimentTaskSpec `json:"tasks"`

	// (optional) A token representing the user to which the object should be attributed.
	// If omitted attribution will be given to the user issuing request.
	AuthorToken string `json:"author_token,omitempty"`

	// (optional) The name of the org to which the object should belong.  If omitted the object
	// will belong to the default org set in the config file.  If no default org is set, the
	// object will belong to the user issuing the request.
	Org string `json:"org,omitempty"`
}

// ExperimentNode describes a task along with its links within an experiment.
type ExperimentNode struct {
	Name     string     `json:"name,omitempty"`
	TaskID   string     `json:"task_id"`
	ResultID string     `json:"result_id"`
	Status   TaskStatus `json:"status"`

	// Identifiers of tasks dependent on this node within the containing experiment.
	ChildTasks []string `json:"child_task_ids"`

	// Identifiers of task on which this node depends within the containing experiment.
	ParentTasks []string `json:"parent_task_ids"`
}

// DisplayID returns the most human-friendly name available for an experiment
// node while guaranteeing that it's unique within the context of its experiment.
func (n *ExperimentNode) DisplayID() string {
	if n.Name != "" {
		return n.Name
	}
	return n.TaskID
}

// ExperimentTaskSpec describes a task spec with optional dependencies on other
// tasks within an experiment. Tasks refer to each other by the Name field.
type ExperimentTaskSpec struct {
	// (optional) Name of the task node, which need only be defined if
	// dependencies reference it.
	Name string `json:"name,omitempty"`

	// (required) Specification describing the task to run.
	Spec TaskSpec `json:"spec"`

	// (optional) Tasks on which this task depends. Mounts will be applied, in
	// the order defined here, after existing mounts in the task spec.
	DependsOn []TaskDependency `json:"depends_on,omitempty"`
}

// TaskDependency describes a single "edge" in a task dependency graph.
type TaskDependency struct {
	// (required) Name of the task on which the referencing task depends.
	ParentName string `json:"parent_name"`

	// (optional) Path in the child task to which parent results will be mounted.
	// If absent, this is treated as an order-only dependency.
	ContainerPath string `json:"container_path,omitempty"`
}

// ExperimentPatchSpec describes a patch to apply to an experiment's editable
// fields. Only one field may be set in a single request.
type ExperimentPatchSpec struct {
	// (optional) Unqualified name to assign to the experiment. It is considered
	// a collision error if another experiment has the same creator and name.
	Name *string `json:"name,omitempty"`

	// (optional) Description to assign to the experiment or empty string to
	// delete an existing description.
	Description *string `json:"description,omitempty"`
}
