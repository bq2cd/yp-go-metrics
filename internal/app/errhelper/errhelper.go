package errhelper

import (
	"errors"
)

// UnwrapJoined recursively unwraps given error and returns flattened
// slice of errors.
func UnwrapJoined(errInitial error) []error {
	var errJoined interface{ Unwrap() []error }
	ok := errors.As(errInitial, &errJoined)
	if !ok {
		return []error{errInitial}
	}
	unwrapped := make([]error, 0)
	for _, err := range errJoined.Unwrap() {
		unwrapped = append(unwrapped, UnwrapJoined(err)...)
	}
	return unwrapped
}
