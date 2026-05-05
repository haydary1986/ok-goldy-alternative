// Package orgunits exposes Workspace organizational units (OUs) so the
// admin UI can render a dropdown and let the operator create new ones
// without leaving Goldy.
package orgunits

import "errors"

// OrgUnit is the Goldy projection of a Workspace organizational unit.
type OrgUnit struct {
	OrgUnitID         string `json:"org_unit_id"`
	Name              string `json:"name"`
	OrgUnitPath       string `json:"org_unit_path"`
	ParentOrgUnitPath string `json:"parent_org_unit_path,omitempty"`
	Description       string `json:"description,omitempty"`
}

// ListResponse is the body returned by GET /api/v1/orgunits.
type ListResponse struct {
	OrgUnits []OrgUnit `json:"org_units"`
}

// CreateRequest is the JSON body for POST /api/v1/orgunits.
type CreateRequest struct {
	Name              string `json:"name"`
	ParentOrgUnitPath string `json:"parent_org_unit_path"`
	Description       string `json:"description,omitempty"`
}

func (r CreateRequest) Validate() error {
	if r.Name == "" {
		return ErrInvalid("name is required")
	}
	if r.ParentOrgUnitPath == "" {
		return ErrInvalid("parent_org_unit_path is required (use \"/\" for root)")
	}
	return nil
}

// ErrWorkspaceUnavailable is returned when no Workspace client is configured.
var ErrWorkspaceUnavailable = errors.New("orgunits: workspace client is not configured")

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
