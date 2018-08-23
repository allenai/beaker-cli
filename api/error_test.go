package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorFormat(t *testing.T) {
	apiErr := Error{Code: 42, Message: "whoops!", Stack: "Some\nstack\nlines"}
	assert.Equal(t, "whoops!", fmt.Sprintf("%v", apiErr))
	assert.Equal(t, "whoops!", fmt.Sprintf("%s", apiErr))
	assert.Equal(t, "\"whoops!\"", fmt.Sprintf("%q", apiErr))
	assert.Equal(t, "whoops!\nSome\nstack\nlines", fmt.Sprintf("%+v", apiErr))
}
