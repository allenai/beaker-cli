package experiment

import (
	"code.cloudfoundry.org/bytefmt"
	"github.com/pkg/errors"

	"github.com/beaker/client/api"
)

// ExperimentSpec describes a set of tasks with optional dependencies.
// This set represents a (potentially disconnected) directed acyclic graph.
type ExperimentSpec struct {
	// (optional) Text description of the experiment.
	Description string `yaml:"description,omitempty"`

	// (optional) If true, create Comet.ml experiments per task in this experiment.
	Comet *ExperimentCometSpec `yaml:"comet,omitempty"`

	// (required) Tasks to create. Tasks may be defined in any order, though all
	// dependencies must be internally resolvable within the experiment.
	Tasks []ExperimentTaskSpec `yaml:"tasks"`
}

// ToAPI converts to an API-compatible struct.
func (s ExperimentSpec) ToAPI() (api.ExperimentSpec, error) {
	var tasks []api.ExperimentTaskSpec
	for _, task := range s.Tasks {
		apiTask, err := task.ToAPI()
		if err != nil {
			return api.ExperimentSpec{}, err
		}
		tasks = append(tasks, apiTask)
	}

	apiSpec := api.ExperimentSpec{Description: s.Description, Tasks: tasks}
	if s.Comet != nil {
		apiSpec.Comet = s.Comet.ToAPI()
	}
	return apiSpec, nil
}

type ExperimentCometSpec struct {
	Enable         bool   `yaml:"enable"`
	ExperimentName string `yaml:"experiment,omitempty"`
	ProjectName    string `yaml:"project,omitempty"`
	Workspace      string `yaml:"workspace,omitempty"`
}

// ToAPI converts to an API-compatible struct.
func (s ExperimentCometSpec) ToAPI() *api.ExperimentCometSpec {
	return &api.ExperimentCometSpec{
		Enable:         s.Enable,
		ExperimentName: s.ExperimentName,
		ProjectName:    s.ProjectName,
		Workspace:      s.Workspace,
	}
}

// ExperimentTaskSpec describes a task spec with optional dependencies on other
// tasks within an experiment. Tasks refer to each other by the Name field.
type ExperimentTaskSpec struct {
	// (optional) Name of the task node, which need only be defined if
	// dependencies reference it.
	Name string `yaml:"name,omitempty"`

	// (required) Specification describing the task to run.
	Spec TaskSpec `yaml:"spec"`

	// (optional) Tasks on which this task depends. Mounts will be applied, in
	// the order defined here, after existing mounts in the task spec.
	DependsOn []TaskDependency `yaml:"dependsOn,omitempty"`
}

// ToAPI converts to an API-compatible struct.
func (e ExperimentTaskSpec) ToAPI() (api.ExperimentTaskSpec, error) {
	spec, err := e.Spec.ToAPI()
	if err != nil {
		return api.ExperimentTaskSpec{}, err
	}

	var deps []api.TaskDependency
	for _, dep := range e.DependsOn {
		deps = append(deps, dep.ToAPI())
	}

	return api.ExperimentTaskSpec{Name: e.Name, Spec: *spec, DependsOn: deps}, nil
}

// TaskDependency describes a single "edge" in a task dependency graph.
type TaskDependency struct {
	// (required) Name of the task on which the referencing task depends.
	ParentName string `yaml:"parentName"`

	// (optional) Path in the child task to which parent results will be mounted.
	// If absent, this is treated as an order-only dependency.
	ContainerPath string `yaml:"containerPath,omitempty"`
}

// ToAPI converts to an API-compatible struct.
func (d TaskDependency) ToAPI() api.TaskDependency {
	return api.TaskDependency{ParentName: d.ParentName, ContainerPath: d.ContainerPath}
}

// TaskSpec contains all information necessary to create a new experiment on the host.
type TaskSpec struct {
	// (required) Blueprint describing the code to be run.
	Blueprint string `yaml:"blueprint"`

	// Image describing code to be run or name of the Docker image to run (deprecated).
	Image string `yaml:"image,omitempty"`

	// (required) Container path in which experiment will save results.
	// Files written to this location will be persisted as a dataset upon experiment completion.
	ResultPath string `yaml:"resultPath"`

	// (optional) Text description of the experiment.
	Description string `yaml:"description,omitempty"`

	// (optional) Command-line arguments to pass to the container.
	Arguments []string `yaml:"args,omitempty"`

	// (optional) Environment variables to pass into the container.
	Env map[string]string `yaml:"env,omitempty"`

	// (optional) Data sources to mount as read-only in the task's container.
	// In the event that mounts overlap partially or in full, they will be
	// applied in order. Later mounts will overlay earlier ones (last wins).
	Mounts []DatasetMount `yaml:"datasetMounts,omitempty"`

	// (optional) Experiment resource requirements for scheduling.
	Requirements Requirements `yaml:"requirements,omitempty"`
}

// ToAPI converts to an API-compatible struct.
func (s *TaskSpec) ToAPI() (*api.TaskSpec, error) {
	var datasetMounts []api.DatasetMount
	for _, mount := range s.Mounts {
		datasetMounts = append(datasetMounts, api.DatasetMount{
			Dataset:       mount.DatasetID,
			SubPath:       mount.SubPath,
			ContainerPath: mount.ContainerPath,
		})
	}

	requirements, err := s.Requirements.ToAPI()
	if err != nil {
		return nil, err
	}

	image := s.Image
	if image == "" {
		image = s.Blueprint
	}

	return &api.TaskSpec{
		Image:        image,
		ResultPath:   s.ResultPath,
		Description:  s.Description,
		Arguments:    s.Arguments,
		Env:          s.Env,
		Mounts:       datasetMounts,
		Requirements: requirements,
	}, nil
}

// DatasetMount describes a read-only source in the experiment container.
type DatasetMount struct {
	// (required) Unique ID of the dataset to mount.
	DatasetID string `yaml:"datasetId"`

	// (optional) Path within the dataset to mount for this experiment container.
	SubPath string `yaml:"subPath,omitempty"`

	// (required) Path within an experiment container to which this dataset will be mounted.
	ContainerPath string `yaml:"containerPath"`
}

// ToAPI converts to an API-compatible struct.
func (m DatasetMount) ToAPI() api.DatasetMount {
	return api.DatasetMount{Dataset: m.DatasetID, SubPath: m.SubPath, ContainerPath: m.ContainerPath}
}

// Requirements describes the runtime requirements for an experiment's container.
type Requirements struct {
	// (optional) Minimum CPUs to allocate as floating point.
	// CPU requirements are rounded to one thousandth of a CPU, i.e. 0.001
	CPU float64 `yaml:"cpu,omitempty"`

	// (optional) Minimum required memory, as a string which includes unit suffix.
	// Examples: "2g", "256m"
	Memory string `yaml:"memory,omitempty"`

	// (optional) GPUs required, in increments of one full core.
	GPUCount int `yaml:"gpuCount,omitempty"`

	// (optional) GPU variant to prefer when scheduling task.
	GPUType string `yaml:"gpuType,omitempty"`

	// (optional) Run on preemptible instances (defaults to false)
	Preemptible bool `json:"preemptible,omitempty" yaml:"preemptible,omitempty"`
}

// ToAPI converts to an API-compatible struct.
func (r Requirements) ToAPI() (api.TaskRequirements, error) {
	if r.CPU < 0 {
		return api.TaskRequirements{}, errors.Errorf("couldn't parse cpu argument '%.2f' because it was negative", r.CPU)
	}

	result := api.TaskRequirements{
		MilliCPU:    int(r.CPU * 1000),
		GPUCount:    r.GPUCount,
		GPUType:     r.GPUType,
		Preemptible: r.Preemptible,
	}

	if r.Memory != "" {
		bytes, err := bytefmt.ToBytes(r.Memory)
		if err != nil {
			return api.TaskRequirements{}, errors.Wrapf(err, "invalid memory value %q", r.Memory)
		}
		result.Memory = int64(bytes)
	}

	return result, nil
}
