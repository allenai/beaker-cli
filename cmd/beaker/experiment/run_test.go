package experiment

import (
	"testing"

	"github.com/allenai/beaker-api/api"
	"github.com/stretchr/testify/assert"

	"github.com/allenai/beaker/cmd/beaker/options"
)

func Test_MakeRequirements_NoValues(t *testing.T) {
	requirements, err := options.MakeRequirements(0, "", 0)
	assert.NoError(t, err)

	expectedReqs := &api.TaskRequirements{}
	assert.Equal(t, expectedReqs, requirements)
}

func Test_MakeRequirements_HappyValues(t *testing.T) {
	requirements, err := options.MakeRequirements(0.5, "1GB", 1)
	assert.NoError(t, err)

	expectedReqs := &api.TaskRequirements{MilliCPU: 500, Memory: 1073741824, GPUCount: 1}
	assert.Equal(t, expectedReqs, requirements)
}

func Test_MakeRequirements_BadMemoryValue(t *testing.T) {
	requirements, err := options.MakeRequirements(0.5, "1BeakerByte", 0)
	assert.Error(t, err)
	assert.Nil(t, requirements)
}

func Test_MakeRequirements_BadCpuValue(t *testing.T) {
	requirements, err := options.MakeRequirements(-4, "1GB", 0)
	assert.Error(t, err)
	assert.Nil(t, requirements)
}

func Test_ExperimentURL(t *testing.T) {
	assert.Equal(t, "http://somewhere:12345/ex/ex_abc123", experimentURL("http://somewhere:12345", "ex_abc123"))
}
