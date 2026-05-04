// Package wsadmin owns the Workspace credentials lifecycle: storing the
// uploaded service-account JSON, hot-swapping the workspace.Provider, and
// surfacing the configuration over HTTP for the Settings page.
package wsadmin

import "time"

// Credentials is the persisted Workspace configuration.
type Credentials struct {
	SAJSON         []byte    `json:"-"`
	DelegatedAdmin string    `json:"delegated_admin"`
	CustomerID     string    `json:"customer_id"`
	SAEmail        string    `json:"sa_email,omitempty"`
	ProjectID      string    `json:"project_id,omitempty"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// StatusResponse is what GET /api/v1/admin/workspace/status returns.
type StatusResponse struct {
	Configured     bool      `json:"configured"`
	Source         string    `json:"source,omitempty"` // "db" | "env" | ""
	DelegatedAdmin string    `json:"delegated_admin,omitempty"`
	CustomerID     string    `json:"customer_id,omitempty"`
	SAEmail        string    `json:"sa_email,omitempty"`
	ProjectID      string    `json:"project_id,omitempty"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
	RequiredScopes []string  `json:"required_scopes"`
}

// UploadRequest captures what the multipart form-data upload supplies.
type UploadRequest struct {
	SAJSON         []byte
	DelegatedAdmin string
	CustomerID     string
}
