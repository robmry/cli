package cli

import (
	"strconv"
	"strings"
)

// Errors is a list of errors.
// Useful in a loop if you don't want to return the error right away and you want to display after the loop,
// all the errors that happened during the loop.
type Errors []error

func (errList Errors) Error() string {
	if len(errList) < 1 {
		return ""
	}

	out := make([]string, len(errList))
	for i := range errList {
		out[i] = errList[i].Error()
	}
	return strings.Join(out, ", ")
}

// StatusError reports an unsuccessful exit by a command.
type StatusError struct {
	Status     string
	StatusCode int
}

// Error formats the error for printing. If a custom Status is provided,
// it is returned as-is, otherwise it generates a generic error-message
// based on the StatusCode.
func (e StatusError) Error() string {
	if e.Status == "" {
		return "exit status " + strconv.Itoa(e.StatusCode)
	}
	return e.Status
}
