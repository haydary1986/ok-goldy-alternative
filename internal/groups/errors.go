package groups

import "errors"

// ErrWorkspaceUnavailable is returned when no Workspace client is configured.
var ErrWorkspaceUnavailable = errors.New("groups: workspace client is not configured")

type invalidError string

func (e invalidError) Error() string { return string(e) }

// ErrInvalid wraps a string into a validation error.
func ErrInvalid(msg string) error { return invalidError(msg) }

// IsInvalid reports whether the error came from request validation.
func IsInvalid(err error) bool {
	if err == nil {
		return false
	}
	var ie invalidError
	return errors.As(err, &ie)
}
