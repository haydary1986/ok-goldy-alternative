package users

import "errors"

// ErrWorkspaceUnavailable is returned when no Workspace client is configured.
var ErrWorkspaceUnavailable = errors.New("users: workspace client is not configured")

// invalidError signals a validation failure on a request payload.
type invalidError string

func (e invalidError) Error() string { return string(e) }

// ErrInvalid wraps a string into a validation error so callers can distinguish
// 400-style errors from infrastructure errors.
func ErrInvalid(msg string) error { return invalidError(msg) }

// IsInvalid reports whether the error came from request validation.
func IsInvalid(err error) bool {
	if err == nil {
		return false
	}
	var ie invalidError
	return errors.As(err, &ie)
}
