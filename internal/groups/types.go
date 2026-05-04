// Package groups implements the Workspace Groups domain (groups + members).
package groups

// Group is the Goldy projection of a Workspace group.
type Group struct {
	ID                 string `json:"id"`
	Email              string `json:"email"`
	Name               string `json:"name,omitempty"`
	Description        string `json:"description,omitempty"`
	DirectMembersCount int64  `json:"direct_members_count,omitempty"`
}

// Member is a single group membership.
type Member struct {
	ID    string `json:"id,omitempty"`
	Email string `json:"email"`
	Role  string `json:"role,omitempty"` // OWNER | MANAGER | MEMBER
	Type  string `json:"type,omitempty"` // USER | GROUP | EXTERNAL
}

// CreateGroupRequest is the JSON body for POST /api/v1/groups.
type CreateGroupRequest struct {
	Email       string `json:"email"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

func (r CreateGroupRequest) Validate() error {
	if r.Email == "" {
		return ErrInvalid("email is required")
	}
	return nil
}

// AddMemberRequest is the JSON body for POST /api/v1/groups/{id}/members.
type AddMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role,omitempty"`
}

func (r AddMemberRequest) Validate() error {
	if r.Email == "" {
		return ErrInvalid("email is required")
	}
	return nil
}

// ListGroupsResponse is the body returned by GET /api/v1/groups.
type ListGroupsResponse struct {
	Groups        []Group `json:"groups"`
	NextPageToken string  `json:"next_page_token,omitempty"`
}

// ListMembersResponse is the body returned by GET /api/v1/groups/{id}/members.
type ListMembersResponse struct {
	Members       []Member `json:"members"`
	NextPageToken string   `json:"next_page_token,omitempty"`
}
