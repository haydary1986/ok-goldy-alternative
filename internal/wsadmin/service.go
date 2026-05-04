package wsadmin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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
		out.SAClientID = creds.SAClientID
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

	saMeta, err := parseSAJSON(req.SAJSON)
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
		SAEmail:        saMeta.ClientEmail,
		SAClientID:     saMeta.ClientID,
		ProjectID:      saMeta.ProjectID,
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

// Diagnostic probes each required scope independently against Google's token
// endpoint, so we can tell exactly which scopes are missing from DWD without
// asking the operator to read raw OAuth errors.
func (s *Service) Diagnostic(ctx context.Context) (*DiagnosticResponse, error) {
	creds, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	if creds == nil {
		return nil, ErrNotConfigured
	}

	probes := make([]ScopeProbe, 0, len(workspace.DefaultScopes))
	for _, scope := range workspace.DefaultScopes {
		probe := ScopeProbe{Scope: scope}
		c, buildErr := workspace.New(ctx, workspace.Config{
			ServiceAccountKey: creds.SAJSON,
			DelegatedAdmin:    creds.DelegatedAdmin,
			CustomerID:        creds.CustomerID,
			Scopes:            []string{scope},
			RateLimitRPS:      s.rateLimitRPS,
			RateLimitBurst:    s.rateLimitBurst,
		})
		if buildErr != nil {
			probe.Error = buildErr.Error()
			probes = append(probes, probe)
			continue
		}
		// Each scope drives a different read; user.alias has no list endpoint
		// so we re-use ListUsers (admin.directory.user.alias is granted by
		// users-scope already in practice).
		var probeErr error
		switch scope {
		case "https://www.googleapis.com/auth/admin.directory.group":
			_, _, probeErr = c.ListGroupsPage(ctx, "", 1)
		default:
			_, _, probeErr = c.ListUsersPage(ctx, "", 1)
		}
		if probeErr != nil {
			probe.Error = sanitiseOAuthError(probeErr.Error())
		} else {
			probe.OK = true
		}
		probes = append(probes, probe)
	}

	summary := buildDiagnosticSummary(creds.SAClientID, probes)
	return &DiagnosticResponse{
		SAClientID:     creds.SAClientID,
		DelegatedAdmin: creds.DelegatedAdmin,
		Probes:         probes,
		Summary:        summary,
	}, nil
}

func buildDiagnosticSummary(clientID string, probes []ScopeProbe) string {
	okCount := 0
	for _, p := range probes {
		if p.OK {
			okCount++
		}
	}
	if okCount == len(probes) {
		return "All required scopes authorized — credentials are working."
	}
	if okCount == 0 {
		return fmt.Sprintf(
			"None of the required scopes authorized. Most likely Domain-Wide Delegation has no entry for Client ID %s, or the entry exists with a different Client ID. Open Workspace Admin → Domain-Wide Delegation, add %s with all four scopes, click Authorize, wait 1–2 min, and re-run the test.",
			clientID, clientID,
		)
	}
	missing := []string{}
	for _, p := range probes {
		if !p.OK {
			missing = append(missing, shortScope(p.Scope))
		}
	}
	return fmt.Sprintf(
		"Some scopes are missing from DWD: %s. Open Workspace Admin → Domain-Wide Delegation → entry for Client ID %s → Edit → re-paste all four scopes (one line, comma-separated) → Authorize.",
		strings.Join(missing, ", "), clientID,
	)
}

func shortScope(s string) string {
	const prefix = "https://www.googleapis.com/auth/"
	if strings.HasPrefix(s, prefix) {
		return s[len(prefix):]
	}
	return s
}

// sanitiseOAuthError trims oauth2 errors to the structured part Google
// returns, hiding the noisy URL prefix.
func sanitiseOAuthError(s string) string {
	if i := strings.Index(s, "oauth2: cannot fetch token"); i >= 0 {
		return s[i:]
	}
	return s
}

// Test attempts a no-op Workspace call to confirm the credentials work.
// On the most common DWD misconfiguration ("unauthorized_client"), the
// returned error is wrapped with a friendly hint that includes the SA's
// Client ID so the operator knows exactly what to paste into Workspace
// Admin → Domain-Wide Delegation.
func (s *Service) Test(ctx context.Context) error {
	if s.provider == nil {
		return ErrNotConfigured
	}
	c := s.provider.Get()
	if c == nil {
		return ErrNotConfigured
	}
	if _, _, err := c.ListUsersPage(ctx, "", 1); err != nil {
		return s.wrapAuthError(ctx, err)
	}
	return nil
}

// wrapAuthError detects "unauthorized_client" errors (the canonical signal
// that Domain-Wide Delegation has not been authorized for the SA's Client
// ID with the right scopes) and rewrites them into something an operator
// can act on.
func (s *Service) wrapAuthError(ctx context.Context, err error) error {
	msg := err.Error()
	if !strings.Contains(msg, "unauthorized_client") {
		return fmt.Errorf("wsadmin: test connection: %w", err)
	}
	creds, _ := s.repo.Get(ctx)
	clientID := ""
	if creds != nil {
		clientID = creds.SAClientID
	}
	if clientID == "" {
		return fmt.Errorf("DWD not authorized: open Workspace Admin → Security → API Controls → Manage Domain-Wide Delegation, add this SA's Client ID with the four required scopes, and retry")
	}
	return fmt.Errorf(
		"DWD not authorized for Client ID %s. Open Workspace Admin → Domain-Wide Delegation, add (or edit) the entry for that exact Client ID and authorize the four required scopes, then retry. Original: %v",
		clientID, err,
	)
}

// saMetadata captures the public, display-safe fields extracted from a
// service-account JSON key. It deliberately excludes private_key.
type saMetadata struct {
	ClientEmail string
	ClientID    string
	ProjectID   string
}

// parseSAJSON pulls the public identifiers out of the SA JSON. The
// client_id is the 21-digit Unique ID that has to be added to Workspace
// Admin → Domain-Wide Delegation; without it, the resulting OAuth token
// exchange returns "unauthorized_client".
func parseSAJSON(b []byte) (*saMetadata, error) {
	var aux struct {
		Type        string `json:"type"`
		ClientEmail string `json:"client_email"`
		ClientID    string `json:"client_id"`
		ProjectID   string `json:"project_id"`
		PrivateKey  string `json:"private_key"`
	}
	if err := json.Unmarshal(b, &aux); err != nil {
		return nil, fmt.Errorf("not valid JSON: %w", err)
	}
	if aux.Type != "service_account" {
		return nil, fmt.Errorf("expected type=service_account, got %q", aux.Type)
	}
	if aux.ClientEmail == "" || aux.PrivateKey == "" {
		return nil, fmt.Errorf("missing client_email or private_key in JSON")
	}
	return &saMetadata{
		ClientEmail: aux.ClientEmail,
		ClientID:    aux.ClientID,
		ProjectID:   aux.ProjectID,
	}, nil
}
