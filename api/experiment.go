package api

import (
	"path"
	"time"
)

// Experiment describes an experiment and its tasks.
type Experiment struct {
	// Identity
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`

	// Ownership
	Owner  Identity `json:"owner"`
	Author Identity `json:"author"`
	User   Identity `json:"user"` // TODO: Deprecated.

	Description string           `json:"description,omitempty"`
	Nodes       []ExperimentNode `json:"nodes"`
	Created     time.Time        `json:"created"`
	Archived    bool             `json:"archived"`
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
	// (optional) Organization on behalf of whom this resource is created. The
	// user issuing the request must be a member of the organization. If omitted,
	// the resource will be owned by the requestor.
	Organization string `json:"org,omitempty" yaml:"org,omitempty"`

	// (optional) Text description of the experiment.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// (required) Tasks to create. Tasks may be defined in any order, though all
	// dependencies must be internally resolvable within the experiment.
	Tasks []ExperimentTaskSpec `json:"tasks" yaml:"tasks"`

	// (optional) A token representing the user to which the object should be attributed.
	// If omitted attribution will be given to the user issuing the request.
	AuthorTokenDeprecated string `json:"author_token,omitempty" yaml:"-"`
	AuthorToken           string `json:"authorToken,omitempty" yaml:"authorToken,omitempty"`

	// (optional) Settings for the Comet.ml integration, if it should be used for this experiment.
	Comet *ExperimentCometSpec `json:"comet,omitempty" yaml:"comet,omitempty"`
}

// ExperimentNode describes a task along with its links within an experiment.
type ExperimentNode struct {
	Name               string     `json:"name,omitempty"`
	TaskIDDeprecated   string     `json:"task_id"`
	TaskID             string     `json:"taskId"`
	ResultIDDeprecated string     `json:"result_id"`
	ResultID           string     `json:"resultId"`
	Status             TaskStatus `json:"status"`
	CometURL           string     `json:"cometUrl,omitempty"`

	// Identifiers of tasks dependent on this node within the containing experiment.
	ChildTasksDeprecated []string `json:"child_task_ids"`
	ChildTasks           []string `json:"childTaskIds"`

	// Identifiers of task on which this node depends within the containing experiment.
	ParentTasksDeprecated []string `json:"parent_task_ids"`
	ParentTasks           []string `json:"parentTaskIds"`
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
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// (required) Specification describing the task to run.
	Spec TaskSpec `json:"spec" yaml:"spec,omitempty"`

	// (optional) Tasks on which this task depends. Mounts will be applied, in
	// the order defined here, after existing mounts in the task spec.
	DependsOnDeprecated []TaskDependency `json:"depends_on,omitempty" yaml:"-"`
	DependsOn           []TaskDependency `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
}

// TaskDependency describes a single "edge" in a task dependency graph.
type TaskDependency struct {
	// (required) Name of the task on which the referencing task depends.
	ParentNameDeprecated string `json:"parent_name" yaml:"-"`
	ParentName           string `json:"parentName" yaml:"parentName"`

	// (optional) Path in the child task to which parent results will be mounted.
	// If absent, this is treated as an order-only dependency.
	ContainerPathDeprecated string `json:"container_path,omitempty" yaml:"-"`
	ContainerPath           string `json:"containerPath,omitempty" yaml:"containerPath,omitempty"`
}

type ExperimentCometSpec struct {
	// (required) Whether or not to enable the integration for this experiment.
	Enable bool `json:"enable"`

	// (optional) The name of the experiment (shown in the Comet.ml interface)
	ExperimentName string `json:"experiment,omitempty"`

	// (optional) The name of the Comet.ml project for this experiment.
	ProjectName string `json:"project,omitempty"`

	// (optional) The name of the Comet.ml workspace for this experiment.
	Workspace string `json:"workspace,omitempty"`
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

	// (optional) Whether the experiment should be archived. Ignored if nil.
	Archive *bool `json:"archive,omitempty"`
}
