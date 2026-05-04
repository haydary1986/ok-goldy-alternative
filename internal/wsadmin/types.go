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
	SAClientID     string    `json:"sa_client_id,omitempty"`
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
	SAClientID     string    `json:"sa_client_id,omitempty"`
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

// ScopeProbe is the result of testing one OAuth scope independently against
// Google's token endpoint, so an operator can see which scopes are missing
// from Domain-Wide Delegation.
type ScopeProbe struct {
	Scope string `json:"scope"`
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// DiagnosticResponse is the body returned by POST /api/v1/admin/workspace/diagnostic.
type DiagnosticResponse struct {
	SAClientID     string       `json:"sa_client_id,omitempty"`
	DelegatedAdmin string       `json:"delegated_admin,omitempty"`
	Probes         []ScopeProbe `json:"probes"`
	Summary        string       `json:"summary"`
}
