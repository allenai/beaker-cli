package api

import (
	"fmt"
)

// Error encodes an error as a JSON-serializable struct.
type Error struct {
	// HTTP status code, such as 404
	Code int `json:"code"`

	// The text of the error, which should follow the guidelines found at:
	// https://github.com/golang/go/wiki/CodeReviewComments#error-strings
	Message string `json:"message"`

	// (optional) The error's call stack as a formatted string.
	Stack string `json:"stack,omitempty"`

	// (optional) An identifier for this error for tracing purposes
	ErrorID string `json:"error_id,omitempty"`
}

// Error implements the standard error interface.
func (e Error) Error() string {
	return e.Message
}

// Format implements the fmt.Formatter interface.
func (e Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%s\n%s", e.messageWithErrorID(), e.Stack)
			return
		}
		fallthrough
	case 's':
		fmt.Fprint(s, e.messageWithErrorID())
	case 'q':
		fmt.Fprintf(s, "%q", e.messageWithErrorID())
	}
}

func (e Error) messageWithErrorID() string {
	if len(e.ErrorID) > 0 {
		return fmt.Sprintf("%s (error_id %s)", e.Message, e.ErrorID)
	}
	return e.Message
}
