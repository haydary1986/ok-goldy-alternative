package wsadmin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/haydary1986/ok-goldy-alternative/internal/workspace"
)

// ErrNotConfigured is returned when the caller asks about credentials that
// don't exist yet.
var ErrNotConfigured = errors.New("wsadmin: workspace credentials not configured")

// invalidError signals a 4xx-style validation failure on the upload.
type invalidError string

func (e invalidError) Error() string { return string(e) }

// errInvalid wraps a string into an invalidError.
func errInvalid(msg string) error { return invalidError(msg) }

// IsInvalid reports whether err originated from upload validation.
func IsInvalid(err error) bool {
	if err == nil {
		return false
	}
	var ie invalidError
	return errors.As(err, &ie)
}

// Service manages the lifecycle of the workspace credentials.
type Service struct {
	repo           *Repository
	provider       *workspace.Provider
	rateLimitRPS   int
	rateLimitBurst int
}

func NewService(repo *Repository, provider *workspace.Provider, rps, burst int) *Service {
	return &Service{repo: repo, provider: provider, rateLimitRPS: rps, rateLimitBurst: burst}
}

// Status returns the current configuration shape (without secrets) for the UI.
func (s *Service) Status(ctx context.Context) (*StatusResponse, error) {
	out := &StatusResponse{RequiredScopes: workspace.DefaultScopes}
	creds, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	if creds != nil {
		out.Configured = true
		out.Source = "db"
		out.DelegatedAdmin = creds.DelegatedAdmin
		out.CustomerID = creds.CustomerID
		out.SAEmail = creds.SAEmail
		out.ProjectID = creds.ProjectID
		out.UpdatedAt = creds.UpdatedAt
		return out, nil
	}
	if s.provider != nil && s.provider.Get() != nil {
		out.Configured = true
		out.Source = "env"
	}
	return out, nil
}

// Save validates and persists the uploaded credentials, then hot-swaps the
// workspace.Provider so subsequent requests use the new client immediately.
func (s *Service) Save(ctx context.Context, req UploadRequest) (*Credentials, error) {
	if req.DelegatedAdmin == "" {
		return nil, errInvalid("delegated_admin is required")
	}
	if len(req.SAJSON) == 0 {
		return nil, errInvalid("service account JSON file is required")
	}

	saEmail, projectID, err := parseSAJSON(req.SAJSON)
	if err != nil {
		return nil, errInvalid(err.Error())
	}

	customerID := req.CustomerID
	if customerID == "" {
		customerID = "my_customer"
	}

	// Build the client first so we surface JSON parsing / scope issues before
	// we touch the database.
	client, err := workspace.New(ctx, workspace.Config{
		ServiceAccountKey: req.SAJSON,
		DelegatedAdmin:    req.DelegatedAdmin,
		CustomerID:        customerID,
		RateLimitRPS:      s.rateLimitRPS,
		RateLimitBurst:    s.rateLimitBurst,
	})
	if err != nil {
		return nil, errInvalid(fmt.Sprintf("invalid service account JSON: %v", err))
	}

	creds := &Credentials{
		SAJSON:         req.SAJSON,
		DelegatedAdmin: req.DelegatedAdmin,
		CustomerID:     customerID,
		SAEmail:        saEmail,
		ProjectID:      projectID,
	}
	if err := s.repo.Upsert(ctx, creds); err != nil {
		return nil, err
	}

	if s.provider != nil {
		s.provider.Set(client)
	}
	// Re-read so UpdatedAt is populated from the DB.
	saved, err := s.repo.Get(ctx)
	if err != nil || saved == nil {
		return creds, nil
	}
	return saved, nil
}

// Delete removes the stored credentials and clears the provider.
func (s *Service) Delete(ctx context.Context) error {
	if err := s.repo.Delete(ctx); err != nil {
		return err
	}
	if s.provider != nil {
		s.provider.Set(nil)
	}
	return nil
}

// Test attempts a no-op Workspace call to confirm the credentials work.
func (s *Service) Test(ctx context.Context) error {
	if s.provider == nil {
		return ErrNotConfigured
	}
	c := s.provider.Get()
	if c == nil {
		return ErrNotConfigured
	}
	if _, _, err := c.ListUsersPage(ctx, "", 1); err != nil {
		return fmt.Errorf("wsadmin: test connection: %w", err)
	}
	return nil
}

// parseSAJSON pulls the client_email and project_id out of the SA JSON for
// display purposes and to confirm the file is well-formed.
func parseSAJSON(b []byte) (email, projectID string, err error) {
	var aux struct {
		Type        string `json:"type"`
		ClientEmail string `json:"client_email"`
		ProjectID   string `json:"project_id"`
		PrivateKey  string `json:"private_key"`
	}
	if err := json.Unmarshal(b, &aux); err != nil {
		return "", "", fmt.Errorf("not valid JSON: %w", err)
	}
	if aux.Type != "service_account" {
		return "", "", fmt.Errorf("expected type=service_account, got %q", aux.Type)
	}
	if aux.ClientEmail == "" || aux.PrivateKey == "" {
		return "", "", fmt.Errorf("missing client_email or private_key in JSON")
	}
	return aux.ClientEmail, aux.ProjectID, nil
}
