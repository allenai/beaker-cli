package experiment

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/allenai/beaker/api"
)

func TestRequirementsToAPI(t *testing.T) {
	cases := []struct {
		input       Requirements
		expected    api.TaskRequirements
		expectedErr string
	}{
		// Defaults
		{ /* all zeroes */ },

		// All values provided
		{
			input:    Requirements{1.5, "2m", 1},
			expected: api.TaskRequirements{MilliCPU: 1500, Memory: 2 * 1024 * 1024, GPUCount: 1},
		},

		// Variations on memory string
		{input: Requirements{Memory: "2mb"}, expected: api.TaskRequirements{Memory: 2 * 1024 * 1024}},
		{input: Requirements{Memory: "2MB"}, expected: api.TaskRequirements{Memory: 2 * 1024 * 1024}},
		{input: Requirements{Memory: "2048k"}, expected: api.TaskRequirements{Memory: 2 * 1024 * 1024}},

		// Bad memory strings
		{
			input:       Requirements{Memory: "-2mb"},
			expectedErr: `invalid memory value "-2mb": byte quantity must be a positive integer with a unit of measurement like M, MB, MiB, G, GiB, or GB`,
		},
		{
			input:       Requirements{Memory: "1BeakerByte"},
			expectedErr: `invalid memory value "1BeakerByte": byte quantity must be a positive integer with a unit of measurement like M, MB, MiB, G, GiB, or GB`,
		},
		{
			input:       Requirements{Memory: "g!bb3rish"},
			expectedErr: `invalid memory value "g!bb3rish": byte quantity must be a positive integer with a unit of measurement like M, MB, MiB, G, GiB, or GB`,
		},
	}

	for _, c := range cases {
		actual, err := c.input.ToAPI()
		assert.Equal(t, c.expected, actual)
		if c.expectedErr == "" {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, c.expectedErr)
		}
	}
}
