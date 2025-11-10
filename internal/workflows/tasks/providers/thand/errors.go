package thand

import (
	"errors"
	"fmt"

	"go.temporal.io/sdk/temporal"
)

// unwrapTemporalError unwraps ActivityError and extracts the underlying error message.
// It handles the Temporal error chain: ActivityError → ApplicationError → cause
// Returns the extracted error or the original error if no special handling is needed.
func unwrapTemporalError(err error) error {
	if err == nil {
		return nil
	}

	var foundError error

	// First unwrap ActivityError if present, then check the underlying error type
	var activityErr *temporal.ActivityError
	unwrappedErr := err
	if errors.As(err, &activityErr) {
		if innerErr := errors.Unwrap(activityErr); innerErr != nil {
			unwrappedErr = innerErr
		}
	}

	// Handle different Temporal error types
	switch {
	case temporal.IsApplicationError(unwrappedErr):
		var appErr *temporal.ApplicationError
		if errors.As(unwrappedErr, &appErr) {
			foundError = fmt.Errorf("application error: %v", appErr.Message())
		}
	default:
		foundError = err
	}

	return foundError
}
