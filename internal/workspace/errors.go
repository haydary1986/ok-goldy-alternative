package workspace

import (
	"errors"

	"google.golang.org/api/googleapi"
)

// ErrNoCredentials is returned by Provider.Variant when no Workspace
// credentials have been configured yet (no DB row, no env vars).
var ErrNoCredentials = errors.New("workspace: no credentials configured")

// HTTPStatus extracts the HTTP status code from a Workspace API error chain.
// Returns 0 if the error did not originate from Google's API.
func HTTPStatus(err error) int {
	var gerr *googleapi.Error
	if errors.As(err, &gerr) {
		return gerr.Code
	}
	return 0
}
