package workspace

import (
	"errors"

	"google.golang.org/api/googleapi"
)

// HTTPStatus extracts the HTTP status code from a Workspace API error chain.
// Returns 0 if the error did not originate from Google's API.
func HTTPStatus(err error) int {
	var gerr *googleapi.Error
	if errors.As(err, &gerr) {
		return gerr.Code
	}
	return 0
}
