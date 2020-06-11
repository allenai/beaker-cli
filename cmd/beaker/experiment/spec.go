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

	return api.ExperimentSpec{Description: s.Description, Tasks: tasks}, nil
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
	DependsOn []api.TaskDependency `yaml:"dependsOn,omitempty"`

	// (optional) Name of a cluster on which the task should run.
	// Cluster affinity supercedes task requirements.
	Cluster string `yaml:"cluster,omitempty"`
}

// ToAPI converts to an API-compatible struct.
func (e ExperimentTaskSpec) ToAPI() (api.ExperimentTaskSpec, error) {
	spec, err := e.Spec.ToAPI()
	if err != nil {
		return api.ExperimentTaskSpec{}, err
	}

	return api.ExperimentTaskSpec{
		Name:      e.Name,
		Spec:      *spec,
		DependsOn: e.DependsOn,
		Cluster:   e.Cluster,
	}, nil
}

// TaskSpec contains all information necessary to create a new experiment on the host.
type TaskSpec struct {
	// (required) Image describing code to be run
	Image       string `yaml:"image,omitempty"`
	DockerImage string `yaml:"dockerImage,omitempty"`

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
	Mounts []api.DatasetMount `yaml:"datasetMounts,omitempty"`

	// (optional) Experiment resource requirements for scheduling.
	Requirements Requirements `yaml:"requirements,omitempty"`
}

// ToAPI converts to an API-compatible struct.
func (s *TaskSpec) ToAPI() (*api.TaskSpec, error) {
	requirements, err := s.Requirements.ToAPI()
	if err != nil {
		return nil, err
	}

	return &api.TaskSpec{
		Image:        s.Image,
		DockerImage:  s.DockerImage,
		ResultPath:   s.ResultPath,
		Description:  s.Description,
		Arguments:    s.Arguments,
		Env:          s.Env,
		Mounts:       s.Mounts,
		Requirements: requirements,
	}, nil
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
