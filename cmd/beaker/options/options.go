package options

import (
	"code.cloudfoundry.org/bytefmt"
	"github.com/allenai/beaker-api/api"
	"github.com/pkg/errors"
)

// AppOptions captures options relevant throughout the application.
type AppOptions struct {
	Debug bool
}

// MakeRequirements converts strings representing user-provided CPU and Memory requirements into
// a type.TaskRequirements struct.
func MakeRequirements(cpu float64, memory string, gpuCount int) (*api.TaskRequirements, error) {
	requirements := &api.TaskRequirements{}

	if cpu < 0 {
		return nil, errors.Errorf("couldn't parse cpu argument '%f' because it was negative", cpu)
	}

	if cpu > 0 {
		requirements.MilliCPU = int(cpu * 1000)
	}

	if len(memory) > 0 {
		memoryBytes, err := bytefmt.ToBytes(memory)
		if err != nil {
			return nil, errors.Errorf("couldn't parse memory argument '%s' because %v", memory, err)
		}
		requirements.Memory = int64(memoryBytes)
	}

	if gpuCount > 0 {
		requirements.GPUCount = gpuCount
	}

	return requirements, nil
}
