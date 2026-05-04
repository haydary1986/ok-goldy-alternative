// Package users implements the Workspace Users domain: types, repository,
// service, and HTTP handlers.
package users

import "time"

// User is the Goldy-internal projection of a Workspace user.
type User struct {
	ID            string    `json:"id"`
	PrimaryEmail  string    `json:"primary_email"`
	GivenName     string    `json:"given_name,omitempty"`
	FamilyName    string    `json:"family_name,omitempty"`
	OrgUnitPath   string    `json:"org_unit_path,omitempty"`
	Suspended     bool      `json:"suspended"`
	IsAdmin       bool      `json:"is_admin"`
	CreationTime  time.Time `json:"creation_time,omitempty"`
	LastLoginTime time.Time `json:"last_login_time,omitempty"`
}

// CreateRequest is the JSON body for POST /api/v1/users.
type CreateRequest struct {
	PrimaryEmail string `json:"primary_email"`
	GivenName    string `json:"given_name"`
	FamilyName   string `json:"family_name"`
	Password     string `json:"password"`
	OrgUnitPath  string `json:"org_unit_path,omitempty"`
}

// Validate enforces minimum field requirements for a create request.
func (r CreateRequest) Validate() error {
	if r.PrimaryEmail == "" {
		return ErrInvalid("primary_email is required")
	}
	if r.GivenName == "" {
		return ErrInvalid("given_name is required")
	}
	if r.FamilyName == "" {
		return ErrInvalid("family_name is required")
	}
	if len(r.Password) < 8 {
		return ErrInvalid("password must be at least 8 characters")
	}
	return nil
}

// UpdateRequest is the JSON body for PATCH /api/v1/users/{id}.
// Pointer fields make absence distinguishable from zero values.
type UpdateRequest struct {
	GivenName   *string `json:"given_name,omitempty"`
	FamilyName  *string `json:"family_name,omitempty"`
	OrgUnitPath *string `json:"org_unit_path,omitempty"`
	Suspended   *bool   `json:"suspended,omitempty"`
}

// ListResponse is the body returned by GET /api/v1/users.
type ListResponse struct {
	Users         []User `json:"users"`
	NextPageToken string `json:"next_page_token,omitempty"`
}
