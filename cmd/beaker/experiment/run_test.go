package experiment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ExperimentURL(t *testing.T) {
	assert.Equal(t, "http://somewhere:12345/ex/ex_abc123", experimentURL("http://somewhere:12345", "ex_abc123"))
}
